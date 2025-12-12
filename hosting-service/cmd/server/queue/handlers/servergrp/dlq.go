package servergrp

import (
	"context"
	"log"
)

func (h *handlers) HandleDLQ(ctx context.Context, body []byte, routingKey string) error {
	log.Printf("Message dropped! RoutingKey: %s", routingKey)
	log.Printf("Body: %s", string(body))

	return nil
}
