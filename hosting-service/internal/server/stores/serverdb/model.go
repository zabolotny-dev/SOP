package serverdb

import (
	"hosting-service/internal/server"
	"time"

	"github.com/google/uuid"
)

type serverDB struct {
	ID          uuid.UUID `db:"id"`
	IPv4Address *string   `db:"ipv4_address"`
	OwnerID     uuid.UUID `db:"owner_id"`
	PoolID      uuid.UUID `db:"pool_id"`
	PlanID      uuid.UUID `db:"plan_id"`
	Name        string    `db:"name"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
}

func toDBServer(s server.Server) serverDB {
	return serverDB{
		ID:          s.ID,
		IPv4Address: s.IPv4Address,
		OwnerID:     s.OwnerID,
		PoolID:      s.PoolID,
		PlanID:      s.PlanID,
		Name:        s.Name,
		Status:      string(s.Status),
		CreatedAt:   s.CreatedAt,
	}
}

func toBusServer(db serverDB) server.Server {
	return server.Server{
		ID:          db.ID,
		IPv4Address: db.IPv4Address,
		OwnerID:     db.OwnerID,
		PoolID:      db.PoolID,
		PlanID:      db.PlanID,
		Name:        db.Name,
		Status:      server.ServerStatus(db.Status),
		CreatedAt:   db.CreatedAt,
	}
}

func toBusServers(dbs []serverDB) []server.Server {
	servers := make([]server.Server, len(dbs))
	for i, db := range dbs {
		servers[i] = toBusServer(db)
	}
	return servers
}
