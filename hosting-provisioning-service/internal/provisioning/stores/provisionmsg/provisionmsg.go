package provisionmsg

import (
	"context"
	"fmt"
	"hosting-contracts/provisioning-service/queue/events"
	"hosting-contracts/topology"
	"hosting-kit/messaging"
	"hosting-provisioning-service/internal/provisioning"

	"time"

	"github.com/google/uuid"
)

type Notifier struct {
	mgr *messaging.MessageManager
}

func NewNotifier(mgr *messaging.MessageManager) *Notifier {
	return &Notifier{
		mgr: mgr,
	}
}

func (s *Notifier) NotifySuccess(ctx context.Context, serverID uuid.UUID, res provisioning.Result) error {
	successEvent := events.ServerProvisionedEvent{
		ServerID:      serverID,
		IPv4Address:   res.IP,
		ProvisionedAt: res.ProvisionedAt,
	}

	if err := s.mgr.Publish(ctx, topology.EventsExchange, events.ProvisionSucceededKey, successEvent); err != nil {
		return fmt.Errorf("failed to publish ServerProvisionedEvent: %w", err)
	}
	return nil
}

func (s *Notifier) NotifyFailure(ctx context.Context, serverID uuid.UUID, reason string, failedAt time.Time) error {
	failedEvent := events.ServerProvisionFailedEvent{
		ServerID: serverID,
		Reason:   reason,
		FailedAt: failedAt,
	}

	if err := s.mgr.Publish(ctx, topology.EventsExchange, events.ProvisionFailedKey, failedEvent); err != nil {
		return fmt.Errorf("failed to publish ServerProvisionFailedEvent: %w", err)
	}
	return nil
}
