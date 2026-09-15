// Package sqlite provides SQLite implementations of the repository interfaces.
//
// All SQL is parameterized — no string interpolation of user data.
// Foreign keys are enabled on every connection.
// WAL mode is used for better concurrent read performance.
//
// The Open function also runs database migrations, creating the schema if needed.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver — CGo
	"github.com/rs/zerolog/log"
)

//go:embed migrations
var migrationFS embed.FS

// DB wraps a *sql.DB with helpers for this package.
type DB struct {
	db *sql.DB
}

// Open opens the SQLite database at the given path, creates the directory if needed,
// configures pragmas, and runs outstanding migrations.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists (skip for in-memory databases).
	if dbPath != ":memory:" && !strings.HasPrefix(dbPath, "file::memory:") {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("create database directory %q: %w", dir, err)
		}
	}

	dsn := dbPath + "?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000"
	if dbPath == ":memory:" {
		// Shared in-memory DB so the connection pool sees the same schema.
		dsn = "file:timeline-memdb?mode=memory&cache=shared&_foreign_keys=on&_busy_timeout=5000"
	}
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite3 %q: %w", dbPath, err)
	}

	// SQLite supports limited concurrent writes; keep a small pool.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite3: %w", err)
	}

	wrapped := &DB{db: sqlDB}
	if err := wrapped.migrate(); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("Database opened and migrations applied")
	return wrapped, nil
}

// migrate reads migration SQL files from the embedded filesystem and applies
// any that have not yet been recorded in schema_migrations.
func (d *DB) migrate() error {
	// Ensure the schema_migrations table exists before querying it.
	bootstrapSQL := `
		PRAGMA foreign_keys=ON;
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT NOT NULL PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		);`
	if _, err := d.db.Exec(bootstrapSQL); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}

	// Read applied migration versions.
	rows, err := d.db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// List migration files.
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, fname := range files {
		version := strings.TrimSuffix(fname, ".sql")
		if applied[version] {
			continue
		}

		sql, err := migrationFS.ReadFile("migrations/" + fname)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", fname, err)
		}

		log.Info().Str("migration", version).Msg("Applying migration")
		if _, err := d.db.Exec(string(sql)); err != nil {
			return fmt.Errorf("apply migration %q: %w", version, err)
		}

		// Record migration (INSERT OR IGNORE in case the SQL file already did it).
		_, err = d.db.Exec(
			`INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)`, version)
		if err != nil {
			return fmt.Errorf("record migration %q: %w", version, err)
		}
		log.Info().Str("migration", version).Msg("Migration applied")
	}
	return nil
}

// Ping checks that the database connection is alive.
func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// DB returns the underlying *sql.DB for use by repository implementations.
func (d *DB) DB() *sql.DB {
	return d.db
}
