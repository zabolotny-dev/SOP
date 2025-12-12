package server

import (
	"context"
	"errors"
	"fmt"
	"hosting-service/internal/plan"
	"hosting-service/internal/platform/page"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrServerNotFound = errors.New("server not found")
	ErrValidation     = errors.New("validation error")
)

type Storer interface {
	FindByID(ctx context.Context, ID uuid.UUID) (Server, error)
	Create(ctx context.Context, server Server) error
	Update(ctx context.Context, server Server) error
	Delete(ctx context.Context, ID uuid.UUID) error
	FindAll(ctx context.Context, pg page.Page) ([]Server, int, error)
}

type Provisioner interface {
	RequestIP(ctx context.Context, server Server) error
}

type Business struct {
	storer      Storer
	planBus     *plan.Business
	provisioner Provisioner
}

func NewBusiness(storer Storer, planBus *plan.Business, provisioner Provisioner) *Business {
	return &Business{
		storer:      storer,
		planBus:     planBus,
		provisioner: provisioner,
	}
}

func NewServer(planID uuid.UUID, name string) (Server, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return Server{}, fmt.Errorf("%w: server name cannot be empty", ErrValidation)
	}
	if planID == uuid.Nil {
		return Server{}, fmt.Errorf("%w: planID cannot be nil", ErrValidation)
	}

	return Server{
		ID:        uuid.New(),
		PlanID:    planID,
		Name:      trimmedName,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Business) FindByID(ctx context.Context, ID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, ID)
	if err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) Create(ctx context.Context, name string, planID uuid.UUID) (Server, error) {
	_, err := s.planBus.FindByID(ctx, planID)
	if err != nil {
		return Server{}, err
	}

	server, err := NewServer(planID, name)
	if err != nil {
		return Server{}, err
	}

	err = s.storer.Create(ctx, server)
	if err != nil {
		return Server{}, err
	}

	if err := s.provisioner.RequestIP(ctx, server); err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) Search(ctx context.Context, pg page.Page) ([]Server, int, error) {
	servers, count, err := s.storer.FindAll(ctx, pg)
	if err != nil {
		return nil, 0, err
	}

	return servers, count, nil
}

func (s *Business) Start(ctx context.Context, serverID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, err
	}

	if server.Status != StatusStopped {
		return Server{}, fmt.Errorf("%w: cannot start server with status '%s', expected STOPPED", ErrValidation, server.Status)
	}

	server.Status = StatusRunning

	if err := s.storer.Update(ctx, server); err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) Stop(ctx context.Context, serverID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, err
	}

	if server.Status != StatusRunning {
		return Server{}, fmt.Errorf("%w: cannot stop server with status '%s', expected RUNNING", ErrValidation, server.Status)
	}

	server.Status = StatusStopped

	if err := s.storer.Update(ctx, server); err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) Delete(ctx context.Context, serverID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, err
	}

	if server.Status != StatusStopped && server.Status != StatusRunning {
		return Server{}, fmt.Errorf("%w: cannot delete server with status '%s', expected RUNNING or STOPPED", ErrValidation, server.Status)
	}

	if err := s.storer.Delete(ctx, serverID); err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) SetIPAddress(ctx context.Context, serverID uuid.UUID, ip string) error {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return err
	}

	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%w: invalid ip address format: %s", ErrValidation, ip)
	}

	if server.IPv4Address != nil && *server.IPv4Address == ip {
		return nil
	}

	if server.Status != StatusPending {
		return fmt.Errorf("%w: cannot set IP to this server with status '%s', expected PENDING", ErrValidation, server.Status)
	}

	server.Status = StatusStopped
	server.IPv4Address = &ip

	if err := s.storer.Update(ctx, server); err != nil {
		return err
	}

	return nil
}

func (s *Business) SetProvisioningFailed(ctx context.Context, serverID uuid.UUID) error {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return err
	}

	if server.Status == StatusProvisionFailed {
		return nil
	}

	server.Status = StatusProvisionFailed

	if err := s.storer.Update(ctx, server); err != nil {
		return err
	}

	return nil
}
