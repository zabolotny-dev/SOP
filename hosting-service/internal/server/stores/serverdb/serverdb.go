package serverdb

import (
	"context"
	"fmt"
	"hosting-service/internal/platform/page"
	"hosting-service/internal/server"

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

func (s *Store) FindByID(ctx context.Context, ID uuid.UUID) (server.Server, error) {
	const q = `
	SELECT 
		id, plan_id, name, ipv4_address, status, created_at
	FROM 
		servers 
	WHERE 
		id = @id`

	args := pgx.NamedArgs{
		"id": ID,
	}

	rows, err := s.db.Query(ctx, q, args)
	if err != nil {
		return server.Server{}, fmt.Errorf("db: %w", err)
	}

	dbServer, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[serverDB])
	if err != nil {
		if err == pgx.ErrNoRows {
			return server.Server{}, server.ErrServerNotFound
		}
		return server.Server{}, fmt.Errorf("db: %w", err)
	}

	return toBusServer(dbServer), nil
}

func (s *Store) Create(ctx context.Context, srv server.Server) error {
	const q = `
	INSERT INTO servers 
		(id, plan_id, name, ipv4_address, status, created_at)
	VALUES 
		(@id, @plan_id, @name, @ipv4_address, @status, @created_at)`

	dbServer := toDBServer(srv)

	args := pgx.NamedArgs{
		"id":           dbServer.ID,
		"plan_id":      dbServer.PlanID,
		"name":         dbServer.Name,
		"ipv4_address": dbServer.IPv4Address,
		"status":       dbServer.Status,
		"created_at":   dbServer.CreatedAt,
	}

	_, err := s.db.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) FindAll(ctx context.Context, pg page.Page) ([]server.Server, int, error) {
	const qCount = `SELECT count(*) FROM servers`

	var total int
	err := s.db.QueryRow(ctx, qCount).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	const q = `
	SELECT 
		id, plan_id, name, ipv4_address, status, created_at
	FROM 
		servers
	ORDER BY 
		created_at DESC
	LIMIT 
		@limit 
	OFFSET 
		@offset`

	args := pgx.NamedArgs{
		"limit":  pg.Size(),
		"offset": pg.Offset(),
	}

	rows, err := s.db.Query(ctx, q, args)
	if err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	dbServers, err := pgx.CollectRows(rows, pgx.RowToStructByName[serverDB])
	if err != nil {
		return nil, 0, fmt.Errorf("db: %w", err)
	}

	return toBusServers(dbServers), total, nil
}

func (s *Store) Update(ctx context.Context, srv server.Server) error {
	const q = `
	UPDATE servers
	SET 
		plan_id = @plan_id,
		name = @name,
		ipv4_address = @ipv4_address,
		status = @status
	WHERE 
		id = @id`

	dbServer := toDBServer(srv)

	args := pgx.NamedArgs{
		"id":           dbServer.ID,
		"plan_id":      dbServer.PlanID,
		"name":         dbServer.Name,
		"ipv4_address": dbServer.IPv4Address,
		"status":       dbServer.Status,
	}

	_, err := s.db.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, ID uuid.UUID) error {
	const q = `
	DELETE FROM servers
	WHERE id = @id`

	args := pgx.NamedArgs{
		"id": ID,
	}

	_, err := s.db.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
