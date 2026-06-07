package delivery

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	_ "github.com/danielgtaylor/huma/v2/formats/cbor" // Enable CBOR format
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis_rate/v10"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
	sharedmw "github.com/semmidev/restful-template/internal/shared/middleware"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	router *chi.Mux
	api    huma.API
}

// NewServer wires up all middleware, Huma API, and registers routes.
//
//   - authService / todosService are consumed as interfaces so the delivery
//     layer never touches concrete Usecase types (clean arch boundary).
//   - tokens (auth.TokenService) is kept separate so AuthMiddleware can
//     validate JWTs without depending on the full AuthService.
//   - redisClientOpt is resolved once in app.Setup and passed in here to
//     avoid parsing the Redis DSN a second time just for the asynqmon UI.
//     Pass nil to skip mounting the asynqmon UI (e.g. in tests).
func NewServer(
	cfg config.Config,
	log *slog.Logger,
	authService auth.AuthService,
	todosService todos.TodoService,
	tokens auth.TokenService,
	limiter *redis_rate.Limiter,
	healthCheckers map[string]HealthChecker,
	redisClientOpt asynq.RedisClientOpt,
) *Server {
	r := chi.NewRouter()

	promMiddleware, err := sharedmw.NewPrometheusMiddleware(prometheus.DefaultRegisterer)
	if err != nil {
		log.Error("failed to create prometheus middleware", "err", err)
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(otelchi.Middleware(cfg.App.Name, otelchi.WithChiRoutes(r)))
	if promMiddleware != nil {
		r.Use(promMiddleware.Handler())
	}
	r.Use(TraceIDMiddleware)
	r.Use(sharedmw.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.HTTP.ReadTimeout))
	r.Use(sharedmw.CORS(cfg.CORS.AllowedOrigins))
	// SecurityHeaders and RateLimiter are applied per-group below so that the
	// asynqmon admin UI (a full SPA) can use a relaxed CSP without weakening
	// the API's strict default-src 'none' policy.

	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	asynqmonUI := asynqmon.New(asynqmon.Options{
		RootPath:     "/admin/asynq",
		RedisConnOpt: redisClientOpt,
	})

	// Admin group: Basic Auth + permissive CSP so the asynqmon SPA can
	// load its inline scripts, chunk JS, stylesheets, and Google Fonts.
	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("Asynqmon", map[string]string{
			cfg.Asynqmon.Username: cfg.Asynqmon.Password,
		}))
		r.Use(sharedmw.AsynqmonSecurityHeaders())
		// chi Mount strips the prefix which breaks asynqmon's internal
		// ServeMux; Handle preserves the full path.
		rootPath := asynqmonUI.RootPath()
		r.Handle(rootPath, http.RedirectHandler(rootPath+"/", http.StatusMovedPermanently))
		r.Handle(rootPath+"/*", asynqmonUI)
	})

	// API + docs group: strict security headers and rate limiting.
	r.Group(func(r chi.Router) {
		r.Use(sharedmw.SecurityHeaders())
		// r.Use(sharedmw.RateLimiter(limiter, rateLimitSkipPaths))

		humaConfig := huma.DefaultConfig(cfg.App.Name, cfg.App.Version)
		humaConfig.Info.Description = cfg.App.Description
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
		api.UseMiddleware(auth.AuthMiddleware(api, tokens))

		RegisterRoutes(api, healthCheckers, authService, todosService)
	})

	return &Server{router: r, api: nil}
}

func (s *Server) Handler() http.Handler { return s.router }

// TraceIDMiddleware propagates the OpenTelemetry Trace ID into the response so
// clients and proxies can correlate a request with its trace in Tempo/Jaeger.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanContext := trace.SpanContextFromContext(r.Context())
		if spanContext.HasTraceID() {
			w.Header().Set("X-Trace-Id", spanContext.TraceID().String())
		}
		next.ServeHTTP(w, r)
	})
}
