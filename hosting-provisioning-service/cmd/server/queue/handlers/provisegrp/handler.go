package provisegrp

import (
	"context"
	"encoding/json"
	"fmt"
	"hosting-contracts/hosting-service/queue/commands"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"

	"log"
)

type handlers struct {
	provBus provisioning.ExtBusiness
}

func new(provBus provisioning.ExtBusiness) *handlers {
	return &handlers{
		provBus: provBus,
	}
}

func (h *handlers) handleProvisionServer(ctx context.Context, body []byte, routingKey string) error {
	if routingKey != commands.ProvisionRequestKey {
		return fmt.Errorf("%w: unknown routing key: %s", messaging.ErrPermanentFailure, routingKey)
	}

	var cmd commands.ProvisionServerCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		log.Printf("ERROR: failed to unmarshal command: %v. Message will be dropped.", err)
		return fmt.Errorf("%w: failed to unmarshal ServerProvisionedEvent: %v", messaging.ErrPermanentFailure, err)
	}

	log.Printf("Received provisioning request for server %s (%s)", cmd.Hostname, cmd.ServerID)
	return h.provBus.GenerateIP(ctx, cmd.ServerID)
}
