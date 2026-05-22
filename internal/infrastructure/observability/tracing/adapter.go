package tracing

import (
	"context"

	"github.com/semmidev/restful-template/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// OtelTracer implements domain.Tracer using OpenTelemetry.
type OtelTracer struct {
	tracer trace.Tracer
}

// NewOtelTracer creates a new OtelTracer with the given instrumentation name.
func NewOtelTracer(name string) domain.Tracer {
	return &OtelTracer{
		tracer: otel.Tracer(name),
	}
}

func (t *OtelTracer) Start(ctx context.Context, spanName string) (context.Context, domain.Span) {
	ctx, span := t.tracer.Start(ctx, spanName)
	return ctx, &OtelSpan{span: span}
}

// OtelSpan implements domain.Span.
type OtelSpan struct {
	span trace.Span
}

func (s *OtelSpan) End() {
	s.span.End()
}

func (s *OtelSpan) RecordError(err error) {
	if err != nil {
		s.span.RecordError(err)
		s.span.SetStatus(codes.Error, err.Error())
	}
}
