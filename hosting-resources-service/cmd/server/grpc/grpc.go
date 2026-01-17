package grpc

import (
	"hosting-resources-service/cmd/server/grpc/handlers/poolgrp"
	"hosting-resources-service/internal/platform/grpc"
	"hosting-resources-service/internal/pool"
	"net"

	"go.opentelemetry.io/otel/trace"
	grpclib "google.golang.org/grpc"
)

type Config struct {
	PoolBus pool.ExtBusiness
	Tracer  trace.Tracer
}

type App struct {
	server *grpclib.Server
}

func New(cfg Config) *App {
	gs := grpc.NewServer(cfg.Tracer)

	poolgrp.Register(gs, cfg.PoolBus)

	return &App{
		server: gs,
	}
}

func (a *App) Serve(lis net.Listener) error {
	return a.server.Serve(lis)
}

func (a *App) Stop() {
	a.server.GracefulStop()
}
