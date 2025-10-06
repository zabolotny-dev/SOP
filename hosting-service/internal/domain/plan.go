package domain

import (
	"fmt"
	"strings"
)

type Plan struct {
	BaseModel
	Name     string `gorm:"type:varchar(255);not null"`
	CPUCores int    `gorm:"not null"`
	RAMMB    int    `gorm:"not null"`
	DiskGB   int    `gorm:"not null"`
}

type NewPlanParams struct {
	Name     string
	CpuCores int
	RamMb    int
	DiskGb   int
}

func NewPlan(params NewPlanParams) (*Plan, error) {
	trimmedName := strings.TrimSpace(params.Name)
	if trimmedName == "" {
		return nil, fmt.Errorf("%w :plan name cannot be empty", ErrValidation)
	}

	if params.CpuCores <= 0 {
		return nil, fmt.Errorf("%w: CPU cores must be a positive number", ErrValidation)
	}
	if params.RamMb <= 0 {
		return nil, fmt.Errorf("%w: RAM in MB must be a positive number", ErrValidation)
	}
	if params.DiskGb <= 0 {
		return nil, fmt.Errorf("%w: disk in GB must be a positive number", ErrValidation)
	}

	plan := &Plan{
		BaseModel: NewBaseModel(),
		Name:      trimmedName,
		CPUCores:  params.CpuCores,
		RAMMB:     params.RamMb,
		DiskGB:    params.DiskGb,
	}

	return plan, nil
}
