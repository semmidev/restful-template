package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
	"go.opentelemetry.io/otel/trace"
)

// logSkipPaths contains exact paths that should not produce a canonical log line.
// These are infrastructure/observability endpoints that are polled frequently
// and would generate noise without any business value.
var logSkipPaths = map[string]struct{}{
	"/metrics":      {},
	"/docs":         {},
	"/openapi.yaml": {},
	"/favicon.ico":  {},
	"/health":       {},
	// Chrome DevTools auto-probe — not a real request
	"/.well-known/appspecific/com.chrome.devtools.json": {},
}

// logSkipPrefixes contains path prefixes whose requests should not produce a
// canonical log line. Use for entire sub-trees (e.g. admin UIs with many static
// assets) where exact-path enumeration is impractical.
var logSkipPrefixes = []string{
	"/admin/asynq/", // asynqmon SPA static assets & polling endpoints
}

// Logger is the canonical wide-event middleware.
// It initialises a fresh WideEvent in the request context, runs the handler,
// then emits ONE structured log line containing every field accumulated during
// the request — HTTP metadata, authenticated identity, business context, timing,
// outcome, and any error detail.
//
// This implements the "Canonical Log Line" pattern from loggingsucks.com:
// one wide event per request instead of many scattered log statements.
//
// Paths in logSkipPaths are served normally but produce no log output.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for noisy infrastructure endpoints (exact match).
			if _, skip := logSkipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}
			// Skip logging for entire sub-trees (prefix match).
			for _, prefix := range logSkipPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()

			ctx := wideevent.New(r.Context())
			r = r.WithContext(ctx)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			reqID := middleware.GetReqID(r.Context())
			var traceID string
			if sc := trace.SpanContextFromContext(r.Context()); sc.HasTraceID() {
				traceID = sc.TraceID().String()
			}

			wideevent.Add(ctx, "request_id", reqID)
			wideevent.Add(ctx, "trace_id", traceID)
			wideevent.Add(ctx, "method", r.Method)
			wideevent.Add(ctx, "path", r.URL.Path)

			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			durationMS := time.Since(start).Milliseconds()

			outcome := "success"
			level := slog.LevelInfo
			if status >= 500 {
				outcome = "error"
				level = slog.LevelError
			} else if status >= 400 {
				outcome = "failure"
				level = slog.LevelWarn
			}

			wideevent.Emit(ctx, log, level, "request",
				"status", status,
				"duration_ms", durationMS,
				"bytes", ww.BytesWritten(),
				"outcome", outcome,
			)
		})
	}
}

// CORS returns a middleware that enforces the given allowedOrigins list.
// Pass []string{"*"} for development (allows all origins).
// In production, pass explicit origins: []string{"https://app.example.com"}.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if _, ok := originSet[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets hardened HTTP security headers on every response.
// CSP uses default-src 'none' because this is a pure API server; adjust if
// you ever serve HTML from the same origin.
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")                                         // Mencegah browser mencoba "menebak" tipe file (MIME type).
			w.Header().Set("X-Frame-Options", "DENY")                                                   // Mencegah clickjacking. Mencegah halaman web dimuat di dalam frame/iframe.
			w.Header().Set("Referrer-Policy", "no-referrer")                                            // Mengontrol berapa banyak informasi referrer yang dikirim ke server tujuan.
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload") // Memaksa browser hanya berkomunikasi via HTTPS.
			w.Header().Set("Content-Security-Policy", "default-src 'none'")                             // Mengontrol sumber daya yang boleh dimuat oleh halaman.
			next.ServeHTTP(w, r)
		})
	}
}

// AsynqmonSecurityHeaders sets permissive HTTP security headers for the
// asynqmon admin UI, allowing it to load inline scripts, stylesheets, and fonts.
func AsynqmonSecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline'; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"img-src 'self' data:; "+
					"connect-src 'self'; "+
					"manifest-src 'self'")
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter enforces a per-IP rate limit using Redis.
// skipPaths is an optional set of exact paths that bypass rate limiting
// (e.g. health-check endpoints polled by Kubernetes probes). The map
// is defined by the caller so this package stays unaware of route structure.
func RateLimiter(limiter *redis_rate.Limiter, skipPaths map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skip := skipPaths[r.URL.Path]; skip {
				next.ServeHTTP(w, r)
				return
			}

			// X-Real-IP is populated by chi's RealIP middleware from X-Forwarded-For.
			// Falling back to RemoteAddr means the load-balancer IP is used behind a
			// proxy — acceptable only in local development.
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				clientIP = r.RemoteAddr
			}

			res, err := limiter.Allow(r.Context(), fmt.Sprintf("rate:%s", clientIP), redis_rate.PerSecond(5))
			if err != nil {
				// Redis is unavailable — fail open rather than taking the service down.
				next.ServeHTTP(w, r)
				return
			}
			if res.Allowed == 0 {
				w.Header().Set("Content-Type", "application/problem+json")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(res.ResetAfter.Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"status": 429, "title": "Too Many Requests", "code": "RATE_LIMITED", "detail": "Rate limit exceeded, please slow down"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
