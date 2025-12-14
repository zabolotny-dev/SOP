package main

import (
	"context"
	"errors"
	"fmt"
	"hosting-events-contract/topology"
	"hosting-kit/messaging"
	graph "hosting-service/cmd/server/graphql"
	"hosting-service/cmd/server/queue"
	"hosting-service/cmd/server/rest"
	"hosting-service/internal/plan"
	"hosting-service/internal/plan/stores/plandb"
	"hosting-service/internal/platform/database"
	"hosting-service/internal/platform/mid"
	"hosting-service/internal/server"
	"hosting-service/internal/server/stores/serverdb"
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
			APIHost      string        `conf:"default:0.0.0.0:8080,env:HTTP_PORT"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
		}
		AMQP struct {
			URL            string        `conf:"default:amqp://guest:guest@localhost:5672/,mask,env:AMQP_URL"`
			HandlerTimeout time.Duration `conf:"default:10s"`
			QueueName      string        `conf:"default:api_events_queue"`
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
		return fmt.Errorf("initializing rabbitmq: %w", err)
	}

	defer func() {
		ctxShut, cancel := context.WithTimeout(ctx, cfg.App.ShutdownTimeout)
		defer cancel()
		if err := rqManager.Stop(ctxShut); err != nil {
			log.Printf("Failed to shutdown rabbit manager: %v", err)
		}
	}()

	planStore := plandb.NewStore(db)
	planBus := plan.NewBusiness(planStore)

	serverPublisher := servermsg.NewPublisher(rqManager)
	serverStore := serverdb.NewStore(db)
	serverBus := server.NewBusiness(serverStore, planBus, serverPublisher)

	err = queue.RegisterAll(rqManager, queue.Config{
		ServerBus: serverBus,
		QueueName: cfg.AMQP.QueueName,
	})
	if err != nil {
		return fmt.Errorf("registering queue handlers: %w", err)
	}

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(mid.Logger)
	mux.Use(mid.Performance)

	restHandler := rest.NewHandler(rest.Config{
		PlanBus:   planBus,
		ServerBus: serverBus,
	})
	mux.Mount("/api", restHandler)

	graphHandler := graph.NewHandler(graph.HandlerConfig{
		PlanBus:   planBus,
		ServerBus: serverBus,
	})
	mux.Mount("/graphql", graphHandler)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      mux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("main: HTTP API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

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
