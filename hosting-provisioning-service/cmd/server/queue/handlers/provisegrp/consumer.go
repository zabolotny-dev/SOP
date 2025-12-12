package provisegrp

import (
	"fmt"
	"hosting-events-contract/events"
	"hosting-events-contract/topology"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"
)

type Config struct {
	ProvBus   *provisioning.Business
	QueueName string
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.ProvBus)

	err := manager.Subscribe(
		cfg.QueueName,
		events.ProvisionRequestKey,
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
