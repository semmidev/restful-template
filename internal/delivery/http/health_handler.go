package delivery

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
)

// HealthChecker is a function that probes a single downstream dependency.
// Return nil if healthy, a non-nil error with a short description if not.
type HealthChecker func(ctx context.Context) error

type HealthResp struct {
	Body struct {
		Data map[string]string `json:"data"`
	}
}

// RegisterHealthRoutes registers the /api/v1/health endpoint.
//
// point 18: the handler now probes each registered HealthChecker dependency.
// If all are healthy → 200 {"status":"ok","postgres":"ok","redis":"ok"}.
// If any fail       → 503 with the degraded dependency listed.
// This ensures Kubernetes readiness probes correctly shed traffic when the
// database or cache is unavailable.
func RegisterHealthRoutes(api huma.API, checkers map[string]HealthChecker) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check — probes all downstream dependencies",
		Tags:        []string{"System"},
	}, func(ctx context.Context, _ *struct{}) (*HealthResp, error) {
		result := make(map[string]string, len(checkers)+1)
		allOK := true

		for name, check := range checkers {
			if err := check(ctx); err != nil {
				result[name] = "unhealthy: " + err.Error()
				allOK = false
			} else {
				result[name] = "ok"
			}
		}

		if !allOK {
			result["status"] = "degraded"
			return nil, &httpapi.APIError{
				Status: http.StatusServiceUnavailable,
				Title:  "Service Unavailable",
				Code:   "SERVICE_DEGRADED",
				Detail: "one or more dependencies are unhealthy",
			}
		}

		result["status"] = "ok"
		resp := &HealthResp{}
		resp.Body.Data = result
		return resp, nil
	})
}
