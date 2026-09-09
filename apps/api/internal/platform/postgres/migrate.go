package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"filvault/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createSchemaMigrations = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// Migrate applies all pending *.up.sql files in version order.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	if err := execSQL(ctx, pool, createSchemaMigrations); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migration connection: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(hashtext('filvault:migrations'))`); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock(hashtext('filvault:migrations'))`)

	versions, err := listVersions()
	if err != nil {
		return err
	}

	appliedRows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("list applied: %w", err)
	}
	applied := map[string]bool{}
	for appliedRows.Next() {
		var version string
		if err := appliedRows.Scan(&version); err != nil {
			appliedRows.Close()
			return err
		}
		applied[version] = true
	}
	if err := appliedRows.Err(); err != nil {
		appliedRows.Close()
		return err
	}
	appliedRows.Close()

	for _, version := range versions {
		if applied[version] {
			continue
		}
		sql, err := fs.ReadFile(migrations.FS, version+".up.sql")
		if err != nil {
			return fmt.Errorf("read %s.up.sql: %w", version, err)
		}
		if err := execSQLConn(ctx, conn, string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", version, err)
		}
		if _, err := conn.Exec(ctx, `
			INSERT INTO schema_migrations (version) VALUES ($1)
		`, version); err != nil {
			return fmt.Errorf("record %s: %w", version, err)
		}
	}
	return nil
}

// SchemaReady reports whether every embedded migration version has been applied.
func SchemaReady(ctx context.Context, pool *pgxpool.Pool) error {
	expected, err := listVersions()
	if err != nil {
		return err
	}
	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return fmt.Errorf("schema not ready: %w", err)
	}
	for _, version := range expected {
		if !applied[version] {
			return fmt.Errorf("schema not ready: missing migration %s", version)
		}
	}
	return nil
}

// MigrateDown rolls back the latest applied version.
func MigrateDown(ctx context.Context, pool *pgxpool.Pool) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("migration connection: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(hashtext('filvault:migrations'))`); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	defer conn.Exec(ctx, `SELECT pg_advisory_unlock(hashtext('filvault:migrations'))`)

	var version string
	err = conn.QueryRow(ctx, `
		SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1
	`).Scan(&version)
	if err != nil {
		return fmt.Errorf("latest version: %w", err)
	}

	sql, err := fs.ReadFile(migrations.FS, version+".down.sql")
	if err != nil {
		return fmt.Errorf("read %s.down.sql: %w", version, err)
	}
	if err := execSQLConn(ctx, conn, string(sql)); err != nil {
		return fmt.Errorf("rollback %s: %w", version, err)
	}
	if _, err := conn.Exec(ctx, `
		DELETE FROM schema_migrations WHERE version = $1
	`, version); err != nil {
		return fmt.Errorf("unrecord %s: %w", version, err)
	}
	return nil
}

func listVersions() ([]string, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	seen := map[string]struct{}{}
	var versions []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		version := strings.TrimSuffix(name, ".up.sql")
		if _, ok := seen[version]; ok {
			continue
		}
		seen[version] = struct{}{}
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return versions, nil
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied: %w", err)
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func execSQL(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return execSQLConn(ctx, conn, sql)
}

func execSQLConn(ctx context.Context, conn *pgxpool.Conn, sql string) error {
	_, err := conn.Conn().PgConn().Exec(ctx, sql).ReadAll()
	return err
}
