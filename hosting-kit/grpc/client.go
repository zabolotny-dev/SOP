package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(url string, log Logger) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithUnaryInterceptor(clientLoggingInterceptor(log)))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
