package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/sandevgo/tuskbot/pkg/log"
	_ "github.com/sandevgo/tuskbot/pkg/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func NewDB(ctx context.Context, dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3_vec", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log.NewGooseLoggerFromCtx(ctx))

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up failed: %w", err)
	}

	return nil
}
