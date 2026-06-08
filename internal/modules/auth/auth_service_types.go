package auth

import (
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/password"
	"github.com/semmidev/restful-template/internal/shared/uuidgen"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	UserID       uuid.UUID `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	ActiveRole   string    `json:"active_role"`
	Roles        []string  `json:"roles"`
	Permissions  []string  `json:"permissions"`
}

type RegisterInput struct {
	Email    string `json:"email" format:"email" required:"true"`
	Password string `json:"password" minLength:"8" required:"true"`
}

func (in *RegisterInput) Validate() error {
	if in.Email == "" || in.Password == "" {
		return errors.NewInvalidInput("Email and password are required", errors.ErrInvalidInput)
	}
	return nil
}

func (in *RegisterInput) ToUser() (*User, error) {
	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, errors.NewInternal("Failed to process registration", err)
	}
	return &User{
		ID:           uuidgen.New(),
		Email:        in.Email,
		PasswordHash: &hash,
	}, nil
}

type LoginInput struct {
	Email    string `json:"email" format:"email" required:"true"`
	Password string `json:"password" required:"true"`
}
