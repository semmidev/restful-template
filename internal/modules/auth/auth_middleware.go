package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/semmidev/restful-template/internal/shared/httpapi"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
)

// publicPaths is a set of exact paths that bypass authentication.
// RegisterPublicPath allows route registration functions to self-declare their
// public status — single source of truth, no fragile hardcoded list here.
//
// point 14: was previously a chained string comparison which required manual
// updates whenever a new public route was added (OCP violation) and did O(n)
// linear scan.
var publicPaths = map[string]struct{}{
	"/api/v1/auth/login":    {},
	"/api/v1/auth/register": {},
	"/api/v1/auth/refresh":  {},
	"/api/v1/health":        {},
	"/openapi.json":         {},
	// /metrics is intentionally public here but MUST be firewalled in production
	// via network policy so it is not reachable from the internet.
	"/metrics": {},
}

// publicPrefixes covers path prefixes (e.g. /docs/...) that are always public.
var publicPrefixes = []string{"/docs"}

// RegisterPublicPath registers a path as bypassing JWT authentication.
// Call this from route registration functions for any endpoint that should
// be accessible without a token.
func RegisterPublicPath(path string) {
	publicPaths[path] = struct{}{}
}

// AuthMiddleware is a Huma middleware that validates Bearer JWTs on protected routes.
// Depends on TokenService (not Usecase) — clean dependency flow.
func AuthMiddleware(api huma.API, tokens TokenService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		path := ctx.URL().Path

		// Fast O(1) set lookup for exact-match public paths
		if _, ok := publicPaths[path]; ok {
			next(ctx)
			return
		}
		// Prefix check for /docs/...
		for _, prefix := range publicPrefixes {
			if strings.HasPrefix(path, prefix) {
				next(ctx)
				return
			}
		}

		authHeader := ctx.Header("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "missing or malformed authorization header")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.ParseAccess(ctx.Context(), token)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Enrich the canonical wide event with the authenticated user's identity.
		// This means every log line for an authenticated request carries user_id
		// and user_email without any handler needing to log it explicitly.
		wideevent.Add(ctx.Context(), "user_id", claims.UserID.String())
		wideevent.Add(ctx.Context(), "user_email", claims.Email)

		ctx = huma.WithValue(ctx, httpapi.UserIDKey, claims.UserID)
		ctx = huma.WithValue(ctx, httpapi.UserEmailKey, claims.Email)
		next(ctx)
	}
}

// GetUserID extracts the authenticated user ID from a standard context.
func GetUserID(ctx context.Context) string {
	id, err := httpapi.ExtractUserID(ctx)
	if err != nil {
		return ""
	}
	return id.String()
}

// GetUserEmail extracts the authenticated user email from a standard context.
func GetUserEmail(ctx context.Context) string {
	if v := ctx.Value(httpapi.UserEmailKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
