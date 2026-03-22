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

	// 執行 migration
	migrator, err := database.NewMigrator(cfg.DataSource, logger)
	if err != nil {
		return nil, err
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		return nil, err
	}

	// 系統必要資料，每個環境都跑
	seeders := []database.Seeder{
		database.NewSQLFileSeeder("accounts", "accounts.sql"),
	}

	//// 開發環境額外跑假資料
	//if os.Getenv("APP_ENV") == "development" {
	//	seeders = append(seeders,
	//		database.NewSQLFileSeeder("dev_fixtures", "dev_fixtures.sql"),
	//	)
	//}

	runner := database.NewSeedRunner(sqlxdb.DB, seeders...)
	if err := runner.Run(context.Background()); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	logger.Info("database ready")

	return sqlxdb, nil
}
