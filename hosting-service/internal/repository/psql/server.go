package psql

import (
	"context"
	"errors"
	"hosting-service/internal/domain"
	"hosting-service/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type serverRepositoryImpl struct {
	db *gorm.DB
}

func (s *serverRepositoryImpl) FindAll(ctx context.Context, page, pageSize int) ([]*domain.Server, int64, error) {
	var servers []*domain.Server
	var count int64

	query := s.db.WithContext(ctx).Model(&domain.Server{})

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return []*domain.Server{}, 0, nil
	}

	paginatedQuery := query.Scopes(repository.PaginationWithParams(page, pageSize)).Find(&servers)
	if paginatedQuery.Error != nil {
		return nil, 0, paginatedQuery.Error
	}

	return servers, count, nil
}

func (s *serverRepositoryImpl) FindByID(ctx context.Context, ID uuid.UUID) (*domain.Server, error) {
	var server domain.Server
	result := s.db.WithContext(ctx).First(&server, "id = ?", ID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrServerNotFound
		}
		return nil, result.Error
	}
	return &server, nil
}

func (s *serverRepositoryImpl) Save(ctx context.Context, server *domain.Server) error {
	return s.db.WithContext(ctx).Save(server).Error
}

func NewServerRepository(db *gorm.DB) repository.ServerRepository {
	return &serverRepositoryImpl{db: db}
}
