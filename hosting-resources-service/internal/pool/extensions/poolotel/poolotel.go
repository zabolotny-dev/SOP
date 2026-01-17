package poolotel

import (
	"context"
	"hosting-kit/otel"
	"hosting-kit/page"
	"hosting-resources-service/internal/pool"

	"github.com/google/uuid"
)

type Extension struct {
	bus pool.ExtBusiness
}

func NewExtension() pool.Extension {
	return func(bus pool.ExtBusiness) pool.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

func (e *Extension) AddResources(ctx context.Context, r pool.Resource, poolID uuid.UUID) (pool.Pool, error) {
	ctx, span := otel.AddSpan(ctx, "pool.addresources")
	defer span.End()

	return e.bus.AddResources(ctx, r, poolID)
}

func (e *Extension) ConsumeResource(ctx context.Context, r pool.Resource) (uuid.UUID, error) {
	ctx, span := otel.AddSpan(ctx, "pool.consumeresource")
	defer span.End()

	return e.bus.ConsumeResource(ctx, r)
}

func (e *Extension) CreatePool(ctx context.Context, p pool.NewPool) (pool.Pool, error) {
	ctx, span := otel.AddSpan(ctx, "pool.create")
	defer span.End()

	return e.bus.CreatePool(ctx, p)
}

func (e *Extension) ReturnResource(ctx context.Context, r pool.Resource, poolID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "pool.returnresource")
	defer span.End()

	return e.bus.ReturnResource(ctx, r, poolID)
}

func (e *Extension) Search(ctx context.Context, pg page.Page) ([]pool.Pool, int, error) {
	ctx, span := otel.AddSpan(ctx, "pool.search")
	defer span.End()

	return e.bus.Search(ctx, pg)
}
