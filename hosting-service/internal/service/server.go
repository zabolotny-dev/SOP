package service

import (
	"context"
	"errors"
	"hosting-service/internal/domain"
	"hosting-service/internal/dto"
	"hosting-service/internal/repository"

	"github.com/google/uuid"
)

type ActionType string

const (
	ActionStart  ActionType = "START"
	ActionStop   ActionType = "STOP"
	ActionReboot ActionType = "REBOOT"
	ActionDelete ActionType = "DELETE"
)

type CreateServerParams struct {
	Name   string
	PlanID uuid.UUID
}

type PerformActionParams struct {
	ServerID uuid.UUID
	Action   ActionType
}

var (
	ErrServerNotFound = errors.New("server not found")
	ErrInvalidAction  = errors.New("invalid action")
)

type ServerService interface {
	Save(ctx context.Context, params CreateServerParams) (*dto.ServerPreview, error)
	Search(ctx context.Context, page, pageSize int) (*dto.ServerSearch, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*dto.ServerPreview, error)
	PerformAction(ctx context.Context, params PerformActionParams) (*dto.ServerPreview, error)
}

type serverServiceImpl struct {
	serverRepository repository.ServerRepository
}

func (s *serverServiceImpl) FindByID(ctx context.Context, ID uuid.UUID) (*dto.ServerPreview, error) {
	server, err := s.serverRepository.FindByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repository.ErrServerNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, err
	}

	return &dto.ServerPreview{
		ID:        server.ID,
		PlanID:    server.PlanID,
		Name:      server.Name,
		Status:    string(server.Status),
		CreatedAt: server.CreatedAt,
	}, nil
}

func (s *serverServiceImpl) PerformAction(ctx context.Context, params PerformActionParams) (*dto.ServerPreview, error) {
	server, err := s.serverRepository.FindByID(ctx, params.ServerID)
	if err != nil {
		if errors.Is(err, repository.ErrServerNotFound) {
			return nil, ErrServerNotFound
		}
		return nil, err
	}

	switch params.Action {
	case ActionStart:
		err = server.Start()
	case ActionStop:
		err = server.Stop()
	case ActionReboot:
		err = server.Reboot()
	case ActionDelete:
		err = server.MarkForDeletion()
	default:
		return nil, ErrInvalidAction
	}

	if err != nil {
		return nil, err
	}

	err = s.serverRepository.Save(ctx, server)

	if err != nil {
		return nil, err
	}

	return &dto.ServerPreview{
		ID:        server.ID,
		PlanID:    server.PlanID,
		Name:      server.Name,
		Status:    string(server.Status),
		CreatedAt: server.CreatedAt,
	}, nil
}

func (s *serverServiceImpl) Save(ctx context.Context, params CreateServerParams) (*dto.ServerPreview, error) {
	server, err := domain.NewServer(params.PlanID, params.Name)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			return nil, err
		}
		return nil, err
	}

	err = s.serverRepository.Save(ctx, server)
	if err != nil {
		return nil, err
	}

	return &dto.ServerPreview{
		ID:        server.ID,
		PlanID:    server.PlanID,
		Name:      server.Name,
		Status:    string(server.Status),
		CreatedAt: server.CreatedAt,
	}, nil
}

func (s *serverServiceImpl) Search(ctx context.Context, page int, pageSize int) (*dto.ServerSearch, error) {
	servers, count, err := s.serverRepository.FindAll(ctx, page, pageSize)

	if err != nil {
		return nil, err
	}

	data := make([]*dto.ServerPreview, len(servers))
	for i, server := range servers {
		data[i] = &dto.ServerPreview{
			ID:        server.ID,
			PlanID:    server.PlanID,
			Name:      server.Name,
			Status:    string(server.Status),
			CreatedAt: server.CreatedAt,
		}
	}

	return &dto.ServerSearch{
		Data: data,
		Meta: repository.CalculatePaginationResult(page, pageSize, count),
	}, nil
}

func NewServerService(serverRepository repository.ServerRepository) ServerService {
	return &serverServiceImpl{serverRepository: serverRepository}
}
