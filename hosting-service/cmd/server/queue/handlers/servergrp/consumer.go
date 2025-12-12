package servergrp

import (
	"fmt"
	"hosting-events-contract/events"
	"hosting-events-contract/topology"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type Config struct {
	ServerBus *server.Business
	QueueName string
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.ServerBus)

	err := manager.Subscribe(
		topology.GetDLQQueueName(cfg.QueueName),
		topology.GetDLQKey(cfg.QueueName),
		topology.DLXExchange,
		handlers.HandleDLQ,
		nil,
	)

	if err != nil {
		return fmt.Errorf("servergrp: subscribe dlq failed: %w", err)
	}

	err = manager.Subscribe(cfg.QueueName,
		events.ProvisionResultKeyPattern,
		topology.EventsExchange,
		handlers.HandleProvisionResult,
		&messaging.DLQConfig{
			ExchangeName: topology.DLXExchange,
			RoutingKey:   topology.GetDLQKey(cfg.QueueName),
		},
	)

	if err != nil {
		return fmt.Errorf("servergrp: subscribe provision failed: %w", err)
	}

	return nil
}
