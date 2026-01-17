package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type HandlerConfig struct {
	PlanBus   plan.ExtBusiness
	ServerBus server.ExtBusiness
}

func RegisterRoutes(router *chi.Mux, cfg HandlerConfig) {
	resolver := &Resolver{
		PlanBus:   cfg.PlanBus,
		ServerBus: cfg.ServerBus,
	}
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	router.Route("/graphql", func(r chi.Router) {
		r.Handle("/", srv)
		r.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))
	})
}
