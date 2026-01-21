package server

import (
	"context"
	"errors"
	"fmt"
	"hosting-kit/page"
	"hosting-service/internal/plan"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrServerNotFound = errors.New("server not found")
	ErrValidation     = errors.New("validation error")
	ErrInvalidPlan    = errors.New("invalid plan provided")
	ErrNoResources    = errors.New("not enough resources available")
	ErrAccessDenied   = errors.New("access denied")
)

type Extension func(ExtBusiness) ExtBusiness

type Notifier interface {
	ServerUpdated(ctx context.Context, server Server)
}

type ResourcesManager interface {
	Consume(ctx context.Context, r Resources) (uuid.UUID, error)
	Return(ctx context.Context, r Resources, poolID uuid.UUID) error
}

type PlanFinder interface {
	FindByID(ctx context.Context, ID uuid.UUID) (plan.Plan, error)
}

type Storer interface {
	FindByID(ctx context.Context, ID uuid.UUID) (Server, error)
	Create(ctx context.Context, server Server) error
	Update(ctx context.Context, server Server) error
	Delete(ctx context.Context, ID uuid.UUID) error
	FindAll(ctx context.Context, pg page.Page, userID uuid.UUID) ([]Server, int, error)
}

type ExtBusiness interface {
	FindByID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (Server, error)
	Create(ctx context.Context, name string, planID uuid.UUID, userID uuid.UUID) (Server, error)
	Search(ctx context.Context, pg page.Page, userID uuid.UUID) ([]Server, int, error)
	Start(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error)
	Stop(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error)
	Delete(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error)
	SetIPAddress(ctx context.Context, serverID uuid.UUID, ip string) error
	SetProvisioningFailed(ctx context.Context, serverID uuid.UUID) error
}

type Provisioner interface {
	RequestIP(ctx context.Context, server Server) error
}

type Business struct {
	storer      Storer
	planBus     PlanFinder
	provisioner Provisioner
	resources   ResourcesManager
	notifier    Notifier
	extensions  []Extension
}

func NewBusiness(storer Storer, planBus PlanFinder, provisioner Provisioner,
	resources ResourcesManager, notifier Notifier, extensions ...Extension) ExtBusiness {
	b := &Business{
		storer:      storer,
		planBus:     planBus,
		provisioner: provisioner,
		resources:   resources,
		notifier:    notifier,
		extensions:  extensions,
	}

	extBus := ExtBusiness(b)

	for i := len(extensions) - 1; i >= 0; i-- {
		ext := extensions[i]
		if ext != nil {
			extBus = ext(extBus)
		}
	}

	return extBus
}

func NewServer(planID uuid.UUID, poolID uuid.UUID, userID uuid.UUID, name string) (Server, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return Server{}, fmt.Errorf("%w: server name cannot be empty", ErrValidation)
	}
	if planID == uuid.Nil {
		return Server{}, fmt.Errorf("%w: planID cannot be nil", ErrValidation)
	}
	if poolID == uuid.Nil {
		return Server{}, fmt.Errorf("%w: poolID cannot be nil", ErrValidation)
	}
	if userID == uuid.Nil {
		return Server{}, fmt.Errorf("%w: ownerID cannot be nil", ErrValidation)
	}

	return Server{
		ID:        uuid.New(),
		OwnerID:   userID,
		PlanID:    planID,
		PoolID:    poolID,
		Name:      trimmedName,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Business) FindByID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, ID)
	if err != nil {
		return Server{}, fmt.Errorf("findbyid: %w", err)
	}

	if err := checkOwnership(server, userID); err != nil {
		return Server{}, err
	}

	return server, nil
}

func (s *Business) Create(ctx context.Context, name string, planID uuid.UUID, userID uuid.UUID) (Server, error) {
	planFound, err := s.planBus.FindByID(ctx, planID)
	if err != nil {
		if errors.Is(err, plan.ErrPlanNotFound) {
			return Server{}, ErrInvalidPlan
		}
		return Server{}, fmt.Errorf("plan.findbyid: %w", err)
	}

	resorce := Resources{
		CPUCores: planFound.CPUCores,
		RAMMB:    planFound.RAMMB,
		DiskGB:   planFound.DiskGB,
		IPCount:  planFound.IpCount,
	}

	poolID, err := s.resources.Consume(ctx, resorce)
	if err != nil {
		return Server{}, fmt.Errorf("resources.consume: %w", err)
	}

	server, err := NewServer(planID, poolID, userID, name)
	if err != nil {
		return Server{}, err
	}

	err = s.storer.Create(ctx, server)
	if err != nil {
		return Server{}, fmt.Errorf("create: %w", err)
	}

	if err := s.provisioner.RequestIP(ctx, server); err != nil {
		return Server{}, fmt.Errorf("provisioner.requestip: %w", err)
	}

	return server, nil
}

func (s *Business) Search(ctx context.Context, pg page.Page, userID uuid.UUID) ([]Server, int, error) {
	servers, count, err := s.storer.FindAll(ctx, pg, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}

	return servers, count, nil
}

func (s *Business) Start(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, fmt.Errorf("start: %w", err)
	}

	if err := checkOwnership(server, userID); err != nil {
		return Server{}, err
	}

	if server.Status != StatusStopped {
		return Server{}, fmt.Errorf("%w: cannot start server with status '%s', expected STOPPED", ErrValidation, server.Status)
	}

	server.Status = StatusRunning

	if err := s.storer.Update(ctx, server); err != nil {
		return Server{}, fmt.Errorf("start: %w", err)
	}

	return server, nil
}

func (s *Business) Stop(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, fmt.Errorf("stop: %w", err)
	}

	if err := checkOwnership(server, userID); err != nil {
		return Server{}, err
	}

	if server.Status != StatusRunning {
		return Server{}, fmt.Errorf("%w: cannot stop server with status '%s', expected RUNNING", ErrValidation, server.Status)
	}

	server.Status = StatusStopped

	if err := s.storer.Update(ctx, server); err != nil {
		return Server{}, fmt.Errorf("stop: %w", err)
	}

	return server, nil
}

func (s *Business) Delete(ctx context.Context, serverID uuid.UUID, userID uuid.UUID) (Server, error) {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return Server{}, fmt.Errorf("delete: %w", err)
	}

	if err := checkOwnership(server, userID); err != nil {
		return Server{}, err
	}

	if server.Status != StatusStopped && server.Status != StatusRunning && server.Status != StatusProvisionFailed {
		return Server{}, fmt.Errorf("%w: cannot delete server with status '%s', expected RUNNING or STOPPED", ErrValidation, server.Status)
	}

	plan, err := s.planBus.FindByID(ctx, server.PlanID)
	if err != nil {
		return Server{}, fmt.Errorf("delete: %w", err)
	}

	resource := Resources{
		CPUCores: plan.CPUCores,
		RAMMB:    plan.RAMMB,
		DiskGB:   plan.DiskGB,
		IPCount:  plan.IpCount,
	}

	if err := s.storer.Delete(ctx, serverID); err != nil {
		return Server{}, fmt.Errorf("delete: %w", err)
	}

	if err := s.resources.Return(ctx, resource, server.PoolID); err != nil {
		return Server{}, fmt.Errorf("resources.return: %w", err)
	}

	return server, nil
}

func (s *Business) SetIPAddress(ctx context.Context, serverID uuid.UUID, ip string) error {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("setipaddress: %w", err)
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
		return fmt.Errorf("setipaddress: %w", err)
	}

	s.notifier.ServerUpdated(ctx, server)

	return nil
}

func (s *Business) SetProvisioningFailed(ctx context.Context, serverID uuid.UUID) error {
	server, err := s.storer.FindByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("setprovisioningfailed: %w", err)
	}

	if server.Status == StatusProvisionFailed {
		return nil
	}

	server.Status = StatusProvisionFailed

	if err := s.storer.Update(ctx, server); err != nil {
		return fmt.Errorf("setprovisioningfailed: %w", err)
	}

	s.notifier.ServerUpdated(ctx, server)

	return nil
}

func checkOwnership(srv Server, userID uuid.UUID) error {
	if srv.OwnerID != userID {
		return ErrAccessDenied
	}
	return nil
}
