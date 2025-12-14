package main

import (
	"errors"
	"fmt"
	"hosting-service/cmd/migrator/commands"
	"hosting-service/internal/platform/database"
	"log"
	"os"

	"github.com/ardanlabs/conf/v3"
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := struct {
		Args conf.Args
		DB   struct {
			User     string `conf:"default:postgres"`
			Password string `conf:"default:vladick,mask"`
			Host     string `conf:"default:localhost:5432"`
			Name     string `conf:"default:sop"`
		}
	}{}

	const prefix = "SERV"

	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	dbConfig := database.Config{
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Host:     cfg.DB.Host,
		Name:     cfg.DB.Name,
	}

	return processCommands(cfg.Args, dbConfig)
}

func processCommands(args conf.Args, dbConfig database.Config) error {
	switch args.Num(0) {
	case "migrate", "up":
		return commands.Migrate(dbConfig)

	case "rollback", "down":
		return commands.Rollback(dbConfig)

	case "status":
		return commands.Status(dbConfig)

	case "reset":
		return commands.Reset(dbConfig)

	default:
		fmt.Println("migrate/up:    create the schema in the database")
		fmt.Println("rollback/down: roll back the most recent migration")
		fmt.Println("status:        print the status of all migrations")
		fmt.Println("reset:         roll back all migrations")

		return errors.New("unknown command")
	}
}
