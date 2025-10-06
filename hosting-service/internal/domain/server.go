package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ServerStatus string

const (
	StatusPending   ServerStatus = "PENDING"
	StatusRunning   ServerStatus = "RUNNING"
	StatusStopped   ServerStatus = "STOPPED"
	StatusRebooting ServerStatus = "REBOOTING"
	StatusDeleting  ServerStatus = "DELETING"
)

type Server struct {
	BaseModel
	//UserID    uuid.UUID    `gorm:"type:uuid;not null"`
	PlanID    uuid.UUID    `gorm:"type:uuid;not null"`
	Name      string       `gorm:"type:varchar(255);not null"`
	Status    ServerStatus `gorm:"type:varchar(32);not null"`
	CreatedAt time.Time    `gorm:"type:timestamptz;not null"`
}

func NewServer(planID uuid.UUID, name string) (*Server, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("%w: server name cannot be empty", ErrValidation)
	}
	//if userID == uuid.Nil {
	//	return nil, errors.New("userID cannot be nil")
	//}
	if planID == uuid.Nil {
		return nil, fmt.Errorf("%w: planID cannot be nil", ErrValidation)
	}

	return &Server{
		BaseModel: NewBaseModel(),
		//UserID:    userID,
		PlanID:    planID,
		Name:      trimmedName,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Server) Provisioned() error {
	if s.Status != StatusPending {
		return fmt.Errorf("%w: cannot provision server with status '%s', expected PENDING", ErrValidation, s.Status)
	}
	s.Status = StatusStopped
	return nil
}

func (s *Server) Start() error {
	if s.Status != StatusStopped {
		return fmt.Errorf("%w: cannot start server with status '%s', expected STOPPED", ErrValidation, s.Status)
	}
	s.Status = StatusRunning
	return nil
}

func (s *Server) Stop() error {
	if s.Status != StatusRunning {
		return fmt.Errorf("%w: cannot stop server with status '%s', expected RUNNING", ErrValidation, s.Status)
	}
	s.Status = StatusStopped
	return nil
}

func (s *Server) Reboot() error {
	if s.Status != StatusRunning {
		return fmt.Errorf("%w: cannot reboot server with status '%s', expected RUNNING", ErrValidation, s.Status)
	}
	s.Status = StatusRebooting
	return nil
}

func (s *Server) MarkForDeletion() error {
	if s.Status != StatusRunning && s.Status != StatusStopped {
		return fmt.Errorf("%w: cannot delete server with status '%s', expected RUNNING or STOPPED", ErrValidation, s.Status)
	}
	s.Status = StatusDeleting
	return nil
}
