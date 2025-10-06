package repository

import (
	"context"
	"hosting-service/internal/domain"

	"github.com/google/uuid"
)

type ServerRepository interface {
	Save(ctx context.Context, server *domain.Server) error
	FindByID(ctx context.Context, ID uuid.UUID) (*domain.Server, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*domain.Server, int64, error)
}
