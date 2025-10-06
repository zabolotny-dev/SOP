package dto

import (
	"time"

	"github.com/google/uuid"
)

type ServerPreview struct {
	ID        uuid.UUID `json:"id"`
	PlanID    uuid.UUID `json:"planId"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type ServerSearch struct {
	Data []*ServerPreview `json:"data"`
	Meta PaginationResult `json:"meta"`
}
