package provisegrp

import (
	"fmt"
	"hosting-contracts/hosting-service/queue/commands"
	"hosting-contracts/topology"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"
)

type Config struct {
	ProvBus   provisioning.ExtBusiness
	QueueName string
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.ProvBus)

	err := manager.Subscribe(
		cfg.QueueName,
		commands.ProvisionRequestKey,
		topology.CommandsExchange,
		handlers.handleProvisionServer,
		&messaging.DLQConfig{
			ExchangeName: topology.DLXExchange,
			RoutingKey:   topology.GetDLQKey(cfg.QueueName),
		},
	)

	if err != nil {
		return fmt.Errorf("provisegrp: subscribe provision failed: %w", err)
	}

	return nil
}
