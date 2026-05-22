package domain

import "context"

// Tracer defines a domain-agnostic interface for tracing execution boundaries.
// It allows the usecase layer to be instrumented without depending on OpenTelemetry.
type Tracer interface {
	// Start begins a new span and returns a context containing the span.
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

// Span defines a domain-agnostic interface for a tracing span.
type Span interface {
	// End completes the span.
	End()
	// RecordError records an error on the span and sets its status to error.
	RecordError(err error)
}
