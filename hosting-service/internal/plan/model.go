package plan

import (
	"github.com/google/uuid"
)

type Plan struct {
	ID       uuid.UUID
	Name     string
	CPUCores int
	RAMMB    int
	DiskGB   int
}

type CreatePlanParams struct {
	Name     string
	CPUCores int
	RAMMB    int
	DiskGB   int
}
