package queue

import (
	"hosting-kit/messaging"
	"hosting-provisioning-service/cmd/server/queue/handlers/provisegrp"
	"hosting-provisioning-service/internal/provisioning"
)

type Config struct {
	ProvBus   *provisioning.Business
	QueueName string
}

func RegisterAll(manager *messaging.MessageManager, cfg Config) error {
	err := provisegrp.Register(
		manager,
		provisegrp.Config{
			ProvBus:   cfg.ProvBus,
			QueueName: cfg.QueueName,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
