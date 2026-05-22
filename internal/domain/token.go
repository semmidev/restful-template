package domain

import (
	"context"

	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID
	Email  string
}

type TokenService interface {
	GeneratePair(ctx context.Context, userID uuid.UUID, email string) (access, refresh string, accessExp, refreshExp int64, err error)
	ParseAccess(ctx context.Context, token string) (*TokenClaims, error)
	ParseRefresh(ctx context.Context, token string) (*TokenClaims, error)
}
