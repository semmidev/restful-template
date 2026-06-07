package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type authHandler struct {
	auth AuthService
}

func RegisterAuthRoutes(api huma.API, auth AuthService) {
	h := &authHandler{auth: auth}

	RegisterPublicPath("/api/v1/auth/google")
	RegisterPublicPath("/api/v1/auth/google/config")

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

	huma.Register(api, huma.Operation{
		OperationID: "auth-google-login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/google",
		Summary:     "Login with Google",
		Tags:        []string{"Auth"},
	}, h.handleGoogleLogin)

	huma.Register(api, huma.Operation{
		OperationID: "auth-google-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/google/config",
		Summary:     "Get Google OAuth Configuration",
		Tags:        []string{"Auth"},
	}, h.handleGoogleConfig)
}
