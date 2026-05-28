package bootstrap

import (
	"akatengu/internal/database"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
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
		database.NewSQLFileSeeder("merchants", "merchants.sql"),
		database.NewSQLFileSeeder("users", "users.sql"),
		database.NewSQLFileSeeder("accounts", "accounts.sql"),
		database.NewSQLFileSeeder("sys_accounts", "sys_accounts.sql"),
		database.NewSQLFileSeeder("account_type_configs", "account_type_configs.sql"),
	}

	runner := database.NewSeedRunner(sqlxdb.DB, logger, seeders...)
	if err := runner.Run(ctx); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	// Event seed bootstrap: 確保 merchant 1 的 seed 資料有對應事件，replay 才能完整重建
	uow := event_store.NewUnitOfWork(sqlxdb)
	queryRepo := query.NewQueryRepository(sqlxdb)
	es := services.NewEventStoreService(uow, queryRepo)
	if err := BootstrapEventSeeds(ctx, sqlxdb, es, logger); err != nil {
		return nil, fmt.Errorf("event seed: %w", err)
	}

	logger.Info("database ready")

	return sqlxdb, nil
}
