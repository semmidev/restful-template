package errors

import (
	"errors"
	"strings"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
)

// SafeError keeps internal details hidden from clients while preserving
// structured context for observability.
//
// The Code field drives machine-readable error identification (used by ToHumaErr
// to set the `code` field in RFC 9457 responses).
// The Internal field allows errors.Is/As chains to work through the stack.
type SafeError struct {
	Code     string // machine-readable (e.g. "TODO_NOT_FOUND")
	UserMsg  string // safe public message — NEVER expose Internal to clients
	Internal error  // full original error (for logs only, never sent to clients)
}

// Error implements the error interface. Returns only the safe user message.
func (e *SafeError) Error() string { return e.UserMsg }

// Unwrap allows errors.Is / errors.As to traverse the error chain.
func (e *SafeError) Unwrap() error { return e.Internal }

// LogString returns a structured log-safe representation including the internal
// cause. Use this when writing to slog / wideevent, never send it to clients.
func (e *SafeError) LogString() string {
	var sb strings.Builder
	sb.WriteString("code=")
	sb.WriteString(e.Code)
	sb.WriteString(" msg=")
	sb.WriteString(e.UserMsg)
	if e.Internal != nil {
		sb.WriteString(" cause=")
		sb.WriteString(e.Internal.Error())
	}
	return sb.String()
}

// newSafeError is the internal constructor. All public constructors delegate here.
func newSafeError(code, userMsg string, internal error) *SafeError {
	return &SafeError{
		Code:     code,
		UserMsg:  userMsg,
		Internal: internal,
	}
}

// Typed constructors — prefer these over raw sentinel errors in business code.

func NewNotFound(userMsg string, internal error) *SafeError {
	return newSafeError("NOT_FOUND", userMsg, internal)
}

func NewConflict(userMsg string, internal error) *SafeError {
	return newSafeError("CONFLICT", userMsg, internal)
}

func NewUnauthorized(userMsg string, internal error) *SafeError {
	return newSafeError("UNAUTHORIZED", userMsg, internal)
}

func NewForbidden(userMsg string, internal error) *SafeError {
	return newSafeError("FORBIDDEN", userMsg, internal)
}

func NewInvalidInput(userMsg string, internal error) *SafeError {
	return newSafeError("INVALID_INPUT", userMsg, internal)
}

func NewInternal(userMsg string, internal error) *SafeError {
	return newSafeError("INTERNAL_ERROR", userMsg, internal)
}
