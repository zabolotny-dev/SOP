package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-events-contract/topology"
	"hosting-kit/debug"
	"hosting-kit/messaging"
	"hosting-provisioning-service/cmd/server/queue"
	"hosting-provisioning-service/internal/provisioning"
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
			DebugHost string `conf:"default:0.0.0.0:7010,env:HTTP_PORT"`
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
	}, cfg.AMQP.HandlerTimeout)
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

	notifier := provisionmsg.NewNotifier(rqManager)
	provisioningBus := provisioning.NewBusiness(cfg.App.ProvisioningTime, notifier)

	err = queue.RegisterAll(rqManager, queue.Config{
		ProvBus:   provisioningBus,
		QueueName: cfg.AMQP.QueueName,
	})
	if err != nil {
		return fmt.Errorf("registering queue handlers: %w", err)
	}

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Println("Debug server error")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	log.Println("Shutdown signal received, exiting...")

	return nil
}
