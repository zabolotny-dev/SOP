package queue

import (
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-notification-service/cmd/server/queue/handlers/servergrp"
	"hosting-notification-service/internal/notification"
)

type Config struct {
	QueueName string
	NotiBus   *notification.Notifier
	Log       *logger.Logger
}

func RegisterAll(manager *messaging.MessageManager, cfg Config) error {
	err := servergrp.Register(
		manager,
		servergrp.Config{
			NotiBus:   cfg.NotiBus,
			QueueName: cfg.QueueName,
			Log:       cfg.Log,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
