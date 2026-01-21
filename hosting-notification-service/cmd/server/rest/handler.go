package rest

import (
	"hosting-kit/auth"
	"hosting-notification-service/internal/notification"
	"hosting-notification-service/internal/platform/websocket"
	"net/http"
)

type handler struct {
	notiBus *notification.Notifier
	wsHub   *websocket.Hub
}

func (h *handler) wsConnect(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := websocket.NewClient(h.wsHub, claims.UserID, conn)

	h.wsHub.RegisterClient(client)

	go client.Serve()
}
