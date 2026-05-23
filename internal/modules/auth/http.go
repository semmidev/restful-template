package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

type RegisterBody struct {
	Email    string `json:"email" format:"email" minLength:"3" maxLength:"254" doc:"User email address"`
	Password string `json:"password" minLength:"8" maxLength:"72" doc:"User password (8-72 chars)"`
}

type LoginBody = RegisterBody

type RefreshBody struct {
	RefreshToken string `json:"refresh_token" minLength:"1"`
}

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in" doc:"Unix timestamp when the access token expires"`
}

type TokenResp struct {
	Body struct {
		Data TokenData `json:"data"`
	}
}

func RegisterAuthRoutes(api huma.API, auth *Usecase) {
	huma.Register(api, huma.Operation{
		OperationID: "auth-register",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body RegisterBody }) (*TokenResp, error) {
		pair, err := auth.Register(ctx, RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			wideevent.Add(ctx, "error", err.Error())
			return nil, httpapi.ToHumaErr(err)
		}
		wideevent.Add(ctx, "user_email", in.Body.Email)
		return tokenResp(pair), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Login and receive tokens",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body LoginBody }) (*TokenResp, error) {
		pair, err := auth.Login(ctx, LoginInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			wideevent.Add(ctx, "error", err.Error())
			return nil, httpapi.ToHumaErr(err)
		}
		wideevent.Add(ctx, "user_email", in.Body.Email)
		return tokenResp(pair), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "auth-refresh",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/refresh",
		Summary:     "Refresh access token using a refresh token",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body RefreshBody }) (*TokenResp, error) {
		pair, err := auth.Refresh(ctx, in.Body.RefreshToken)
		if err != nil {
			return nil, httpapi.ToHumaErr(err)
		}
		return tokenResp(pair), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "auth-delete-account",
		Method:        http.MethodDelete,
		Path:          "/api/v1/auth/account",
		Summary:       "Delete user account and all associated data",
		Tags:          []string{"Auth"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *struct{}) (*struct{}, error) {
		userIDStr := GetUserID(ctx)
		if userIDStr == "" {
			return nil, httpapi.ToHumaErr(errors.ErrUnauthorized)
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, httpapi.ToHumaErr(errors.ErrUnauthorized)
		}

		if err := auth.DeleteAccount(ctx, userID); err != nil {
			wideevent.Add(ctx, "error", err.Error())
			return nil, httpapi.ToHumaErr(err)
		}
		return &struct{}{}, nil
	})
}

func tokenResp(pair TokenPair) *TokenResp {
	r := &TokenResp{}
	r.Body.Data.AccessToken = pair.AccessToken
	r.Body.Data.RefreshToken = pair.RefreshToken
	r.Body.Data.ExpiresIn = pair.ExpiresIn
	return r
}
