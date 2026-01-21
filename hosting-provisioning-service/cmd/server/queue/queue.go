package queue

import (
	"hosting-kit/logger"
	"hosting-kit/messaging"
	"hosting-provisioning-service/cmd/server/queue/handlers/provisegrp"
	"hosting-provisioning-service/internal/provisioning"
)

type Config struct {
	ProvBus   provisioning.ExtBusiness
	QueueName string
	Log       *logger.Logger
}

func RegisterAll(manager *messaging.MessageManager, cfg Config) error {
	err := provisegrp.Register(
		manager,
		provisegrp.Config{
			ProvBus:   cfg.ProvBus,
			QueueName: cfg.QueueName,
			Log:       cfg.Log,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
