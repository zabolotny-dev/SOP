package plandb

import (
	"hosting-service/internal/plan"

	"github.com/google/uuid"
)

type planDB struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	CPUCores int       `db:"cpu_cores"`
	RAMMB    int       `db:"ram_mb"`
	DiskGB   int       `db:"disk_gb"`
}

func toDBPlan(p plan.Plan) planDB {
	return planDB{
		ID:       p.ID,
		Name:     p.Name,
		CPUCores: p.CPUCores,
		RAMMB:    p.RAMMB,
		DiskGB:   p.DiskGB,
	}
}

func toBusPlan(db planDB) plan.Plan {
	return plan.Plan{
		ID:       db.ID,
		Name:     db.Name,
		CPUCores: db.CPUCores,
		RAMMB:    db.RAMMB,
		DiskGB:   db.DiskGB,
	}
}

func toBusPlans(dbs []planDB) []plan.Plan {
	plans := make([]plan.Plan, len(dbs))
	for i, db := range dbs {
		plans[i] = toBusPlan(db)
	}
	return plans
}
