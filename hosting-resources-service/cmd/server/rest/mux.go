package rest

import (
	"fmt"
	"hosting-contracts/resources-service/openapi"
	"hosting-resources-service/cmd/server/rest/gen"
	"hosting-resources-service/internal/pool"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Config struct {
	PoolBus pool.ExtBusiness
	Prefix  string
}

func RegisterRoutes(router *chi.Mux, cfg Config) {
	apiImpl := New(cfg.PoolBus, cfg.Prefix)
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
