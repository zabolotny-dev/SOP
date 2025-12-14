package commands

import (
	"context"
	"database/sql"
	"fmt"
	"hosting-service/internal/platform/database"
	"hosting-service/internal/platform/migration"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Migrate(cfg database.Config) error {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Applying migrations...")

	if err := migration.Migrate(ctx, db); err != nil {
		return err
	}

	fmt.Println("Migrations applied successfully!")

	return nil
}
