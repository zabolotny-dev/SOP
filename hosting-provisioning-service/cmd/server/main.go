package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-contracts/topology"
	"hosting-kit/debug"
	"hosting-kit/messaging"
	"hosting-kit/otel"
	"hosting-provisioning-service/cmd/server/queue"
	"hosting-provisioning-service/internal/provisioning"
	"hosting-provisioning-service/internal/provisioning/extensions/provisioningotel"
	"hosting-provisioning-service/internal/provisioning/stores/provisionmsg"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"
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
		AMQP struct {
			URL            string        `conf:"default:amqp://guest:guest@localhost:5672/,mask,env:AMQP_URL"`
			HandlerTimeout time.Duration `conf:"default:10s"`
			QueueName      string        `conf:"default:provisioning_queue"`
		}
		App struct {
			ProvisioningTime time.Duration `conf:"default:10s"`
			ShutdownTimeout  time.Duration `conf:"default:20s"`
		}
		Web struct {
			DebugHost string `conf:"default:0.0.0.0:7010"`
		}
		Tempo struct {
			Host        string  `conf:"default:hosting-tempo:4317"`
			ServiceName string  `conf:"default:provisioning-service"`
			Probability float64 `conf:"default:0.05"`
		}
	}{}

	const prefix = "PROV"

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
		return fmt.Errorf("creating rabbit manager: %w", err)
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

	provisioningOtelExt := provisioningotel.NewExtension()
	provisioningPublisher := provisionmsg.NewNotifier(rqManager)
	provisioningBus := provisioning.NewBusiness(cfg.App.ProvisioningTime, provisioningPublisher, provisioningOtelExt)

	// -------------------------------------------------------------------------
	// Start API Service

	err = queue.RegisterAll(rqManager, queue.Config{
		ProvBus:   provisioningBus,
		QueueName: cfg.AMQP.QueueName,
	})
	if err != nil {
		return fmt.Errorf("registering queue handlers: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start Debug Service

	serverErrors := make(chan error, 1)

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
	case <-shutdown:
		log.Println("main: shutdown signal received")
	}

	return nil
}
