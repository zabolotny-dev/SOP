package servermsg

import (
	"context"
	"hosting-contracts/hosting-service/queue/events"
	"hosting-contracts/topology"
	"hosting-kit/messaging"
	"hosting-service/internal/server"
)

type Notifier struct {
	publisher *messaging.MessageManager
}

func NewNotifier(publisher *messaging.MessageManager) *Notifier {
	return &Notifier{
		publisher: publisher,
	}
}

func (p *Notifier) ServerUpdated(ctx context.Context, server server.Server) {
	event := events.ServerStatusChangedEvent{
		ServerID:    server.ID,
		OwnerID:     server.OwnerID,
		Status:      string(server.Status),
		IPv4Address: server.IPv4Address,
	}

	p.publisher.Publish(ctx, topology.EventsExchange, events.ServerStatusUpdated, event)
}
