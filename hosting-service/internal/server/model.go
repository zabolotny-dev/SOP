package server

import (
	"time"

	"github.com/google/uuid"
)

type ServerStatus string

type ActionType string

const (
	StatusPending         ServerStatus = "PENDING"
	StatusRunning         ServerStatus = "RUNNING"
	StatusStopped         ServerStatus = "STOPPED"
	StatusProvisionFailed ServerStatus = "PROVISION_FAILED"
)

const (
	ActionStart  ActionType = "START"
	ActionStop   ActionType = "STOP"
	ActionDelete ActionType = "DELETE"
)

type Server struct {
	ID          uuid.UUID
	IPv4Address *string
	PlanID      uuid.UUID
	Name        string
	Status      ServerStatus
	CreatedAt   time.Time
}
