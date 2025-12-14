package migration

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func setup() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	return nil
}

func Migrate(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}

func Rollback(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.DownContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("migrate down failed: %w", err)
	}

	return nil
}

func Status(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.StatusContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("migrate status failed: %w", err)
	}
	return nil
}

func Reset(ctx context.Context, db *sql.DB) error {
	if err := setup(); err != nil {
		return err
	}

	if err := goose.ResetContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("migrate reset failed: %w", err)
	}

	return nil
}
