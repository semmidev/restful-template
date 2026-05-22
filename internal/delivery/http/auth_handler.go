package delivery

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/domain"
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

func RegisterAuthRoutes(api huma.API, auth domain.AuthUsecase) {
	huma.Register(api, huma.Operation{
		OperationID: "auth-register",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, in *struct{ Body RegisterBody }) (*TokenResp, error) {
		pair, err := auth.Register(ctx, domain.RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			wideevent.Add(ctx, "error", err.Error())
			return nil, toHumaErr(err)
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
		pair, err := auth.Login(ctx, domain.LoginInput{Email: in.Body.Email, Password: in.Body.Password})
		if err != nil {
			wideevent.Add(ctx, "error", err.Error())
			return nil, toHumaErr(err)
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
			return nil, toHumaErr(err)
		}
		return tokenResp(pair), nil
	})
}

func tokenResp(pair domain.TokenPair) *TokenResp {
	r := &TokenResp{}
	r.Body.Data.AccessToken = pair.AccessToken
	r.Body.Data.RefreshToken = pair.RefreshToken
	r.Body.Data.ExpiresIn = pair.ExpiresIn
	return r
}
