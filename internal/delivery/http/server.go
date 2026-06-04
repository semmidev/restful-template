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
// tokens (auth.TokenService) is passed separately so AuthMiddleware can
// validate JWTs without depending on the full AuthService (clean arch).
func NewServer(
	cfg config.Config,
	log *slog.Logger,
	authUsecase *auth.Usecase,
	todosUsecase *todos.Usecase,
	tokens auth.TokenService,
	limiter *redis_rate.Limiter,
	healthCheckers map[string]HealthChecker,
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
	r.Use(sharedmw.SecurityHeaders())
	r.Use(sharedmw.RateLimiter(limiter))

	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.DSN)
	if err == nil {
		clientOpt, ok := redisOpt.(asynq.RedisClientOpt)
		if ok {
			asynqmonUI := asynqmon.New(asynqmon.Options{
				RootPath:     "/admin/asynq",
				RedisConnOpt: clientOpt,
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.BasicAuth("Asynqmon", map[string]string{
					cfg.Asynqmon.Username: cfg.Asynqmon.Password,
				}))
				r.Mount(asynqmonUI.RootPath(), asynqmonUI)
			})
		} else {
			log.Error("parsed redis DSN is not a RedisClientOpt")
		}
	} else {
		log.Error("failed to parse redis DSN for asynqmon", "err", err)
	}

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

	RegisterRoutes(api, healthCheckers, authUsecase, todosUsecase)

	return &Server{router: r, api: api}
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
