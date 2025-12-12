package servermsg

import (
	"context"
	"fmt"
	"hosting-events-contract/events"
	"hosting-events-contract/topology"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type Publisher struct {
	publisher *messaging.MessageManager
}

func NewPublisher(publisher *messaging.MessageManager) *Publisher {
	return &Publisher{
		publisher: publisher,
	}
}

func (p *Publisher) RequestIP(ctx context.Context, server server.Server) error {

	command := events.ProvisionServerCommand{
		ServerID: server.ID,
		Hostname: server.Name,
	}

	if err := p.publisher.Publish(topology.CommandsExchange, events.ProvisionRequestKey, command); err != nil {
		return fmt.Errorf("msg: failed to queue server for provisioning")
	}

	return nil
}
