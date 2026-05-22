package delivery

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"
	"github.com/semmidev/restful-template/internal/shared/wideevent"
	"go.opentelemetry.io/otel/trace"
)


// Logger is the canonical wide-event middleware.
// It initialises a fresh WideEvent in the request context, runs the handler,
// then emits ONE structured log line containing every field accumulated during
// the request — HTTP metadata, authenticated identity, business context, timing,
// outcome, and any error detail.
//
// This implements the "Canonical Log Line" pattern from loggingsucks.com:
// one wide event per request instead of many scattered log statements.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Initialise the wide event for this request.
			ctx := wideevent.New(r.Context())
			r = r.WithContext(ctx)

			// Wrap the writer so we can read the status code afterwards.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Seed the event with infrastructure context available at the start.
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

			// Determine outcome and log level from HTTP status.
			outcome := "success"
			level := slog.LevelInfo
			if status >= 500 {
				outcome = "error"
				level = slog.LevelError
			} else if status >= 400 {
				outcome = "failure"
				level = slog.LevelWarn
			}

			// Emit the single canonical wide event.
			wideevent.Emit(ctx, log, level, "request",
				"status", status,
				"duration_ms", durationMS,
				"bytes", ww.BytesWritten(),
				"outcome", outcome,
			)
		})
	}
}


func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
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

func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimiter(limiter *redis_rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res, err := limiter.Allow(r.Context(), "rate_limit:"+r.RemoteAddr, redis_rate.PerSecond(5))
			if err != nil {
				// Log error and fallback to allowing request (fail open)
				next.ServeHTTP(w, r)
				return
			}
			if res.Allowed == 0 {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"status": 429, "title": "Too Many Requests", "detail": "Rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
