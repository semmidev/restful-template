package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

// APIError is the canonical application error response.
// It extends the RFC 9457 Problem Details format with a machine-readable
// `code` field that clients can use for i18n and programmatic error handling.
//
// Example response body:
//
//	{
//	  "type":    "about:blank",
//	  "title":   "Not Found",
//	  "status":  404,
//	  "detail":  "The requested todo does not exist",
//	  "instance": "/api/v1/todos/123",
//	  "code":    "TODO_NOT_FOUND"
//	}
type APIError struct {
	Type     string `json:"type" default:"about:blank"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

// GetStatus implements huma.StatusError so Huma uses our custom struct as the
// error response body instead of its own ErrorModel.
func (e *APIError) GetStatus() int { return e.Status }

// Error implements the error interface.
func (e *APIError) Error() string { return e.Detail }

func newAPIError(status int, code, detail string) *APIError {
	return &APIError{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return uuid.Nil, apperrors.ErrUnauthorized
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid type for user_id in context")
	}
	return id, nil
}

// ToHumaErr maps an application error to a Huma-compatible *APIError.
// It always carries:
//   - The safe user-facing message (never leaks internal details)
//   - A machine-readable code (from SafeError.Code or derived from the sentinel)
//   - Enriches the wide event with error context for canonical log lines
func ToHumaErr(ctx context.Context, err error) error {
	var safeErr *apperrors.SafeError
	code := "INTERNAL_ERROR"
	userMsg := "internal server error"

	if errors.As(err, &safeErr) {
		code = safeErr.Code
		userMsg = safeErr.UserMsg
		wideevent.Add(ctx, "error", safeErr.LogString())
	} else {
		wideevent.Add(ctx, "error", err.Error())
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			code = "NOT_FOUND"
			userMsg = "resource not found"
		case errors.Is(err, apperrors.ErrConflict):
			code = "CONFLICT"
			userMsg = "resource conflict"
		case errors.Is(err, apperrors.ErrUnauthorized):
			code = "UNAUTHORIZED"
			userMsg = "unauthorized access"
		case errors.Is(err, apperrors.ErrForbidden):
			code = "FORBIDDEN"
			userMsg = "forbidden access"
		case errors.Is(err, apperrors.ErrInvalidInput):
			code = "INVALID_INPUT"
			userMsg = "invalid input data"
		}
	}

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return newAPIError(http.StatusNotFound, code, userMsg)
	case errors.Is(err, apperrors.ErrConflict):
		return newAPIError(http.StatusConflict, code, userMsg)
	case errors.Is(err, apperrors.ErrUnauthorized):
		return newAPIError(http.StatusUnauthorized, code, userMsg)
	case errors.Is(err, apperrors.ErrForbidden):
		return newAPIError(http.StatusForbidden, code, userMsg)
	case errors.Is(err, apperrors.ErrInvalidInput):
		return newAPIError(http.StatusBadRequest, code, userMsg)
	default:
		return newAPIError(http.StatusInternalServerError, code, userMsg)
	}
}

// ToHumaErrUnauthorized is a convenience for returning a plain 401 without a SafeError.
// Used in middleware where we don't have a full SafeError context.
func ToHumaErrUnauthorized(msg string) *APIError {
	return newAPIError(http.StatusUnauthorized, "UNAUTHORIZED", msg)
}

// ToHumaErrBadRequest is a convenience for 400 without a SafeError.
func ToHumaErrBadRequest(msg string) *APIError {
	return newAPIError(http.StatusBadRequest, "BAD_REQUEST", msg)
}

// WriteHumaErr writes an *APIError directly to a huma.Context (for use in Huma middleware).
func WriteHumaErr(api huma.API, ctx huma.Context, apiErr *APIError) {
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(apiErr.Status)
	b, _ := json.Marshal(apiErr)
	_, _ = ctx.BodyWriter().Write(b)
}
