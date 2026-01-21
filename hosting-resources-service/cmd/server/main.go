package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-kit/auth/kratos"
	"hosting-kit/database"
	"hosting-kit/debug"
	"hosting-kit/logger"
	"hosting-kit/mid"
	"hosting-kit/otel"
	"hosting-resources-service/cmd/server/grpc"
	"hosting-resources-service/cmd/server/rest"
	"hosting-resources-service/internal/pool"
	"hosting-resources-service/internal/pool/extensions/poolotel"
	"hosting-resources-service/internal/pool/stores/pooldb"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()

	log := logger.New(os.Stdout, logger.LevelInfo, "resources-service", otel.GetTraceID)

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

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
		Auth struct {
			Host string `conf:"default:http://hosting-kratos:4433"`
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
			log.Error(cleanupCtx, "telemetry shutdown error", "error", err)
		}
	}()

	tracer := traceProvider.Tracer(cfg.Tempo.ServiceName)

	// -------------------------------------------------------------------------
	// Create Business Packages

	poolOtelExt := poolotel.NewExtension()
	poolStore := pooldb.NewStore(db)
	poolBus := pool.NewBusiness(poolStore, poolOtelExt)

	// -------------------------------------------------------------------------
	// Initialize authentication support

	authClient := kratos.New(cfg.Auth.Host)

	// -------------------------------------------------------------------------
	// Start API Service

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(mid.Otel(tracer))
	mux.Use(mid.Logger(log))
	mux.Use(mid.Performance(log))

	rest.RegisterRoutes(mux, rest.Config{
		PoolBus:    poolBus,
		Prefix:     cfg.Web.APIPrefix,
		AuthClient: authClient,
		Log:        log,
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
		log.Info(ctx, "HTTP API listening", "addr", api.Addr)
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
		Log:     log,
	})

	go func() {
		log.Info(ctx, "GRPC API listening", "addr", cfg.Web.GRPCHost)
		serverErrors <- grpcApp.Serve(lis)
	}()

	defer grpcApp.Stop()

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "HTTP Debug listening", "addr", cfg.Web.DebugHost)
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
		log.Info(ctx, "start shutdown", "signal", sig.String())
		ctxShut, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctxShut); err != nil {
			api.Close()
			return fmt.Errorf("could not stop http server gracefully: %w", err)
		}
	}

	return nil
}
