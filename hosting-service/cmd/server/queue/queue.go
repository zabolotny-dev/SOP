package queue

import (
	"hosting-kit/messaging"
	"hosting-service/cmd/server/queue/handlers/servergrp"
	"hosting-service/internal/server"
)

type Config struct {
	ServerBus *server.Business
	QueueName string
}

func RegisterAll(manager *messaging.MessageManager, cfg Config) error {
	err := servergrp.Register(
		manager,
		servergrp.Config{
			ServerBus: cfg.ServerBus,
			QueueName: cfg.QueueName,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
