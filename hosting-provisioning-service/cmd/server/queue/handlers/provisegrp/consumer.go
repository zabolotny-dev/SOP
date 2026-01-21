package provisegrp

import (
	"context"
	"fmt"
	"hosting-contracts/hosting-service/queue/commands"
	"hosting-contracts/topology"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"
)

type Config struct {
	ProvBus   provisioning.ExtBusiness
	QueueName string
	Log       *logger.Logger
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.Log, cfg.ProvBus)

	wrappedHandler := messaging.LogErrors(func(ctx context.Context, err error, key string) {
		cfg.Log.Error(ctx, "message processing failed", "error", err, "routing_key", key)
	}, handlers.handleProvisionServer)

	err := manager.Subscribe(
		cfg.QueueName,
		commands.ProvisionRequestKey,
		topology.CommandsExchange,
		wrappedHandler,
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
