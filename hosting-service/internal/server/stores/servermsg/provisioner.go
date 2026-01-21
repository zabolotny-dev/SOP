package servermsg

import (
	"context"
	"fmt"
	"hosting-contracts/hosting-service/queue/commands"
	"hosting-contracts/topology"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type Provisioner struct {
	publisher *messaging.MessageManager
}

func NewProvisioner(publisher *messaging.MessageManager) *Provisioner {
	return &Provisioner{
		publisher: publisher,
	}
}

func (p *Provisioner) RequestIP(ctx context.Context, server server.Server) error {
	command := commands.ProvisionServerCommand{
		ServerID: server.ID,
		Hostname: server.Name,
	}

	if err := p.publisher.Publish(ctx, topology.CommandsExchange, commands.ProvisionRequestKey, command); err != nil {
		return fmt.Errorf("msg: failed to queue server for provisioning: %w", err)
	}

	return nil
}
