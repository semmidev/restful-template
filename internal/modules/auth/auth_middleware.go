package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

// Context keys are now in httpapi

// AuthMiddleware is a Huma middleware that validates Bearer JWTs on protected routes.
// Depends on TokenService (not Usecase) — clean dependency flow.
func AuthMiddleware(api huma.API, tokens TokenService) func(ctx huma.Context, next func(huma.Context)) {
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

		// Enrich the canonical wide event with the authenticated user's identity.
		// This means every log line for an authenticated request carries user_id
		// and user_email without any handler needing to log it explicitly.
		wideevent.Add(ctx.Context(), "user_id", claims.UserID.String())
		wideevent.Add(ctx.Context(), "user_email", claims.Email)

		ctx = huma.WithValue(ctx, httpapi.UserIDKey, claims.UserID.String())
		ctx = huma.WithValue(ctx, httpapi.UserEmailKey, claims.Email)
		next(ctx)
	}
}

func isPublicPath(path string) bool {
	return path == "/api/v1/auth/login" ||
		path == "/api/v1/auth/register" ||
		path == "/api/v1/auth/refresh" ||
		strings.HasPrefix(path, "/docs") ||
		path == "/openapi.json" ||
		path == "/api/v1/health"
}

// GetUserID extracts the authenticated user ID from a standard context.
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(httpapi.UserIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetUserEmail extracts the authenticated user email from a standard context.
func GetUserEmail(ctx context.Context) string {
	if v := ctx.Value(httpapi.UserEmailKey); v != nil {
		return v.(string)
	}
	return ""
}
