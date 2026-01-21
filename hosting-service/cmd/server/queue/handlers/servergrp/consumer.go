package servergrp

import (
	"context"
	"fmt"
	"hosting-contracts/provisioning-service/queue/events"
	"hosting-contracts/topology"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type Config struct {
	ServerBus server.ExtBusiness
	QueueName string
	Log       *logger.Logger
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.ServerBus, cfg.Log)

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

	wrappedHandler := messaging.LogErrors(func(ctx context.Context, err error, key string) {
		cfg.Log.Error(ctx, "message processing failed", "error", err, "routing_key", key)
	}, handlers.HandleProvisionResult)

	err = manager.Subscribe(cfg.QueueName,
		events.ProvisionResultKeyPattern,
		topology.EventsExchange,
		wrappedHandler,
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
