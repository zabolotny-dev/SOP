package provisegrp

import (
	"context"
	"encoding/json"
	"fmt"
	"hosting-events-contract/events"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"
	"log"
)

type handlers struct {
	provBus *provisioning.Business
}

func new(provBus *provisioning.Business) *handlers {
	return &handlers{
		provBus: provBus,
	}
}

func (h *handlers) handleProvisionServer(ctx context.Context, body []byte, routingKey string) error {
	if routingKey != events.ProvisionRequestKey {
		return fmt.Errorf("%w: unknown routing key: %s", messaging.ErrPermanentFailure, routingKey)
	}

	var cmd events.ProvisionServerCommand
	if err := json.Unmarshal(body, &cmd); err != nil {
		log.Printf("ERROR: failed to unmarshal command: %v. Message will be dropped.", err)
		return fmt.Errorf("%w: failed to unmarshal ServerProvisionedEvent: %v", messaging.ErrPermanentFailure, err)
	}

	log.Printf("Received provisioning request for server %s (%s)", cmd.Hostname, cmd.ServerID)
	return h.provBus.GenerateIP(ctx, cmd.ServerID)
}
