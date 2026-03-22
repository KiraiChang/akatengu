package database

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "modernc.org/sqlite" // 純 Go SQLite driver
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	m      *migrate.Migrate
	logger *zap.Logger
}

func NewMigrator(dbPath string, logger *zap.Logger) (*Migrator, error) {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create iofs source: %w", err)
	}

	// SQLite 連線字串格式
	databaseURL := fmt.Sprintf("sqlite3://%s", dbPath)

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create migrate instance: %w", err)
	}

	return &Migrator{m: m,
		logger: logger}, nil
}

func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			mg.logger.Info("migrations: no changes")
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}

	version, _, _ := mg.m.Version()
	mg.logger.Info("migrations applied", zap.Uint("current_version", version))
	return nil
}

func (mg *Migrator) Close() {
	srcErr, dbErr := mg.m.Close()
	if srcErr != nil {
		mg.logger.Error("close migration source", zap.Error(srcErr))
	}
	if dbErr != nil {
		mg.logger.Error("close migration db", zap.Error(dbErr))
	}
}
