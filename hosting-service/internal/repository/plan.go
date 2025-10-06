package repository

import (
	"context"
	"hosting-service/internal/domain"

	"github.com/google/uuid"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *domain.Plan) error
	FindAll(ctx context.Context, page, pageSize int) ([]*domain.Plan, int64, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*domain.Plan, error)
}
