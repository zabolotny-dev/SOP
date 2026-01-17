package grpc

import (
	"context"

	"hosting-kit/otel"

	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func traceIDInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = otel.InjectTracing(ctx, tracer)

		return handler(ctx, req)
	}
}

func NewServer(trace trace.Tracer) *grpc.Server {
	gs := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(traceIDInterceptor(trace)),
	)

	return gs
}
