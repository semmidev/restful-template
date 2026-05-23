package httpapi

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type CtxKey string

const (
	UserIDKey    CtxKey = "user_id"
	UserEmailKey CtxKey = "user_email"
)

func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return uuid.Nil, apperrors.ErrUnauthorized
	}
	id, err := uuid.Parse(val.(string))
	if err != nil {
		return uuid.Nil, apperrors.ErrUnauthorized
	}
	return id, nil
}

func ToHumaErr(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return huma.Error404NotFound("resource not found")
	case errors.Is(err, apperrors.ErrConflict):
		return huma.Error409Conflict("resource already exists")
	case errors.Is(err, apperrors.ErrUnauthorized):
		return huma.Error401Unauthorized("unauthorized")
	case errors.Is(err, apperrors.ErrForbidden):
		return huma.Error403Forbidden("forbidden")
	case errors.Is(err, apperrors.ErrInvalidInput):
		return huma.Error400BadRequest("invalid input")
	default:
		return huma.Error500InternalServerError("internal server error")
	}
}
