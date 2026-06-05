package asynqtask

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// TaskHandler is the domain-facing handler signature.
// Modules implement this instead of the raw asynq handler so they never need to
// import github.com/hibiken/asynq directly — the asynq library stays contained
// within this package.
//
// payload holds the raw JSON bytes of the task; the handler is responsible for
// unmarshalling into its own payload struct.
type TaskHandler func(ctx context.Context, payload []byte) error

// ErrSkipRetry is a re-export of asynq.SkipRetry so module handlers can signal
// "do not retry" without importing the asynq library directly.
var ErrSkipRetry = asynq.SkipRetry

// WrapHandler adapts a TaskHandler to the asynq internal handler signature.
// Called internally by Processor.AddTask so callers never touch *asynq.Task.
func WrapHandler(h TaskHandler) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		return h(ctx, t.Payload())
	}
}

// Processor manages the asynq worker lifecycle.
// Task handlers are registered via AddTask before Start is called.
// The concrete implementation is intentionally unexported — callers depend on
// this interface, which makes it easy to swap or mock in tests.
type Processor interface {
	// AddTask registers a TaskHandler for the given task type pattern.
	// All AddTask calls must happen before Start.
	AddTask(pattern string, handler TaskHandler)
	// Start begins consuming tasks from the queue. Blocks until an error occurs.
	Start() error
	// Shutdown gracefully drains in-flight tasks and stops the server.
	Shutdown()
}

type redisProcessor struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	logger *slog.Logger
}

// NewProcessor constructs a Redis-backed Processor with priority queues.
// Register all task handlers with AddTask before calling Start.
func NewProcessor(redisOpt asynq.RedisClientOpt, logger *slog.Logger) Processor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			Concurrency: 10,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.ErrorContext(ctx, "process task failed",
					slog.Any("error", err),
					slog.String("type", task.Type()),
					slog.String("payload", string(task.Payload())),
				)
			}),
		},
	)

	return &redisProcessor{
		server: server,
		mux:    asynq.NewServeMux(),
		logger: logger,
	}
}

// AddTask registers a TaskHandler for the given task type pattern.
// WrapHandler adapts the domain-facing TaskHandler to asynq's internal signature
// so *asynq.Task never leaks into module code.
func (p *redisProcessor) AddTask(pattern string, handler TaskHandler) {
	p.mux.HandleFunc(pattern, WrapHandler(handler))
}

// Start begins listening for tasks using all registered handlers.
func (p *redisProcessor) Start() error {
	return p.server.Start(p.mux)
}

// Shutdown waits for in-flight tasks to finish before stopping.
func (p *redisProcessor) Shutdown() {
	p.server.Shutdown()
}
