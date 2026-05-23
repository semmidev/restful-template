package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	once            sync.Once
)

// PrometheusMiddleware holds the prometheus metrics and registry.
type PrometheusMiddleware struct {
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewPrometheusMiddleware creates a new PrometheusMiddleware.
func NewPrometheusMiddleware(reg prometheus.Registerer) (*PrometheusMiddleware, error) {
	var errCount, errDuration error

	once.Do(func() {
		requestCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed.",
			},
			[]string{"method", "path", "status"},
		)

		requestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		)

		errCount = reg.Register(requestCount)
		errDuration = reg.Register(requestDuration)
	})

	if errCount != nil {
		var are prometheus.AlreadyRegisteredError
		if !errors.As(errCount, &are) {
			return nil, errCount
		}
		requestCount = are.ExistingCollector.(*prometheus.CounterVec)
	}

	if errDuration != nil {
		var are prometheus.AlreadyRegisteredError
		if !errors.As(errDuration, &are) {
			return nil, errDuration
		}
		requestDuration = are.ExistingCollector.(*prometheus.HistogramVec)
	}

	return &PrometheusMiddleware{
		requestCount:    requestCount,
		requestDuration: requestDuration,
	}, nil
}

// Handler returns the chi middleware handler.
func (m *PrometheusMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Exclude /metrics from being counted
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Use chi's built-in response wrapper to capture status code
			// Wait, chi doesn't have a built-in response wrapper exposed easily, we'll need a custom wrapper
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			routeContext := chi.RouteContext(r.Context())
			route := "UNMATCHED"
			if routeContext != nil && routeContext.RoutePattern() != "" {
				route = routeContext.RoutePattern()
			}

			path := r.URL.Path

			statusStr := strconv.Itoa(rw.status)
			method := strings.Clone(r.Method)

			m.requestCount.WithLabelValues(
				method,
				path,
				statusStr,
			).Inc()

			m.requestDuration.WithLabelValues(
				method,
				route,
				statusStr,
			).Observe(duration)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
