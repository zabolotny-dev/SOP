package serverotel

import (
	"context"
	"hosting-kit/otel"
	"hosting-kit/page"
	"hosting-service/internal/server"

	"github.com/google/uuid"
)

type Extension struct {
	bus server.ExtBusiness
}

func NewExtension() server.Extension {
	return func(bus server.ExtBusiness) server.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

func (e *Extension) Create(ctx context.Context, name string, planID uuid.UUID) (server.Server, error) {
	ctx, span := otel.AddSpan(ctx, "server.create")
	defer span.End()

	return e.bus.Create(ctx, name, planID)
}

func (e *Extension) Delete(ctx context.Context, serverID uuid.UUID) (server.Server, error) {
	ctx, span := otel.AddSpan(ctx, "server.delete")
	defer span.End()

	return e.bus.Delete(ctx, serverID)
}

func (e *Extension) FindByID(ctx context.Context, ID uuid.UUID) (server.Server, error) {
	ctx, span := otel.AddSpan(ctx, "server.findbyid")
	defer span.End()

	return e.bus.FindByID(ctx, ID)
}

func (e *Extension) Search(ctx context.Context, pg page.Page) ([]server.Server, int, error) {
	ctx, span := otel.AddSpan(ctx, "server.search")
	defer span.End()

	return e.bus.Search(ctx, pg)
}

func (e *Extension) SetIPAddress(ctx context.Context, serverID uuid.UUID, ip string) error {
	ctx, span := otel.AddSpan(ctx, "server.setipaddress")
	defer span.End()

	return e.bus.SetIPAddress(ctx, serverID, ip)
}

func (e *Extension) SetProvisioningFailed(ctx context.Context, serverID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "server.setprovisioningfailed")
	defer span.End()

	return e.bus.SetProvisioningFailed(ctx, serverID)
}

func (e *Extension) Start(ctx context.Context, serverID uuid.UUID) (server.Server, error) {
	ctx, span := otel.AddSpan(ctx, "server.start")
	defer span.End()

	return e.bus.Start(ctx, serverID)
}

func (e *Extension) Stop(ctx context.Context, serverID uuid.UUID) (server.Server, error) {
	ctx, span := otel.AddSpan(ctx, "server.stop")
	defer span.End()

	return e.bus.Stop(ctx, serverID)
}
