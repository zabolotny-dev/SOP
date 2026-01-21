package events

import "github.com/google/uuid"

const (
	ServerStatusUpdated = "server.updated"
)

type ServerStatusChangedEvent struct {
	OwnerID     uuid.UUID `json:"ownerId"`
	ServerID    uuid.UUID `json:"serverId"`
	Status      string    `json:"status"`
	IPv4Address *string   `json:"ip,omitempty"`
}
