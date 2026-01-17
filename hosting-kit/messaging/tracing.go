package messaging

import (
	"context"

	otelglobal "go.opentelemetry.io/otel"
)

func InjectTraceHeaders(ctx context.Context) map[string]interface{} {
	headers := make(map[string]interface{})
	carrier := AMQPHeaderCarrier(headers)

	otelglobal.GetTextMapPropagator().Inject(ctx, carrier)

	return headers
}

func ExtractTraceHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	if headers == nil {
		return ctx
	}

	carrier := AMQPHeaderCarrier(headers)

	ctx = otelglobal.GetTextMapPropagator().Extract(ctx, carrier)

	return ctx
}
