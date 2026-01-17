package provisioningotel

import (
	"context"
	"hosting-kit/otel"
	"hosting-provisioning-service/internal/provisioning"

	"github.com/google/uuid"
)

type Extension struct {
	bus provisioning.ExtBusiness
}

func NewExtension() provisioning.Extension {
	return func(bus provisioning.ExtBusiness) provisioning.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

func (e *Extension) GenerateIP(ctx context.Context, serverID uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "provisioning.generateip")
	defer span.End()

	return e.bus.GenerateIP(ctx, serverID)
}
