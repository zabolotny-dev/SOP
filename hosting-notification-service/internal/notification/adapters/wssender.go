package adapters

import (
	"context"
	"encoding/json"
	"hosting-kit/otel"
	"hosting-notification-service/internal/notification"
	"hosting-notification-service/internal/platform/websocket"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type WSAdapter struct {
	hub *websocket.Hub
}

func NewWSAdapter(hub *websocket.Hub) *WSAdapter {
	return &WSAdapter{hub: hub}
}

func (w *WSAdapter) Send(ctx context.Context, userID uuid.UUID, event notification.Event) error {
	ctx, span := otel.AddSpan(ctx, "notification.websocket.send")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("event_type", event.Type),
	)

	data, err := json.Marshal(event)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := w.hub.Send(ctx, userID, data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
