package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub    *Hub
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte

	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

func NewClient(hub *Hub, userID uuid.UUID, conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		hub:    hub,
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.cancel()
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.hub.cfg.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.hub.cfg.PongWait))

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.hub.cfg.PongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.cfg.PingInterval)
	defer func() {
		c.cancel()
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.cfg.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.cfg.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.send)
	})
}

func (c *Client) Serve() {
	go c.writePump()
	c.readPump()
}
