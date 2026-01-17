package otel

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const defaultTraceID = "00000000000000000000000000000000"

type Config struct {
	ServiceName string
	Host        string
	Probability float64
}

func InitTracing(cfg Config) (trace.TracerProvider, func(ctx context.Context) error, error) {

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(cfg.Host),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	var traceProvider trace.TracerProvider
	teardown := func(ctx context.Context) error { return nil }

	switch cfg.Host {
	case "":
		traceProvider = noop.NewTracerProvider()

	default:
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Probability))),
			sdktrace.WithBatcher(exporter,
				sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
				sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			),
			sdktrace.WithResource(
				resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNameKey.String(cfg.ServiceName),
				),
			),
		)

		teardown = func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		}

		traceProvider = tp
	}

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider, teardown, nil
}

func InjectTracing(ctx context.Context, tracer trace.Tracer) context.Context {
	ctx = setTracer(ctx, tracer)

	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	if traceID == defaultTraceID {
		traceID = uuid.NewString()
	}
	ctx = setTraceID(ctx, traceID)

	return ctx
}

func AddSpan(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer, ok := ctx.Value(tracerKey).(trace.Tracer)
	if !ok || tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := tracer.Start(ctx, spanName)

	span.SetAttributes(keyValues...)

	return ctx, span
}
