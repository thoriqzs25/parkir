package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("backend/.env")

	direction := flag.String("direction", "up", "migration direction: up or down")
	steps := flag.Int("steps", 0, "number of migrations to apply (0 = all)")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/parkir?sslmode=disable"
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrate instance: %v\n", err)
		os.Exit(1)
	}

	var migrateErr error

	switch *direction {
	case "up":
		if *steps > 0 {
			migrateErr = m.Steps(*steps)
		} else {
			migrateErr = m.Up()
		}
	case "down":
		if *steps > 0 {
			migrateErr = m.Steps(-*steps)
		} else {
			migrateErr = m.Down()
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid direction: %s\n", *direction)
		os.Exit(1)
	}

	if migrateErr != nil && migrateErr != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", migrateErr)
		os.Exit(1)
	}

	version, dirty, _ := m.Version()
	fmt.Printf("migrations applied. version: %d, dirty: %s\n", version, strconv.FormatBool(dirty))
}
