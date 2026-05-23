package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer defines the interface for distributed tracing operations.
type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, trace.Span)
}

type OtelTracer struct {
	tracer trace.Tracer
}

func NewOtelTracer(name string) *OtelTracer {
	return &OtelTracer{
		tracer: otel.Tracer(name),
	}
}

func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName)
}
