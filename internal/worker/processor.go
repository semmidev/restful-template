package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/semmidev/restful-template/internal/shared/asynqtask"
)

// TaskProcessor defines the interface for an async worker that manages
// life-cycles and handler registrations for background jobs.
type TaskProcessor interface {
	// Start begins non-blocking consumption of tasks.
	Start() error
	// Shutdown gracefully halts task processing.
	Shutdown()
	// ProcessTaskSendWelcomeEmail decodes the payload and simulates sending a welcome email.
	ProcessTaskSendWelcomeEmail(ctx context.Context, task *asynq.Task) error
}

// RedisTaskProcessor implements TaskProcessor using an Asynq server.
type RedisTaskProcessor struct {
	server *asynq.Server
	logger *slog.Logger
}

// NewRedisTaskProcessor initializes an Asynq server with priority levels.
func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	logger *slog.Logger,
) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				asynqtask.QueueCritical: 10,
				asynqtask.QueueDefault:  5,
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

	return &RedisTaskProcessor{
		server: server,
		logger: logger,
	}
}

// Start registers task handlers and blocks while listening for new work.
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(asynqtask.TaskSendWelcomeEmail, processor.ProcessTaskSendWelcomeEmail)

	return processor.server.Start(mux)
}

// Shutdown ensures all in-flight tasks are completed before exiting.
func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
