package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type HandlerConfig struct {
	PlanBus   *plan.Business
	ServerBus *server.Business
}

func NewHandler(cfg HandlerConfig) http.Handler {
	router := chi.NewRouter()

	resolver := &Resolver{
		PlanBus:   cfg.PlanBus,
		ServerBus: cfg.ServerBus,
	}

	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	router.Handle("/", srv)
	router.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	return router
}
