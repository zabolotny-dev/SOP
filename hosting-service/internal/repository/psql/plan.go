package psql

import (
	"context"
	"errors"
	"hosting-service/internal/domain"
	"hosting-service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRepositoryImpl struct {
	db *gorm.DB
}

func (p *planRepositoryImpl) FindByID(ctx context.Context, ID uuid.UUID) (*domain.Plan, error) {
	var plan domain.Plan
	result := p.db.WithContext(ctx).First(&plan, "id = ?", ID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrPlanNotFound
		}
		return nil, result.Error
	}
	return &plan, nil
}

func (p *planRepositoryImpl) FindAll(ctx context.Context, page, pageSize int) ([]*domain.Plan, int64, error) {
	var plans []*domain.Plan
	var count int64

	query := p.db.WithContext(ctx).Model(&domain.Plan{})

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return []*domain.Plan{}, 0, nil
	}

	paginatedQuery := query.Scopes(repository.PaginationWithParams(page, pageSize)).Find(&plans)
	if paginatedQuery.Error != nil {
		return nil, 0, paginatedQuery.Error
	}

	return plans, count, nil
}

func (p *planRepositoryImpl) Save(ctx context.Context, plan *domain.Plan) error {
	return p.db.WithContext(ctx).Save(plan).Error
}

func NewPlanRepository(db *gorm.DB) repository.PlanRepository {
	return &planRepositoryImpl{db: db}
}
