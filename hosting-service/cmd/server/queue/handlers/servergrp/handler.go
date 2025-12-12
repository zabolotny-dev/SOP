package servergrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hosting-events-contract/events"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type handlers struct {
	serverBus *server.Business
}

func new(serverBus *server.Business) *handlers {
	return &handlers{
		serverBus: serverBus,
	}
}

func (h *handlers) HandleProvisionResult(ctx context.Context, body []byte, routingKey string) error {
	switch routingKey {
	case events.ProvisionSucceededKey:
		return h.handleSuccessProvision(ctx, body)
	case events.ProvisionFailedKey:
		return h.handleFailureProvision(ctx, body)
	default:
		return fmt.Errorf("%w: unknown routing key: %s", messaging.ErrPermanentFailure, routingKey)
	}
}

func (h *handlers) handleSuccessProvision(ctx context.Context, body []byte) error {
	var event events.ServerProvisionedEvent

	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("%w: unmarshal ServerProvisionedEvent failed: %v", messaging.ErrPermanentFailure, err)
	}

	if err := h.serverBus.SetIPAddress(ctx, event.ServerID, event.IPv4Address); err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return fmt.Errorf("%w: server with ID: '%s' not found", messaging.ErrPermanentFailure, event.ServerID)
		}
		if errors.Is(err, server.ErrValidation) {
			return fmt.Errorf("%w: server with ID: '%s' has validation errors: %v", messaging.ErrPermanentFailure, event.ServerID, err)
		}
		return err
	}
	return nil
}

func (h *handlers) handleFailureProvision(ctx context.Context, body []byte) error {
	var event events.ServerProvisionFailedEvent

	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("%w: unmarshal ServerProvisionFailedEvent failed: %v", messaging.ErrPermanentFailure, err)
	}

	if err := h.serverBus.SetProvisioningFailed(ctx, event.ServerID); err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return fmt.Errorf("%w: server with ID: '%s' not found", messaging.ErrPermanentFailure, event.ServerID)
		}
		return err
	}

	return nil
}
