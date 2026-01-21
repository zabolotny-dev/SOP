package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-contracts/topology"
	"hosting-kit/auth/kratos"
	"hosting-kit/debug"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-kit/mid"
	"hosting-kit/otel"
	"hosting-notification-service/cmd/server/queue"
	"hosting-notification-service/cmd/server/rest"
	"hosting-notification-service/internal/notification"
	"hosting-notification-service/internal/notification/adapters"
	"hosting-notification-service/internal/platform/websocket"
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

	log := logger.New(os.Stdout, logger.LevelInfo, "notification-service", otel.GetTraceID)

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
		Auth struct {
			Host string `conf:"default:http://hosting-kratos:4433"`
		}
		Web struct {
			APIHost      string        `conf:"default:0.0.0.0:1080"`
			DebugHost    string        `conf:"default:0.0.0.0:1010"`
			APIPrefix    string        `conf:"default:/api/notification"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
		}
		AMQP struct {
			URL            string        `conf:"default:amqp://guest:guest@localhost:5672/,mask,env:AMQP_URL"`
			HandlerTimeout time.Duration `conf:"default:10s"`
			QueueName      string        `conf:"default:api_notifications_queue"`
		}
		WebSocket struct {
			PingInterval   time.Duration `conf:"default:50s"`
			PongWait       time.Duration `conf:"default:60s"`
			WriteWait      time.Duration `conf:"default:10s"`
			MaxMessageSize int64         `conf:"default:524288"`
		}
		Tempo struct {
			Host        string  `conf:"default:hosting-tempo:4317"`
			ServiceName string  `conf:"default:notification-service"`
			Probability float64 `conf:"default:0.05"`
		}
	}{}

	const prefix = "NOTI"

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return fmt.Errorf("parsing config: %w", err)
	}

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
	// Create RabbitMQ Support

	rqManager, err := messaging.NewMessageManager(cfg.AMQP.URL, []messaging.ExchangeConfig{
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
			log.Error(ctxShut, "failed to shutdown rabbit manager", "error", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Initialize authentication support

	authClient := kratos.New(cfg.Auth.Host)

	// -------------------------------------------------------------------------
	// Create Business Packages

	wsConfig := websocket.Config{
		PingInterval:   cfg.WebSocket.PingInterval,
		PongWait:       cfg.WebSocket.PongWait,
		WriteWait:      cfg.WebSocket.WriteWait,
		MaxMessageSize: cfg.WebSocket.MaxMessageSize,
	}
	wsHub := websocket.NewHub(wsConfig)
	wsSender := adapters.NewWSAdapter(wsHub)

	notiBus := notification.New(wsSender)

	// -------------------------------------------------------------------------
	// Start API Service

	mux := chi.NewRouter()

	mux.Use(mid.Otel(tracer))
	mux.Use(middleware.Recoverer)
	mux.Use(mid.Logger(log))
	mux.Use(mid.Performance(log))

	rest.RegisterRoutes(mux, rest.Config{
		Prefix:     cfg.Web.APIPrefix,
		AuthClient: authClient,
		WSHub:      wsHub,
		NotiBus:    notiBus,
	})

	err = queue.RegisterAll(rqManager, queue.Config{
		QueueName: cfg.AMQP.QueueName,
		NotiBus:   notiBus,
		Log:       log,
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
		log.Info(ctx, "HTTP API listening", "addr", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

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
			log.Error(ctx, "could not stop http server gracefully", "error", err)
		}

		if err := wsHub.Stop(ctxShut); err != nil {
			log.Error(ctx, "hub stop error", "error", err)
		}
	}

	return nil
}
