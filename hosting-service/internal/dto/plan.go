package dto

import "github.com/google/uuid"

type PlanPreview struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	CPUCores int       `json:"cpuCores"`
	RAMMB    int       `json:"ramMb"`
	DiskGB   int       `json:"diskGb"`
}

type PlanSearch struct {
	Data []*PlanPreview   `json:"data"`
	Meta PaginationResult `json:"meta"`
}
