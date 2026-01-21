package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func NewServer(trace trace.Tracer, log Logger) *grpc.Server {
	gs := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			traceIDInterceptor(trace),
			serverLoggingInterceptor(log),
		),
	)

	return gs
}
