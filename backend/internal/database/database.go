// Package database opens the PostgreSQL connection pool and applies embedded
// versioned migrations.
package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/abudemy/backend/migrations"
)

// Options tunes the connection pool and migration set.
type Options struct {
	MaxConns int
	// SeedDemo applies the demo-data migrations (see IsDemoMigration).
	SeedDemo bool
}

// IsDemoMigration reports whether a migration only loads demonstration data
// (accounts with publicly documented passwords). They are skipped unless
// Options.SeedDemo is set, so a production database never receives them.
func IsDemoMigration(name string) bool {
	return strings.Contains(name, "_seed_demo_") ||
		strings.Contains(name, "_seed_web_demo_") ||
		strings.Contains(name, "_demo_data_")
}

// Connect opens a pgx connection pool and verifies connectivity. It retries
// for a short while because serverless Postgres (e.g. Neon) may need a few
// seconds to wake up from scale-to-zero.
func Connect(ctx context.Context, databaseURL string, maxConns int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if maxConns <= 0 {
		maxConns = 10
	}
	cfg.MaxConns = int32(maxConns)
	// Keep no idle connection open permanently so a scale-to-zero database can
	// suspend when the API is idle.
	cfg.MinConns = 0
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	const attempts = 5
	var pingErr error
	for i := 1; i <= attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		pingErr = pool.Ping(pingCtx)
		cancel()
		if pingErr == nil {
			return pool, nil
		}
		if ctx.Err() != nil {
			break
		}
		slog.Warn("database not reachable yet", "attempt", i, "of", attempts, "error", pingErr)
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(i) * 2 * time.Second):
		}
	}
	pool.Close()
	return nil, fmt.Errorf("ping database: %w", pingErr)
}

// Migrate applies every embedded migration that has not been applied yet.
// Migrations run in lexical filename order; each file is applied atomically.
// Demo-data migrations are skipped unless opts.SeedDemo is true.
func Migrate(ctx context.Context, pool *pgxpool.Pool, opts Options) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if IsDemoMigration(e.Name()) && !opts.SeedDemo {
			slog.Info("skipping demo-data migration", "migration", e.Name())
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		applied, err := isApplied(ctx, pool, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := applyMigration(ctx, pool, name, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, version, body string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Multi-statement DDL runs with the simple query protocol (no arguments).
	if _, err := tx.Exec(ctx, body); err != nil {
		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}
