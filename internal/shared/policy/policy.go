package policy

import (
	"context"
	_ "embed"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/rego"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
)

//go:embed policy.rego
var regoPolicy string

//go:embed permissions.json
var defaultPermissionsJSON string

type Evaluator struct {
	query      rego.PreparedEvalQuery
	permsQuery rego.PreparedEvalQuery
}

var evaluator atomic.Pointer[Evaluator]

// buildPermissionsModule generates a dynamic Rego module loading the permissions JSON
func buildPermissionsModule(jsonStr string) string {
	return fmt.Sprintf(`package permissions
import rego.v1
role_permissions := json.unmarshal(%q)
`, jsonStr)
}

// Init initializes the OPA evaluator with the embedded default permissions.
func Init(ctx context.Context) error {
	return Reload(ctx, defaultPermissionsJSON)
}

// Reload allows dynamic update/swap of the permissions schema at runtime.
func Reload(ctx context.Context, permissionsJSON string) error {
	permsModule := buildPermissionsModule(permissionsJSON)

	r := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module("policy.rego", regoPolicy),
		rego.Module("permissions.rego", permsModule),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego query: %w", err)
	}

	rPerms := rego.New(
		rego.Query("data.authz.role_permissions"),
		rego.Module("policy.rego", regoPolicy),
		rego.Module("permissions.rego", permsModule),
	)
	permsQuery, err := rPerms.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego permissions query: %w", err)
	}

	newEvaluator := &Evaluator{
		query:      query,
		permsQuery: permsQuery,
	}

	evaluator.Store(newEvaluator)
	return nil
}

// GetRolePermissions queries OPA to return all permissions assigned to a given role.
func GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	ev := evaluator.Load()
	if ev == nil {
		return nil, apperrors.NewInternal("policy evaluator is not initialized", nil)
	}

	results, err := ev.permsQuery.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate permissions: %w", err)
	}

	if len(results) == 0 || len(results[0].Expressions) == 0 {
		return nil, nil
	}

	rawMap, ok := results[0].Expressions[0].Value.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	rawList, ok := rawMap[role]
	if !ok {
		return nil, nil
	}

	list, ok := rawList.([]interface{})
	if !ok {
		return nil, nil
	}

	perms := make([]string, len(list))
	for i, v := range list {
		if s, ok := v.(string); ok {
			perms[i] = s
		}
	}

	return perms, nil
}

type Input struct {
	UserID          string                 `json:"user_id"`
	ActiveRole      string                 `json:"active_role"`
	Roles           []string               `json:"roles"`
	Action          string                 `json:"action"`
	ResourceOwnerID string                 `json:"resource_owner_id,omitempty"` // legacy
	Resource        map[string]interface{} `json:"resource,omitempty"`          // new scalable/ABAC context
}

func (e *Evaluator) IsAllowed(ctx context.Context, input Input) (bool, error) {
	results, err := e.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, fmt.Errorf("rego evaluation failed: %w", err)
	}

	if len(results) == 0 {
		return false, nil
	}

	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, nil
	}

	return allowed, nil
}

// Authorize extracts credentials from context and checks permissions against the OPA engine.
func Authorize(ctx context.Context, action string, resourceOwnerID string) error {
	var resource map[string]interface{}
	if resourceOwnerID != "" {
		resource = map[string]interface{}{
			"owner_id": resourceOwnerID,
		}
	}
	return AuthorizeWithResource(ctx, action, resource)
}

// AuthorizeWithResource checks permissions against the OPA engine with a custom resource context.
func AuthorizeWithResource(ctx context.Context, action string, resource map[string]interface{}) error {
	ev := evaluator.Load()
	if ev == nil {
		return apperrors.NewInternal("policy evaluator is not initialized", nil)
	}

	userIDVal := ctx.Value(httpapi.UserIDKey)
	if userIDVal == nil {
		return apperrors.ErrUnauthorized
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return apperrors.NewInternal("invalid user_id type in context", nil)
	}

	activeRoleVal := ctx.Value(httpapi.UserActiveRoleKey)
	var activeRole string
	if activeRoleVal != nil {
		activeRole, _ = activeRoleVal.(string)
	}
	if activeRole == "" {
		activeRole = "user"
	}

	rolesVal := ctx.Value(httpapi.UserRolesKey)
	var roles []string
	if rolesVal != nil {
		roles, _ = rolesVal.([]string)
	}
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	allowed, err := ev.IsAllowed(ctx, Input{
		UserID:     userID.String(),
		ActiveRole: activeRole,
		Roles:      roles,
		Action:     action,
		Resource:   resource,
	})
	if err != nil {
		return apperrors.NewInternal("authorization policy evaluation failed", err)
	}

	if !allowed {
		if resource != nil && resource["owner_id"] != nil {
			return apperrors.ErrNotFound
		}
		return apperrors.ErrForbidden
	}

	return nil
}
