package observability

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTelemetry configures OpenTelemetry tracing and metrics.
// It returns a shutdown function to be called on graceful exit.
func InitTelemetry(ctx context.Context, serviceName, serviceVersion, otlpEndpoint string) (func(context.Context) error, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// 1. Set up Tracing
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(), // Adjust for production
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Combine shutdown
	shutdown := func(shutdownCtx context.Context) error {
		var errs error
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			errs = errors.Join(errs, err)
		}
		return errs
	}

	return shutdown, nil
}
