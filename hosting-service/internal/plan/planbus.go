package plan

import (
	"context"
	"errors"
	"fmt"
	"hosting-kit/page"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrValidation   = errors.New("validation error")
	ErrPlanNotFound = errors.New("plan not found")
)

type Extension func(ExtBusiness) ExtBusiness

type Storer interface {
	FindByID(ctx context.Context, ID uuid.UUID) (Plan, error)
	Create(ctx context.Context, plan Plan) error
	FindAll(ctx context.Context, pg page.Page) ([]Plan, int, error)
}

type ExtBusiness interface {
	FindByID(ctx context.Context, ID uuid.UUID) (Plan, error)
	Create(ctx context.Context, params CreatePlanParams) (Plan, error)
	Search(ctx context.Context, pg page.Page) ([]Plan, int, error)
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

func NewPlan(params CreatePlanParams) (Plan, error) {
	trimmedName := strings.TrimSpace(params.Name)
	if trimmedName == "" {
		return Plan{}, fmt.Errorf("%w :plan name cannot be empty", ErrValidation)
	}

	if params.CPUCores <= 0 {
		return Plan{}, fmt.Errorf("%w: CPU cores must be a positive number", ErrValidation)
	}
	if params.RAMMB <= 0 {
		return Plan{}, fmt.Errorf("%w: RAM in MB must be a positive number", ErrValidation)
	}
	if params.DiskGB <= 0 {
		return Plan{}, fmt.Errorf("%w: disk in GB must be a positive number", ErrValidation)
	}
	if params.IpCount <= 0 {
		return Plan{}, fmt.Errorf("%w: IP count must be a positive number", ErrValidation)
	}

	plan := Plan{
		ID:       uuid.New(),
		Name:     trimmedName,
		CPUCores: params.CPUCores,
		RAMMB:    params.RAMMB,
		DiskGB:   params.DiskGB,
		IpCount:  params.IpCount,
	}

	return plan, nil
}

func (b *Business) FindByID(ctx context.Context, ID uuid.UUID) (Plan, error) {
	plan, err := b.storer.FindByID(ctx, ID)
	if err != nil {
		return Plan{}, fmt.Errorf("findbyid: %w", err)
	}

	return plan, nil
}

func (b *Business) Create(ctx context.Context, params CreatePlanParams) (Plan, error) {
	plan, err := NewPlan(params)
	if err != nil {
		return Plan{}, err
	}

	err = b.storer.Create(ctx, plan)

	if err != nil {
		return Plan{}, fmt.Errorf("create: %w", err)
	}

	return plan, nil
}

func (b *Business) Search(ctx context.Context, pg page.Page) ([]Plan, int, error) {
	plans, total, err := b.storer.FindAll(ctx, pg)

	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}

	return plans, total, nil
}
