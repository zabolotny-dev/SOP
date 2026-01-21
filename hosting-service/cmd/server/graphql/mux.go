package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"

	"hosting-kit/auth"
	"hosting-kit/mid"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type HandlerConfig struct {
	PlanBus    plan.ExtBusiness
	ServerBus  server.ExtBusiness
	AuthClient auth.Client
	Prefix     string
}

func RegisterRoutes(router *chi.Mux, cfg HandlerConfig) {
	resolver := &Resolver{
		PlanBus:   cfg.PlanBus,
		ServerBus: cfg.ServerBus,
	}
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	url := cfg.Prefix + "/graphql"

	router.Route(url, func(r chi.Router) {
		r.With(mid.AuthenticateOptional(cfg.AuthClient)).Handle("/", srv)
		r.Handle("/playground", playground.Handler("GraphQL Playground", url))
	})
}
