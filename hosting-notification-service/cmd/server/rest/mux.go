package rest

import (
	"hosting-kit/auth"
	"hosting-kit/mid"
	"hosting-notification-service/internal/notification"
	"hosting-notification-service/internal/platform/websocket"
	"net/http"

	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Config struct {
	NotiBus    *notification.Notifier
	AuthClient auth.Client
	WSHub      *websocket.Hub
	Prefix     string
}

func RegisterRoutes(router *chi.Mux, cfg Config) {
	h := handler{
		notiBus: cfg.NotiBus,
		wsHub:   cfg.WSHub,
	}

	router.Route(cfg.Prefix, func(r chi.Router) {
		r.Use(mid.Authenticate(cfg.AuthClient))

		r.Get("/ws", h.wsConnect)
	})
}
