package grpc

import (
	"context"
	"hosting-kit/otel"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

func traceIDInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = otel.InjectTracing(ctx, tracer)

		return handler(ctx, req)
	}
}
func clientLoggingInterceptor(log Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error(ctx, "grpc client call failed",
				"method", method,
				"status", st.Code().String(),
				"duration", duration.String(),
				"error", err,
			)
		} else {
			log.Info(ctx, "grpc client call success",
				"method", method,
				"status", "OK",
				"duration", duration.String(),
			)
		}
		return err
	}
}

func serverLoggingInterceptor(log Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error(ctx, "grpc server request failed",
				"method", info.FullMethod,
				"status", st.Code().String(),
				"duration", duration.String(),
				"error", err,
			)
		} else {
			log.Info(ctx, "grpc server request success",
				"method", info.FullMethod,
				"status", "OK",
				"duration", duration.String(),
			)
		}
		return resp, err
	}
}
