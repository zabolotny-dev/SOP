package servergrp

import (
	"context"
)

func (h *handlers) HandleDLQ(ctx context.Context, body []byte, routingKey string) error {
	h.log.Error(ctx, "message dropped to DLQ",
		"routing_key", routingKey,
		"body", string(body),
	)

	return nil
}
