package servergrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hosting-contracts/provisioning-service/queue/events"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type handlers struct {
	serverBus server.ExtBusiness
	log       *logger.Logger
}

func new(serverBus server.ExtBusiness, log *logger.Logger) *handlers {
	return &handlers{
		serverBus: serverBus,
		log:       log,
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

	h.log.Info(ctx, "provisioning succeeded",
		"server_id", event.ServerID,
		"ip_address", event.IPv4Address,
	)

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

	h.log.Info(ctx, "provisioning failed reported",
		"server_id", event.ServerID,
		"reason", event.Reason,
	)

	if err := h.serverBus.SetProvisioningFailed(ctx, event.ServerID); err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return fmt.Errorf("%w: server with ID: '%s' not found", messaging.ErrPermanentFailure, event.ServerID)
		}
		return err
	}

	return nil
}
