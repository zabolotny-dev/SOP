package servergrp

import (
	"context"
	"fmt"
	"hosting-contracts/hosting-service/queue/events"
	"hosting-contracts/topology"
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-notification-service/internal/notification"
)

type Config struct {
	NotiBus   *notification.Notifier
	QueueName string
	Log       *logger.Logger
}

func Register(manager *messaging.MessageManager, cfg Config) error {
	handlers := new(cfg.Log, cfg.NotiBus)

	wrappedHandler := messaging.LogErrors(
		func(ctx context.Context, err error, key string) {
			cfg.Log.Error(ctx, "message processing failed", "error", err, "routing_key", key)
		}, handlers.HandleServerUpdated)

	err := manager.Subscribe(
		cfg.QueueName,
		events.ServerStatusUpdated,
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
