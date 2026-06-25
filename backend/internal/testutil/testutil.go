package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thoriqzs/PARKIR/backend/internal/db"
	"github.com/thoriqzs/PARKIR/backend/internal/store"
)

// TestDB wraps a connection pool and store for integration tests.
type TestDB struct {
	Pool    *pgxpool.Pool
	Store   *store.Store
	Cleanup func()
}

// NewTestDB creates a test database connection, runs migrations, and returns a cleanup function.
// Tests are skipped if PARKIR_TEST_DATABASE_URL is not set or the database is unreachable.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	databaseURL := os.Getenv("PARKIR_TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		t.Skip("PARKIR_TEST_DATABASE_URL or DATABASE_URL not set; skipping integration test")
	}

	pool, err := db.NewPool(databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		pool.Close()
		t.Fatal("failed to determine testutil path")
	}
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	if err := db.Migrate(databaseURL, migrationsPath); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	s := store.New(pool)

	reset := func() {
		ctx := context.Background()
		tables := []string{
			"audit_logs",
			"transactions",
			"sessions",
			"shifts",
			"receipt_sequences",
			"location_rates",
			"user_role_locations",
			"user_permission_grants",
			"locations",
			"users",
			"roles",
		}
		for _, table := range tables {
			if _, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
				t.Logf("failed to truncate %s: %v", table, err)
			}
		}
	}

	cleanup := func() {
		reset()
		pool.Close()
	}

	reset()

	return &TestDB{Pool: pool, Store: s, Cleanup: cleanup}
}

// Ctx returns a background context.
func Ctx() context.Context {
	return context.Background()
}
