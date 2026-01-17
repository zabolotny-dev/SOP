package rest

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"hosting-contracts/hosting-service/openapi"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type Config struct {
	PlanBus   plan.ExtBusiness
	ServerBus server.ExtBusiness
	Prefix    string
}

func RegisterRoutes(router *chi.Mux, cfg Config) {
	apiImpl := New(cfg.PlanBus, cfg.ServerBus, cfg.Prefix)
	strictHandler := gen.NewStrictHandler(apiImpl, nil)

	router.Route(cfg.Prefix, func(r chi.Router) {
		specURL := fmt.Sprintf("%s/swagger/doc.yaml", cfg.Prefix)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(specURL),
		))
		r.Get("/swagger/doc.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write(openapi.OpenApiSpec)
		})

		gen.HandlerFromMux(strictHandler, r)
	})
}
