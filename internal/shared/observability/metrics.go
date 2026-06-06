package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AuthRegistrationsTotal tracks the number of successful user registrations.
	AuthRegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_registrations_total",
		Help: "The total number of successful user registrations",
	})

	// TodosCreatedTotal tracks the number of newly created todos.
	TodosCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "todos_created_total",
		Help: "The total number of todos created",
	})

	// CacheHitsTotal tracks the number of cache hits for domain entities.
	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "The total number of cache hits",
	}, []string{"entity"})

	// CacheMissesTotal tracks the number of cache misses for domain entities.
	CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "The total number of cache misses",
	}, []string{"entity"})
)
