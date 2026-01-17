package database

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := cfg.DSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Name)
}
