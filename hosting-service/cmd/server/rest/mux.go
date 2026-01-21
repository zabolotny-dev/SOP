package rest

import (
	"fmt"
	"hosting-kit/auth"
	"hosting-kit/logger"
	"hosting-kit/mid"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"hosting-contracts/hosting-service/openapi"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type Config struct {
	PlanBus    plan.ExtBusiness
	ServerBus  server.ExtBusiness
	Prefix     string
	AuthClient auth.Client
	Log        *logger.Logger
}

func RegisterRoutes(router *chi.Mux, cfg Config) {
	apiImpl := New(cfg.PlanBus, cfg.ServerBus, cfg.Prefix)

	strictHandler := gen.NewStrictHandlerWithOptions(apiImpl, nil, gen.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: makeResponseErrorHandler(cfg.Log),
		RequestErrorHandlerFunc:  makeRequestErrorHandler(cfg.Log),
	})

	wrapper := &gen.ServerInterfaceWrapper{
		Handler:          strictHandler,
		ErrorHandlerFunc: makeWrapperErrorHandler(cfg.Log),
	}

	authen := mid.Authenticate(cfg.AuthClient)
	adminOnly := mid.RequireAdmin()

	router.Route(cfg.Prefix, func(r chi.Router) {
		specURL := fmt.Sprintf("%s/swagger/doc.yaml", cfg.Prefix)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(specURL),
		))
		r.Get("/swagger/doc.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write(openapi.OpenApiSpec)
		})

		r.Get("/", wrapper.GetRoot)
		r.Get("/plans", wrapper.ListPlans)
		r.Get("/plans/{planId}", wrapper.GetPlanById)

		r.Group(func(r chi.Router) {
			r.Use(authen)

			r.Get("/servers", wrapper.ListServers)
			r.Post("/servers", wrapper.OrderServer)
			r.Get("/servers/{serverId}", wrapper.GetServerById)
			r.Post("/servers/{serverId}/actions", wrapper.PerformServerAction)

			r.Group(func(r chi.Router) {
				r.Use(adminOnly)
				r.Post("/plans", wrapper.CreatePlan)
			})
		})
	})
}
