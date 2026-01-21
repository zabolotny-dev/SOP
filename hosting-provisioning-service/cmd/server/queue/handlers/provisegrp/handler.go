package provisegrp

import (
	"context"
	"encoding/json"
	"fmt"
	"hosting-contracts/hosting-service/queue/commands"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"
)

type handlers struct {
	log     *logger.Logger
	provBus provisioning.ExtBusiness
}

func new(log *logger.Logger, provBus provisioning.ExtBusiness) *handlers {
	return &handlers{
		log:     log,
		provBus: provBus,
	}
}

func (h *handlers) handleProvisionServer(ctx context.Context, body []byte, routingKey string) error {
	if routingKey != commands.ProvisionRequestKey {
		return fmt.Errorf("%w: unknown routing key: %s", messaging.ErrPermanentFailure, routingKey)
	}

	var cmd commands.ProvisionServerCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		return fmt.Errorf("%w: failed to unmarshal command: %v", messaging.ErrPermanentFailure, err)
	}

	h.log.Info(ctx, "received provisioning request",
		"hostname", cmd.Hostname,
		"server_id", cmd.ServerID,
	)

	return h.provBus.GenerateIP(ctx, cmd.ServerID)
}
