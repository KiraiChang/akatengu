package testutil

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"akatengu/internal/database"
	_ "modernc.org/sqlite"
)

// NewTestDB opens an in-memory SQLite DB, runs all migrations, and seeds accounts.
// The DB is closed automatically when the test ends.
func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.RunMigrations(context.Background(), db, zap.NewNop()); err != nil {
		db.Close()
		t.Fatalf("run migrations: %v", err)
	}
	seeder := database.NewSQLFileSeeder("accounts", "accounts.sql")
	if err := seeder.Seed(context.Background(), db.DB); err != nil {
		db.Close()
		t.Fatalf("seed accounts: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
