package httpapi

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
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

func ToHumaErr(ctx context.Context, err error) error {
	var safeErr *apperrors.SafeError
	userMsg := "internal server error"

	if errors.As(err, &safeErr) {
		userMsg = safeErr.UserMsg
		wideevent.Add(ctx, "error", safeErr.LogString())
	} else {
		wideevent.Add(ctx, "error", err.Error())
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			userMsg = "resource not found"
		case errors.Is(err, apperrors.ErrConflict):
			userMsg = "resource conflict"
		case errors.Is(err, apperrors.ErrUnauthorized):
			userMsg = "unauthorized access"
		case errors.Is(err, apperrors.ErrForbidden):
			userMsg = "forbidden access"
		case errors.Is(err, apperrors.ErrInvalidInput):
			userMsg = "invalid input data"
		}
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return huma.Error404NotFound(userMsg)
	case errors.Is(err, apperrors.ErrConflict):
		return huma.Error409Conflict(userMsg)
	case errors.Is(err, apperrors.ErrUnauthorized):
		return huma.Error401Unauthorized(userMsg)
	case errors.Is(err, apperrors.ErrForbidden):
		return huma.Error403Forbidden(userMsg)
	case errors.Is(err, apperrors.ErrInvalidInput):
		return huma.Error400BadRequest(userMsg)
	default:
		return huma.Error500InternalServerError(userMsg)
	}
}
