package delivery

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type HealthResp struct {
	Body struct {
		Data map[string]string `json:"data"`
	}
}

func RegisterHealthRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, _ *struct{}) (*HealthResp, error) {
		resp := &HealthResp{}
		resp.Body.Data = map[string]string{"status": "ok"}
		return resp, nil
	})
}
