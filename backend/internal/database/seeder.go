package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
)

//go:embed seeds/*.sql
var seedsFS embed.FS

// Seeder 介面
type Seeder interface {
	Name() string
	Seed(ctx context.Context, db *sql.DB) error
}

// SQLFileSeeder 從 embed SQL 檔案執行 seed
type SQLFileSeeder struct {
	name     string
	filename string
}

func NewSQLFileSeeder(name, filename string) *SQLFileSeeder {
	return &SQLFileSeeder{name: name, filename: filename}
}

func (s *SQLFileSeeder) Name() string { return s.name }

func (s *SQLFileSeeder) Seed(ctx context.Context, db *sql.DB) error {
	// 從 embed.FS 讀取 SQL 檔案
	content, err := seedsFS.ReadFile(fmt.Sprintf("seeds/%s", s.filename))
	if err != nil {
		return fmt.Errorf("read seed file %s: %w", s.filename, err)
	}

	// 在 transaction 內執行，失敗自動回滾
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("exec seed %s: %w", s.filename, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed %s: %w", s.filename, err)
	}

	return nil
}

// SeedRunner 管理並執行所有 seeder
type SeedRunner struct {
	db      *sql.DB
	seeders []Seeder
}

func NewSeedRunner(db *sql.DB, seeders ...Seeder) *SeedRunner {
	return &SeedRunner{db: db, seeders: seeders}
}

func (sr *SeedRunner) Run(ctx context.Context) error {
	for _, s := range sr.seeders {
		slog.Info("running seeder", "name", s.Name())

		if err := s.Seed(ctx, sr.db); err != nil {
			return fmt.Errorf("seeder [%s]: %w", s.Name(), err)
		}

		slog.Info("seeder completed", "name", s.Name())
	}
	return nil
}
