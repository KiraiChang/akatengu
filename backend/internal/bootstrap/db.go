package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"

	"akatengu/internal/database"
)

func InitDB(cfg DBConfig, logger *zap.Logger) (*sqlx.DB, error) {
	db, err := sql.Open(
		cfg.Driver,
		"file:"+cfg.DataSource+cfg.Param,
	)
	if err != nil {
		return nil, err
	}

	sqlxdb := sqlx.NewDb(db, cfg.Driver)

	sqlxdb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlxdb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlxdb.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlxdb.Ping(); err != nil {
		return nil, err
	}

	ctx := context.Background()

	if err := database.RunMigrations(ctx, sqlxdb, logger); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	seeders := []database.Seeder{
		database.NewSQLFileSeeder("accounts", "accounts.sql"),
	}

	runner := database.NewSeedRunner(sqlxdb.DB, logger, seeders...)
	if err := runner.Run(ctx); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	logger.Info("database ready")

	return sqlxdb, nil
}
