package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func setupGoose() error {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	return nil
}

func RunMigrations(ctx context.Context, db *sqlx.DB, logger *zap.Logger) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db.DB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	version, err := goose.GetDBVersion(db.DB)
	if err == nil {
		logger.Info("migrations applied", zap.Int64("version", version))
	}

	return nil
}

// MigrateUpTo applies the embedded migrations up to and including version, with
// the same goose settings as RunMigrations. Tests use it to write fixtures on an
// earlier schema before running a data migration.
func MigrateUpTo(ctx context.Context, db *sqlx.DB, version int64) error {
	if err := setupGoose(); err != nil {
		return err
	}

	if err := goose.UpToContext(ctx, db.DB, "migrations", version); err != nil {
		return fmt.Errorf("goose up to %d: %w", version, err)
	}
	return nil
}
