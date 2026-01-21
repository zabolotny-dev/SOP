package websocket

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
}

type Hub struct {
	clients map[uuid.UUID]map[*Client]bool

	register   chan *Client
	unregister chan *Client

	sendToUser chan userMessage

	cfg Config

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type userMessage struct {
	userID uuid.UUID
	data   []byte
}

func NewHub(cfg Config) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sendToUser: make(chan userMessage, 256),
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
	}

	h.wg.Add(1)
	go h.run()
	return h
}

func (h *Hub) run() {
	defer h.wg.Done()

	for {
		select {
		case <-h.ctx.Done():
			h.closeAll()
			return

		case client := <-h.register:
			if _, ok := h.clients[client.userID]; !ok {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true

		case client := <-h.unregister:
			if userClients, ok := h.clients[client.userID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					client.close()

					if len(userClients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}

		case msg := <-h.sendToUser:
			if clients, ok := h.clients[msg.userID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.data:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
		}
	}
}

func (h *Hub) Send(ctx context.Context, userID uuid.UUID, data []byte) error {
	select {
	case h.sendToUser <- userMessage{userID: userID, data: data}:
		return nil
	case <-h.ctx.Done():
		return errors.New("hub is shutting down")
	default:
		return errors.New("hub send channel full")
	}
}

func (h *Hub) Stop(ctx context.Context) error {
	h.cancel()

	c := make(chan struct{})
	go func() {
		defer close(c)
		h.wg.Wait()
	}()

	select {
	case <-c:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) closeAll() {
	for _, userClients := range h.clients {
		for client := range userClients {
			client.close()
		}
	}
}

func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}
