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
// was previously a chained string comparison which required manual
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

// AuthMiddleware validates Bearer JWTs on protected routes.
// TokenService is injected instead of the full Usecase so this middleware
// has no dependency on business logic or the database.
func AuthMiddleware(api huma.API, tokens TokenService) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		path := ctx.URL().Path

		if _, ok := publicPaths[path]; ok {
			next(ctx)
			return
		}
		for _, prefix := range publicPrefixes {
			if strings.HasPrefix(path, prefix) {
				next(ctx)
				return
			}
		}

		rawCookie := ctx.Header("Cookie")
		req := &http.Request{Header: http.Header{"Cookie": {rawCookie}}}
		cookie, err := req.Cookie("access_token")
		if err != nil {
			httpapi.WriteHumaErr(api, ctx, httpapi.ToHumaErrUnauthorized("missing or malformed authorization cookie"))
			return
		}

		token := cookie.Value
		claims, err := tokens.ParseAccess(ctx.Context(), token)
		if err != nil {
			httpapi.WriteHumaErr(api, ctx, httpapi.ToHumaErrUnauthorized("invalid or expired token"))
			return
		}

		// Enrich the wide event so every authenticated request's log line carries
		// user identity without handlers needing to add it manually.
		wideevent.Add(ctx.Context(), "user_id", claims.UserID.String())
		wideevent.Add(ctx.Context(), "user_email", claims.Email)

		ctx = huma.WithValue(ctx, httpapi.UserIDKey, claims.UserID)
		ctx = huma.WithValue(ctx, httpapi.UserEmailKey, claims.Email)
		next(ctx)
	}
}

// GetUserEmail retrieves the authenticated user's email from the request context.
func GetUserEmail(ctx context.Context) string {
	if v := ctx.Value(httpapi.UserEmailKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
