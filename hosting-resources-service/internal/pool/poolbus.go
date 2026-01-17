package pool

import (
	"context"
	"errors"
	"fmt"
	"hosting-kit/page"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrNotEnoughResources = errors.New("not enough resources available")
	ErrPoolNotFound       = errors.New("pool not found")
)

type Extension func(ExtBusiness) ExtBusiness

type Storer interface {
	AppendResource(ctx context.Context, r Resource, poolID uuid.UUID) (Pool, error)
	SubtractResource(ctx context.Context, r Resource) (uuid.UUID, error)
	CreatePool(ctx context.Context, p Pool) error
	FindAll(ctx context.Context, pg page.Page) ([]Pool, int, error)
}

type ExtBusiness interface {
	CreatePool(ctx context.Context, p NewPool) (Pool, error)
	ConsumeResource(ctx context.Context, r Resource) (uuid.UUID, error)
	ReturnResource(ctx context.Context, r Resource, poolID uuid.UUID) error
	AddResources(ctx context.Context, r Resource, poolID uuid.UUID) (Pool, error)
	Search(ctx context.Context, pg page.Page) ([]Pool, int, error)
}

type Business struct {
	storer     Storer
	extensions []Extension
}

func NewBusiness(storer Storer, extensions ...Extension) ExtBusiness {
	b := &Business{
		storer:     storer,
		extensions: extensions,
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

func (b *Business) CreatePool(ctx context.Context, p NewPool) (Pool, error) {
	trimmedName := strings.TrimSpace(p.Name)
	if trimmedName == "" {
		return Pool{}, fmt.Errorf("%w: pool name cannot be empty", ErrValidation)
	}

	resource := Resource{
		CPUCores: p.CPUCores,
		RAMMB:    p.RAMMB,
		DiskGB:   p.DiskGB,
		IPCount:  p.IPCount,
	}

	if err := validateResource(resource); err != nil {
		return Pool{}, err
	}

	pool := Pool{
		ID:        uuid.New(),
		Name:      p.Name,
		Resources: resource,
	}

	if err := b.storer.CreatePool(ctx, pool); err != nil {
		return Pool{}, fmt.Errorf("create: %w", err)
	}

	return pool, nil
}

func (b *Business) ConsumeResource(ctx context.Context, r Resource) (uuid.UUID, error) {
	if err := validateResource(r); err != nil {
		return uuid.UUID{}, err
	}

	poolID, err := b.storer.SubtractResource(ctx, r)

	if err != nil {
		return uuid.Nil, fmt.Errorf("consume resourses: %w", err)
	}

	return poolID, nil
}

func (b *Business) ReturnResource(ctx context.Context, r Resource, poolID uuid.UUID) error {
	if err := validateResource(r); err != nil {
		return err
	}

	if _, err := b.storer.AppendResource(ctx, r, poolID); err != nil {
		return fmt.Errorf("return resources: %w", err)
	}

	return nil
}

func (b *Business) AddResources(ctx context.Context, r Resource, poolID uuid.UUID) (Pool, error) {
	if err := validateResource(r); err != nil {
		return Pool{}, err
	}

	pool, err := b.storer.AppendResource(ctx, r, poolID)

	if err != nil {
		return Pool{}, fmt.Errorf("add resources: %w", err)
	}

	return pool, nil
}

func (b *Business) Search(ctx context.Context, pg page.Page) ([]Pool, int, error) {
	pools, count, err := b.storer.FindAll(ctx, pg)

	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}

	return pools, count, nil
}

func validateResource(r Resource) error {
	if r.CPUCores < 0 {
		return fmt.Errorf("%w: CPU cores cannot be negative", ErrValidation)
	}
	if r.RAMMB < 0 {
		return fmt.Errorf("%w: RAM cannot be negative", ErrValidation)
	}
	if r.DiskGB < 0 {
		return fmt.Errorf("%w: disk size cannot be negative", ErrValidation)
	}
	if r.IPCount < 0 {
		return fmt.Errorf("%w: IP count cannot be negative", ErrValidation)
	}

	return nil
}
