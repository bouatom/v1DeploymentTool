package db

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("pool is required")
	}

	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, filename := range files {
		applied, err := hasMigration(ctx, pool, filename)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if err := applyMigration(ctx, pool, filename); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL
		);
	`)
	return err
}

func hasMigration(ctx context.Context, pool *pgxpool.Pool, filename string) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = $1`, filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, filename string) error {
	content, err := migrationFiles.ReadFile("migrations/" + filename)
	if err != nil {
		return err
	}

	parts := splitStatements(string(content))
	if len(parts) == 0 {
		return nil
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	for _, statement := range parts {
		if _, err := tx.Exec(ctx, statement); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`, filename, time.Now().UTC()); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func splitStatements(content string) []string {
	var statements []string
	raw := strings.Split(content, ";")
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		statements = append(statements, trimmed+";")
	}

	return statements
}
