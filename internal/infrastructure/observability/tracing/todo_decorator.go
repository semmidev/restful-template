package tracing

import (
	"context"

	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TodoServiceDecorator struct {
	base   domain.TodoUsecase
	tracer trace.Tracer
}

func NewTodoServiceDecorator(base domain.TodoUsecase) *TodoServiceDecorator {
	return &TodoServiceDecorator{
		base:   base,
		tracer: otel.Tracer("todo_service"),
	}
}

func (d *TodoServiceDecorator) Create(ctx context.Context, in domain.CreateTodoInput) (*domain.Todo, error) {
	ctx, span := d.tracer.Start(ctx, "TodoService.Create")
	defer span.End()

	res, err := d.base.Create(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *TodoServiceDecorator) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Todo, error) {
	ctx, span := d.tracer.Start(ctx, "TodoService.Get")
	defer span.End()

	res, err := d.base.Get(ctx, userID, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *TodoServiceDecorator) List(ctx context.Context, q domain.ListTodosQuery) ([]*domain.Todo, int, error) {
	ctx, span := d.tracer.Start(ctx, "TodoService.List")
	defer span.End()

	res, count, err := d.base.List(ctx, q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, count, err
}

func (d *TodoServiceDecorator) Update(ctx context.Context, in domain.UpdateTodoInput) (*domain.Todo, error) {
	ctx, span := d.tracer.Start(ctx, "TodoService.Update")
	defer span.End()

	res, err := d.base.Update(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *TodoServiceDecorator) Delete(ctx context.Context, userID, id uuid.UUID) error {
	ctx, span := d.tracer.Start(ctx, "TodoService.Delete")
	defer span.End()

	err := d.base.Delete(ctx, userID, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
