package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-kit/database"
	"hosting-kit/debug"
	"hosting-kit/mid"
	"hosting-kit/otel"
	"hosting-resources-service/cmd/server/grpc"
	"hosting-resources-service/cmd/server/rest"
	"hosting-resources-service/internal/pool"
	"hosting-resources-service/internal/pool/extensions/poolotel"
	"hosting-resources-service/internal/pool/stores/pooldb"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		App struct {
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:vladick,mask"`
			Host         string `conf:"default:localhost:5432"`
			Name         string `conf:"default:sop_pool"`
			MaxOpenConns int    `conf:"default:25"`
		}
		Web struct {
			APIHost      string        `conf:"default:0.0.0.0:2080"`
			DebugHost    string        `conf:"default:0.0.0.0:2010"`
			GRPCHost     string        `conf:"default:0.0.0.0:2001"`
			APIPrefix    string        `conf:"default:/api/resources"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
		}
		Tempo struct {
			Host        string  `conf:"default:hosting-tempo:4317"`
			ServiceName string  `conf:"default:resource-service"`
			Probability float64 `conf:"default:0.05"`
		}
	}{}

	const prefix = "RES"

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// Database Support

	db, err := database.Open(ctx, database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxOpenConns: cfg.DB.MaxOpenConns,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer db.Close()

	// -------------------------------------------------------------------------
	// Start Tracing Support

	traceProvider, teardown, err := otel.InitTracing(otel.Config{
		ServiceName: cfg.Tempo.ServiceName,
		Host:        cfg.Tempo.Host,
		Probability: cfg.Tempo.Probability,
	})
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := teardown(cleanupCtx); err != nil {
			log.Printf("telemetry shutdown error: %v", err)
		}
	}()

	tracer := traceProvider.Tracer(cfg.Tempo.ServiceName)

	// -------------------------------------------------------------------------
	// Create Business Packages

	poolOtelExt := poolotel.NewExtension()
	poolStore := pooldb.NewStore(db)
	poolBus := pool.NewBusiness(poolStore, poolOtelExt)

	// -------------------------------------------------------------------------
	// Start API Service

	mux := chi.NewRouter()

	mux.Use(mid.Otel(tracer))
	mux.Use(middleware.Recoverer)
	mux.Use(mid.Logger)
	mux.Use(mid.Performance)

	rest.RegisterRoutes(mux, rest.Config{
		PoolBus: poolBus,
		Prefix:  cfg.Web.APIPrefix,
	})

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 3)

	go func() {
		log.Printf("main: HTTP API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Create GRPC Service

	lis, err := net.Listen("tcp", cfg.Web.GRPCHost)
	if err != nil {
		return fmt.Errorf("failed to listen on host %s : %w", cfg.Web.GRPCHost, err)
	}

	grpcApp := grpc.New(grpc.Config{
		PoolBus: poolBus,
		Tracer:  tracer,
	})

	go func() {
		log.Printf("main: gRPC API listening on %s", cfg.Web.GRPCHost)
		serverErrors <- grpcApp.Serve(lis)
	}()

	defer grpcApp.Stop()

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Printf("main: Debug server listening on %s", cfg.Web.DebugHost)
		serverErrors <- http.ListenAndServe(cfg.Web.DebugHost, debug.Mux())
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Printf("main: %v : Start shutdown", sig)
		ctxShut, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctxShut); err != nil {
			api.Close()
			return fmt.Errorf("could not stop http server gracefully: %w", err)
		}
	}

	return nil
}
