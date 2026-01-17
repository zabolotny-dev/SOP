package planotel

import (
	"context"
	"hosting-kit/otel"
	"hosting-kit/page"
	"hosting-service/internal/plan"

	"github.com/google/uuid"
)

type Extension struct {
	bus plan.ExtBusiness
}

func NewExtension() plan.Extension {
	return func(bus plan.ExtBusiness) plan.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

func (e *Extension) Create(ctx context.Context, params plan.CreatePlanParams) (plan.Plan, error) {
	ctx, span := otel.AddSpan(ctx, "plan.create")
	defer span.End()

	return e.bus.Create(ctx, params)
}

func (e *Extension) FindByID(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
	ctx, span := otel.AddSpan(ctx, "plan.findbyid")
	defer span.End()

	return e.bus.FindByID(ctx, ID)
}

func (e *Extension) Search(ctx context.Context, pg page.Page) ([]plan.Plan, int, error) {
	ctx, span := otel.AddSpan(ctx, "plan.search")
	defer span.End()

	return e.bus.Search(ctx, pg)
}
