package delivery

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"
	"github.com/riandyrn/otelchi"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/domain"
	"go.opentelemetry.io/otel/trace"
)

// Server holds the chi router and huma API.
type Server struct {
	router *chi.Mux
	api    huma.API
}

// NewServer wires up all middleware, Huma API, and registers routes.
// tokens (domain.TokenService) is passed separately so AuthMiddleware can
// validate JWTs without depending on the full AuthService (clean arch).
func NewServer(cfg config.Config, log *slog.Logger, auth domain.AuthUsecase, todos domain.TodoUsecase, tokens domain.TokenService, limiter *redis_rate.Limiter) *Server {
	r := chi.NewRouter()

	// Middleware stack (order matters)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(otelchi.Middleware(cfg.App.Name, otelchi.WithChiRoutes(r)))
	r.Use(TraceIDMiddleware)
	r.Use(Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.HTTP.ReadTimeout))
	r.Use(CORS())
	r.Use(SecurityHeaders())
	r.Use(RateLimiter(limiter))

	humaConfig := huma.DefaultConfig("Todo API", cfg.App.Version)
	humaConfig.Info.Description = "Production-ready Todo REST API built with Huma v2 + Chi."
	humaConfig.Components = &huma.Components{
		SecuritySchemes: map[string]*huma.SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	}

	api := humachi.New(r, humaConfig)
	api.UseMiddleware(AuthMiddleware(api, tokens))

	RegisterRoutes(api, auth, todos, log)

	return &Server{router: r, api: api}
}

func (s *Server) Handler() http.Handler { return s.router }

// TraceIDMiddleware extracts the OpenTelemetry Trace ID from the context
// and injects it into the HTTP response header for debugging.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanContext := trace.SpanContextFromContext(r.Context())
		if spanContext.HasTraceID() {
			w.Header().Set("X-Trace-Id", spanContext.TraceID().String())
		}
		next.ServeHTTP(w, r)
	})
}
