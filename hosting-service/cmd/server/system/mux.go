package system

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(router *chi.Mux) {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("TEST OK"))
	})

	router.Method(http.MethodGet, "/metrics", promhttp.Handler())
}
