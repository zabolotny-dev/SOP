package servergrp

import (
	"context"
	"encoding/json"
	"fmt"
	"hosting-contracts/hosting-service/queue/events"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-notification-service/internal/notification"
)

type handlers struct {
	log      *logger.Logger
	notifier *notification.Notifier
}

func new(log *logger.Logger, notifier *notification.Notifier) *handlers {
	return &handlers{notifier: notifier, log: log}
}

func (h *handlers) HandleServerUpdated(ctx context.Context, body []byte, routingKey string) error {
	var serverEvent events.ServerStatusChangedEvent
	if err := json.Unmarshal(body, &serverEvent); err != nil {
		return fmt.Errorf("%w: unmarshal failed: %v", messaging.ErrPermanentFailure, err)
	}

	h.log.Info(ctx, "handling server update",
		"owner_id", serverEvent.OwnerID,
		"server_id", serverEvent.ServerID,
		"status", serverEvent.Status,
		"routing_key", routingKey,
	)

	event := notification.Event{
		Type:    routingKey,
		Payload: json.RawMessage(body),
	}

	if err := h.notifier.Notify(ctx, serverEvent.OwnerID, event); err != nil {
		return err
	}

	return nil
}
