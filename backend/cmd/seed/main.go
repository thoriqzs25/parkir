package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	authsvc "github.com/thoriqzs/PARKIR/backend/internal/auth"
)

func main() {
	_ = godotenv.Load("backend/.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/parkir?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := seed(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to seed database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("seed completed successfully")
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	roles := []struct {
		name        string
		permissions []string
	}{
		{
			name: "owner",
			permissions: []string{
				"sessions:*", "payments:*", "adjustments:*", "incidents:*",
				"reports:*", "finance:*", "users:*", "locations:*", "rates:*",
				"observability:*", "shifts:*",
			},
		},
		{
			name: "admin",
			permissions: []string{
				"sessions:*", "payments:*", "incidents:*", "reports:*",
				"users:*", "locations:*", "rates:*", "observability:*", "shifts:*",
			},
		},
		{
			name: "manager",
			permissions: []string{
				"sessions:view", "reports:*", "incidents:*", "adjustments:*",
				"locations:*", "rates:*", "shifts:view", "shifts:force_close",
				"shifts:resolve_discrepancy", "observability:view_health",
				"observability:view_alerts", "users:*", "payments:view",
				"payments:void", "finance:view_transactions",
			},
		},
		{
			name: "operator",
			permissions: []string{
				"sessions:view", "sessions:create", "sessions:close",
				"payments:collect_cash", "payments:collect_digital",
				"incidents:create", "shifts:start", "shifts:end",
			},
		},
	}

	for _, role := range roles {
		_, err := pool.Exec(ctx, `
			INSERT INTO roles (name, permissions)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET permissions = EXCLUDED.permissions
		`, role.name, role.permissions)
		if err != nil {
			return fmt.Errorf("insert role %s: %w", role.name, err)
		}
	}

	var ownerRoleID string
	err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'owner'`).Scan(&ownerRoleID)
	if err != nil {
		return fmt.Errorf("find owner role: %w", err)
	}

	passwordHash, err := authsvc.HashPassword("owner123")
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	pinHash, err := authsvc.HashPIN("123456")
	if err != nil {
		return fmt.Errorf("hash pin: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO users (name, email, password_hash, pin_hash, role_id)
		VALUES ('Root Owner', 'owner@parkir.local', $1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			pin_hash = EXCLUDED.pin_hash,
			role_id = EXCLUDED.role_id
	`, passwordHash, pinHash, ownerRoleID)
	if err != nil {
		return fmt.Errorf("insert owner user: %w", err)
	}

	return nil
}
