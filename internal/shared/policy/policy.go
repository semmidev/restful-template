package policy

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/v1/rego"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
)

//go:embed policy.rego
var regoPolicy string

type Evaluator struct {
	query      rego.PreparedEvalQuery
	permsQuery rego.PreparedEvalQuery
}

var evaluator *Evaluator

func Init(ctx context.Context) error {
	r := rego.New(
		rego.Query("data.authz.allow"), // Evaluasi rule allow yang ada di package authz, lalu berikan hasilnya kepada saya
		rego.Module("policy.rego", regoPolicy),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego query: %w", err)
	}

	rPerms := rego.New(
		rego.Query("data.authz.role_permissions"),
		rego.Module("policy.rego", regoPolicy),
	)
	permsQuery, err := rPerms.PrepareForEval(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare rego permissions query: %w", err)
	}

	evaluator = &Evaluator{
		query:      query,
		permsQuery: permsQuery,
	}
	return nil
}

// GetRolePermissions queries OPA to return all permissions assigned to a given role.
func GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	if evaluator == nil {
		return nil, apperrors.NewInternal("policy evaluator is not initialized", nil)
	}

	results, err := evaluator.permsQuery.Eval(ctx)
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
	UserID          string   `json:"user_id"`
	ActiveRole      string   `json:"active_role"`
	Roles           []string `json:"roles"`
	Action          string   `json:"action"`
	ResourceOwnerID string   `json:"resource_owner_id,omitempty"`
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

// Authorize extracts credentials from the context and checks permissions against the OPA engine.
func Authorize(ctx context.Context, action string, resourceOwnerID string) error {
	if evaluator == nil {
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

	allowed, err := evaluator.IsAllowed(ctx, Input{
		UserID:          userID.String(),
		ActiveRole:      activeRole,
		Roles:           roles,
		Action:          action,
		ResourceOwnerID: resourceOwnerID,
	})
	if err != nil {
		return apperrors.NewInternal("authorization policy evaluation failed", err)
	}

	if !allowed {
		if resourceOwnerID != "" {
			return apperrors.ErrNotFound
		}
		return apperrors.ErrForbidden
	}

	return nil
}
