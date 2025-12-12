package plandb

import (
	"context"
	"fmt"
	"hosting-service/internal/plan"
	"hosting-service/internal/platform/page"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, p plan.Plan) error {
	const q = `
	INSERT INTO plans 
		(id, name, cpu_cores, ram_mb, disk_gb)
	VALUES 
		(@id, @name, @cpu_cores, @ram_mb, @disk_gb)`

	dbPlan := toDBPlan(p)

	args := pgx.NamedArgs{
		"id":        dbPlan.ID,
		"name":      dbPlan.Name,
		"cpu_cores": dbPlan.CPUCores,
		"ram_mb":    dbPlan.RAMMB,
		"disk_gb":   dbPlan.DiskGB,
	}

	_, err := s.db.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) FindByID(ctx context.Context, ID uuid.UUID) (plan.Plan, error) {
	const q = `
	SELECT 
		id, name, cpu_cores, ram_mb, disk_gb 
	FROM 
		plans 
	WHERE 
		id = @id`

	args := pgx.NamedArgs{
		"id": ID,
	}

	rows, err := s.db.Query(ctx, q, args)
	if err != nil {
		return plan.Plan{}, fmt.Errorf("db: %w", err)
	}

	dbPlan, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[planDB])
	if err != nil {
		if err == pgx.ErrNoRows {
			return plan.Plan{}, plan.ErrPlanNotFound
		}
		return plan.Plan{}, fmt.Errorf("db: %w", err)
	}

	return toBusPlan(dbPlan), nil
}

func (s *Store) FindAll(ctx context.Context, pg page.Page) ([]plan.Plan, int, error) {
	const qCount = `SELECT count(*) FROM plans`

	var total int
	if err := s.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	if total == 0 {
		return []plan.Plan{}, 0, nil
	}

	const qSelect = `
	SELECT 
		id, name, cpu_cores, ram_mb, disk_gb 
	FROM 
		plans
	ORDER BY 
		id ASC
	LIMIT @limit OFFSET @offset`

	args := pgx.NamedArgs{
		"limit":  pg.Size(),
		"offset": pg.Offset(),
	}

	rows, err := s.db.Query(ctx, qSelect, args)
	if err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	dbPlans, err := pgx.CollectRows(rows, pgx.RowToStructByName[planDB])
	if err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	return toBusPlans(dbPlans), total, nil
}
