// Package uuidgen wraps google/uuid to provide UUID v7 generation.
// UUID v7 is time-ordered (monotonically sortable), making it ideal as
// a database primary key — it avoids index fragmentation unlike UUID v4.
package uuidgen

import (
	"fmt"

	"github.com/google/uuid"
)

// New generates a new UUID v7 (time-ordered, sortable).
// It panics only on rand failure, which is unrecoverable in practice.
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		// uuid.NewV7 only fails if the OS random source is exhausted — fatal.
		panic(fmt.Sprintf("uuidgen: uuid.NewV7 failed: %v", err))
	}
	return id
}

// NewSafe generates a UUID v7 and returns any error.
func NewSafe() (uuid.UUID, error) {
	return uuid.NewV7()
}

// MustParse parses a UUID string, panicking on error (useful in tests/init).
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}
