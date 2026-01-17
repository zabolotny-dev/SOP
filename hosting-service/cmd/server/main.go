package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-contracts/topology"
	"hosting-kit/database"
	"hosting-kit/debug"
	"hosting-kit/messaging"
	"hosting-kit/mid"
	"hosting-kit/otel"
	"hosting-service/cmd/server/graphql"
	"hosting-service/cmd/server/queue"
	"hosting-service/cmd/server/rest"
	"hosting-service/internal/plan"
	"hosting-service/internal/plan/extensions/planotel"
	"hosting-service/internal/plan/stores/plandb"
	"hosting-service/internal/platform/grpc"
	"hosting-service/internal/server"
	"hosting-service/internal/server/extensions/serverotel"
	"hosting-service/internal/server/stores/serverdb"
	"hosting-service/internal/server/stores/servergrpc"
	"hosting-service/internal/server/stores/servermsg"
	"log"
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
			Name         string `conf:"default:sop"`
			MaxOpenConns int    `conf:"default:25"`
		}
		Web struct {
			APIHost      string        `conf:"default:0.0.0.0:8080"`
			DebugHost    string        `conf:"default:0.0.0.0:8010"`
			APIPrefix    string        `conf:"default:/api/hosting"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
		}
		AMQP struct {
			URL            string        `conf:"default:amqp://guest:guest@localhost:5672/,mask,env:AMQP_URL"`
			HandlerTimeout time.Duration `conf:"default:10s"`
			QueueName      string        `conf:"default:api_events_queue"`
		}
		Resources struct {
			Host    string        `conf:"default:hosting-resources-service:2001"`
			Timeout time.Duration `conf:"default:5s"`
		}
		Tempo struct {
			Host        string  `conf:"default:hosting-tempo:4317"`
			ServiceName string  `conf:"default:hosting-service"`
			Probability float64 `conf:"default:0.05"`
		}
	}{}

	const prefix = "SERV"

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
	// Create GRPC Support

	grpcConn, err := grpc.NewClient(cfg.Resources.Host)
	if err != nil {
		return fmt.Errorf("initializing grpc client: %w", err)
	}

	defer grpcConn.Close()

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
	// Create RabbitMQ Support

	rqManager, err := messaging.NewMessageManager(cfg.AMQP.URL, []messaging.ExchangeConfig{
		{
			Name: topology.CommandsExchange,
			Type: messaging.ExchangeDirect,
		},
		{
			Name: topology.EventsExchange,
			Type: messaging.ExchangeTopic,
		},
		{
			Name: topology.DLXExchange,
			Type: messaging.ExchangeDirect,
		},
	}, cfg.AMQP.HandlerTimeout, tracer)
	if err != nil {
		return fmt.Errorf("initializing rabbitmq: %w", err)
	}

	defer func() {
		ctxShut, cancel := context.WithTimeout(ctx, cfg.App.ShutdownTimeout)
		defer cancel()
		if err := rqManager.Stop(ctxShut); err != nil {
			log.Printf("Failed to shutdown rabbit manager: %v", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Create Business Packages

	planOtelExt := planotel.NewExtension()
	planStore := plandb.NewStore(db)
	planBus := plan.NewBusiness(planStore, planOtelExt)

	serverOtelExt := serverotel.NewExtension()
	serverPublisher := servermsg.NewPublisher(rqManager)
	serverStore := serverdb.NewStore(db)
	serverGrpc := servergrpc.NewGrpc(grpcConn, cfg.Resources.Timeout)
	serverBus := server.NewBusiness(serverStore, planBus, serverPublisher, serverGrpc, serverOtelExt)

	// -------------------------------------------------------------------------
	// Start API Service

	mux := chi.NewRouter()

	mux.Use(mid.Otel(tracer))
	mux.Use(middleware.Recoverer)
	mux.Use(mid.Logger)
	mux.Use(mid.Performance)

	rest.RegisterRoutes(mux, rest.Config{
		PlanBus:   planBus,
		ServerBus: serverBus,
		Prefix:    cfg.Web.APIPrefix,
	})

	graphql.RegisterRoutes(mux, graphql.HandlerConfig{
		PlanBus:   planBus,
		ServerBus: serverBus,
	})

	err = queue.RegisterAll(rqManager, queue.Config{
		ServerBus: serverBus,
		QueueName: cfg.AMQP.QueueName,
	})
	if err != nil {
		return fmt.Errorf("registering queue handlers: %w", err)
	}

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 2)

	go func() {
		log.Printf("main: HTTP API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

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
