package delivery

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/domain"
)

// ─── Route Registration ─────────────────────────────────────────────────────

func RegisterRoutes(api huma.API, auth domain.AuthUsecase, todos domain.TodoUsecase, log *slog.Logger) {
	RegisterHealthRoutes(api)
	RegisterAuthRoutes(api, auth)
	RegisterTodoRoutes(api, todos)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func extractUserID(ctx context.Context) (uuid.UUID, error) {
	raw := GetUserID(ctx)
	if raw == "" {
		return uuid.Nil, domain.ErrUnauthorized
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return id, nil
}

// toHumaErr maps domain errors to RFC 9457 problem+json Huma errors.
func toHumaErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return huma.Error404NotFound("resource not found")
	case errors.Is(err, domain.ErrConflict):
		return huma.Error409Conflict("resource already exists")
	case errors.Is(err, domain.ErrUnauthorized):
		return huma.Error401Unauthorized("unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, domain.ErrInvalidInput):
		return huma.Error400BadRequest("invalid input")
	default:
		return huma.Error500InternalServerError("internal server error")
	}
}
