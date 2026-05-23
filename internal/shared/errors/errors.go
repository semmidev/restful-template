package errors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
)

// SafeError keeps internal details hidden from clients.
type SafeError struct {
	Code     string            // machine-readable (e.g. "USER_NOT_FOUND")
	UserMsg  string            // safe public message
	Internal error             // full original error (for logs only)
	Metadata map[string]string // sanitized context for logging
}

func (e *SafeError) Error() string {
	return e.UserMsg // CRITICAL: never leaks Internal
}

func (e *SafeError) Unwrap() error {
	return e.Internal // allows errors.Is/As to work
}

func (e *SafeError) LogString() string {
	return "Code: " + e.Code + " | UserMsg: " + e.UserMsg + " | Cause: " + (func() string {
		if e.Internal != nil {
			return e.Internal.Error()
		}
		return "nil"
	})()
}

// New creates a safe error (use this everywhere).
func New(code, userMsg string, internal error, meta map[string]string) *SafeError {
	if meta == nil {
		meta = make(map[string]string)
	}
	return &SafeError{
		Code:     code,
		UserMsg:  userMsg,
		Internal: internal,
		Metadata: meta,
	}
}

// Predefined constructors for common HTTP-mapped errors
func NewInvalidInput(userMsg string, internal error) *SafeError {
	return New("INVALID_INPUT", userMsg, internal, nil)
}

func NewConflict(userMsg string, internal error) *SafeError {
	return New("CONFLICT", userMsg, internal, nil)
}

func NewInternal(userMsg string, internal error) *SafeError {
	return New("INTERNAL_ERROR", userMsg, internal, nil)
}

func NewUnauthorized(userMsg string, internal error) *SafeError {
	return New("UNAUTHORIZED", userMsg, internal, nil)
}

func NewNotFound(userMsg string, internal error) *SafeError {
	return New("NOT_FOUND", userMsg, internal, nil)
}
