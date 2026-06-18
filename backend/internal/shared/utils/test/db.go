package test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"akatengu/internal/database"
	"akatengu/internal/enums"

	_ "modernc.org/sqlite"
)

var initEnumsOnce sync.Once

// NewTestDB opens an in-memory SQLite DB, runs all migrations, and seeds accounts.
// The DB is closed automatically when the test ends.
func NewTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, closeDB := newTestDBInner(t.Fatalf)
	t.Cleanup(closeDB)
	return db
}

// NewTestDBGinkgo is the Ginkgo-compatible variant: the caller must register
// cleanup with DeferCleanup(closeDB) or call closeDB() in AfterEach.
func NewTestDBGinkgo() (*sqlx.DB, func()) {
	return newTestDBInner(func(format string, args ...any) {
		panic(fmt.Sprintf("testutil.NewTestDBGinkgo: "+format, args...))
	})
}

func newTestDBInner(fatalf func(string, ...any)) (*sqlx.DB, func()) {
	initEnumsOnce.Do(enums.InitEnums)
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		fatalf("open test db: %v", err)
	}
	if err := database.RunMigrations(context.Background(), db, zap.NewNop()); err != nil {
		db.Close()
		fatalf("run migrations: %v", err)
	}
	seeder := database.NewSQLFileSeeder("accounts", "accounts.sql")
	if err := seeder.Seed(context.Background(), db.DB); err != nil {
		db.Close()
		fatalf("seed accounts: %v", err)
	}
	return db, func() { db.Close() }
}
