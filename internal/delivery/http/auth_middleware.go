package delivery

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/domain"
)

type ctxKey string

const (
	userIDCtxKey    ctxKey = "user_id"
	userEmailCtxKey ctxKey = "user_email"
)

// AuthMiddleware is a Huma middleware that validates Bearer JWTs on protected routes.
// Depends on domain.TokenService (not domain.AuthUsecase) — clean dependency flow.
func AuthMiddleware(api huma.API, tokens domain.TokenService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		path := ctx.URL().Path

		// Public paths — skip auth
		if isPublicPath(path) {
			next(ctx)
			return
		}

		authHeader := ctx.Header("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing or malformed authorization header")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.ParseAccess(ctx.Context(), token)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx = huma.WithValue(ctx, userIDCtxKey, claims.UserID.String())
		ctx = huma.WithValue(ctx, userEmailCtxKey, claims.Email)
		next(ctx)
	}
}

func isPublicPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/auth") ||
		strings.HasPrefix(path, "/docs") ||
		path == "/openapi.json" ||
		path == "/api/v1/health"
}

// GetUserID extracts the authenticated user ID from a standard context.
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDCtxKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetUserEmail extracts the authenticated user email from a standard context.
func GetUserEmail(ctx context.Context) string {
	if v := ctx.Value(userEmailCtxKey); v != nil {
		return v.(string)
	}
	return ""
}
