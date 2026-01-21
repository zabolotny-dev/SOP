package queue

import (
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-service/cmd/server/queue/handlers/servergrp"
	"hosting-service/internal/server"
)

type Config struct {
	ServerBus server.ExtBusiness
	QueueName string
	Log       *logger.Logger
}

func RegisterAll(manager *messaging.MessageManager, cfg Config) error {
	err := servergrp.Register(
		manager,
		servergrp.Config{
			ServerBus: cfg.ServerBus,
			QueueName: cfg.QueueName,
			Log:       cfg.Log,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
