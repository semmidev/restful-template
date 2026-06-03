package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

// init overrides huma.NewError so that every error path in Huma —
// auth middleware huma.WriteErr calls, built-in validation failures,
// 404/405/415 responses — all produce the same APIError shape.
//
// This is the canonical hook documented at https://huma.rocks/features/response-errors/
func init() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		code := statusToCode(status)
		apiErr := newAPIError(status, code, msg)
		for _, e := range errs {
			if e == nil {
				continue
			}
			var det *huma.ErrorDetail
			if ed, ok := e.(huma.ErrorDetailer); ok {
				det = ed.ErrorDetail()
			} else {
				det = &huma.ErrorDetail{Message: e.Error()}
			}
			apiErr.Errors = append(apiErr.Errors, det)
		}
		return apiErr
	}
}

type CtxKey string

const (
	UserIDKey    CtxKey = "user_id"
	UserEmailKey CtxKey = "user_email"
)

// APIError is the canonical application error response.
// It fully implements RFC 9457 Problem Details and is extended with:
//   - `code`: machine-readable string for i18n / programmatic handling
//   - `errors`: optional list of sub-errors (Huma validation details)
//
// Every Huma error path — handler returns, huma.WriteErr, built-in
// validation, 404/405/415 — produces this exact shape via the huma.NewError
// override in init().
//
// Example response body:
//
//	{
//	  "type":    "about:blank",
//	  "title":   "Not Found",
//	  "status":  404,
//	  "detail":  "The requested todo does not exist",
//	  "code":    "NOT_FOUND"
//	}
type APIError struct {
	Type     string              `json:"type"`
	Title    string              `json:"title"`
	Status   int                 `json:"status"`
	Detail   string              `json:"detail,omitempty"`
	Instance string              `json:"instance,omitempty"`
	Code     string              `json:"code,omitempty"`
	Errors   []*huma.ErrorDetail `json:"errors,omitempty"`
}

// GetStatus implements huma.StatusError — Huma uses our struct as the body.
func (e *APIError) GetStatus() int { return e.Status }

// Error implements the error interface.
func (e *APIError) Error() string { return e.Detail }

// ContentType tells Huma to return application/problem+json (RFC 9457)
// instead of application/json for all error responses.
func (e *APIError) ContentType(ct string) string {
	if ct == "application/json" {
		return "application/problem+json"
	}
	if ct == "application/cbor" {
		return "application/problem+cbor"
	}
	return ct
}

// statusToCode derives a default machine-readable code from an HTTP status.
// This is used for errors that originate inside Huma itself (routing, validation)
// rather than from application SafeErrors (which carry their own Code).
func statusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		return http.StatusText(status)
	}
}

func newAPIError(status int, code, detail string) *APIError {
	return &APIError{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

// ExtractUserID retrieves the authenticated user's UUID from the request context.
// Returns ErrUnauthorized if the key is missing (i.e. the auth middleware did not run).
func ExtractUserID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return uuid.Nil, apperrors.ErrUnauthorized
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, apperrors.NewInternal("invalid user_id type in context", nil)
	}
	return id, nil
}

// ToHumaErr maps an application error to a Huma-compatible *APIError.
//
// Priority order:
//  1. If the error wraps a *SafeError, use its Code and UserMsg directly —
//     the message is already safe for clients.
//  2. Otherwise fall back to the sentinel (ErrNotFound etc.) for HTTP status,
//     and derive a generic code + message.
//
// The wide event is always enriched with error context for canonical log lines.
func ToHumaErr(ctx context.Context, err error) error {
	var safeErr *apperrors.SafeError
	if errors.As(err, &safeErr) {
		// SafeError carries both the safe user message and a structured log string.
		wideevent.Add(ctx, "error", safeErr.LogString())
		return newAPIError(sentinelStatus(err), safeErr.Code, safeErr.UserMsg)
	}

	// Plain sentinel — log the raw error, derive safe defaults.
	wideevent.Add(ctx, "error", err.Error())
	return newAPIError(sentinelStatus(err), statusToCode(sentinelStatus(err)), sentinelMsg(err))
}

// sentinelStatus maps the error chain to an HTTP status code.
func sentinelStatus(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, apperrors.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// sentinelMsg returns a generic safe message for bare sentinel errors.
func sentinelMsg(err error) string {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return "resource not found"
	case errors.Is(err, apperrors.ErrConflict):
		return "resource conflict"
	case errors.Is(err, apperrors.ErrUnauthorized):
		return "unauthorized access"
	case errors.Is(err, apperrors.ErrForbidden):
		return "forbidden"
	case errors.Is(err, apperrors.ErrInvalidInput):
		return "invalid input data"
	default:
		return "internal server error"
	}
}

// ToHumaErrUnauthorized is a convenience for returning a plain 401 without a SafeError.
// Used in middleware where we don't have a full SafeError context.
func ToHumaErrUnauthorized(msg string) *APIError {
	return newAPIError(http.StatusUnauthorized, "UNAUTHORIZED", msg)
}

// WriteHumaErr writes an *APIError directly to a huma.Context.
// The content type is set to application/problem+json per RFC 9457.
func WriteHumaErr(api huma.API, ctx huma.Context, apiErr *APIError) {
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(apiErr.Status)
	b, _ := json.Marshal(apiErr)
	_, _ = ctx.BodyWriter().Write(b)
}
