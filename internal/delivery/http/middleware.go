package delivery

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"
)

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"req_id", middleware.GetReqID(r.Context()),
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
