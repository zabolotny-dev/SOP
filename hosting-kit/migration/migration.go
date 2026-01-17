package migration

import (
	"context"
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

func setup(fs embed.FS) error {
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return nil
}

func Migrate(ctx context.Context, db *sql.DB, fs embed.FS, dir string) error {
	if err := setup(fs); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db, dir); err != nil {
		return err
	}

	return nil
}

func Rollback(ctx context.Context, db *sql.DB, fs embed.FS, dir string) error {
	if err := setup(fs); err != nil {
		return err
	}

	if err := goose.DownContext(ctx, db, dir); err != nil {
		return err
	}

	return nil
}

func Status(ctx context.Context, db *sql.DB, fs embed.FS, dir string) error {
	if err := setup(fs); err != nil {
		return err
	}

	if err := goose.StatusContext(ctx, db, dir); err != nil {
		return err
	}
	return nil
}

func Reset(ctx context.Context, db *sql.DB, fs embed.FS, dir string) error {
	if err := setup(fs); err != nil {
		return err
	}

	if err := goose.ResetContext(ctx, db, dir); err != nil {
		return err
	}

	return nil
}
