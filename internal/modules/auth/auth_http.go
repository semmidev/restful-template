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

type authRegisterReq struct{ Body RegisterBody }
type authLoginReq struct{ Body LoginBody }
type authRefreshReq struct{ Body RefreshBody }
type authDeleteAccountReq struct{}

type authHandler struct {
	auth AuthService
}

func RegisterAuthRoutes(api huma.API, auth AuthService) {
	h := &authHandler{auth: auth}

	huma.Register(api, huma.Operation{
		OperationID: "auth-register",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"Auth"},
	}, h.handleRegister)

	huma.Register(api, huma.Operation{
		OperationID: "auth-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Login and receive tokens",
		Tags:        []string{"Auth"},
	}, h.handleLogin)

	huma.Register(api, huma.Operation{
		OperationID: "auth-refresh",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/refresh",
		Summary:     "Refresh access token using a refresh token",
		Tags:        []string{"Auth"},
	}, h.handleRefresh)

	huma.Register(api, huma.Operation{
		OperationID:   "auth-delete-account",
		Method:        http.MethodDelete,
		Path:          "/api/v1/auth/account",
		Summary:       "Delete user account and all associated data",
		Tags:          []string{"Auth"},
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.handleDeleteAccount)
}

func (h *authHandler) handleRegister(ctx context.Context, in *authRegisterReq) (*TokenResp, error) {
	pair, err := h.auth.Register(ctx, RegisterInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "user_email", in.Body.Email)
	return tokenResp(pair), nil
}

func (h *authHandler) handleLogin(ctx context.Context, in *authLoginReq) (*TokenResp, error) {
	pair, err := h.auth.Login(ctx, LoginInput{Email: in.Body.Email, Password: in.Body.Password})
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	wideevent.Add(ctx, "user_email", in.Body.Email)
	return tokenResp(pair), nil
}

func (h *authHandler) handleRefresh(ctx context.Context, in *authRefreshReq) (*TokenResp, error) {
	pair, err := h.auth.Refresh(ctx, in.Body.RefreshToken)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	return tokenResp(pair), nil
}

func (h *authHandler) handleDeleteAccount(ctx context.Context, in *authDeleteAccountReq) (*struct{}, error) {
	userIDStr := GetUserID(ctx)
	if userIDStr == "" {
		return nil, httpapi.ToHumaErr(ctx, errors.ErrUnauthorized)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, httpapi.ToHumaErr(ctx, errors.ErrUnauthorized)
	}

	if err := h.auth.DeleteAccount(ctx, userID); err != nil {
		return nil, httpapi.ToHumaErr(ctx, err)
	}
	return &struct{}{}, nil
}

func tokenResp(pair TokenPair) *TokenResp {
	r := &TokenResp{}
	r.Body.Data.AccessToken = pair.AccessToken
	r.Body.Data.RefreshToken = pair.RefreshToken
	r.Body.Data.ExpiresIn = pair.ExpiresIn
	return r
}
