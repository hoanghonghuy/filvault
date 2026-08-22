package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"filvault/internal/platform/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrate_CreatesPhase1Tables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	pool := openTestPool(t, ctx)
	t.Cleanup(pool.Close)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	assertPhase1Tables(t, ctx, pool, true)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate second time: %v", err)
	}
	assertPhase1Tables(t, ctx, pool, true)
}

func TestMigrate_DownRemovesPhase1Tables(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	pool := openTestPool(t, ctx)
	t.Cleanup(pool.Close)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown phase 4: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown phase 3: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown phase 2: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown phase 1: %v", err)
	}
	assertPhase1Tables(t, ctx, pool, false)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate after down: %v", err)
	}
	assertPhase1Tables(t, ctx, pool, true)
}

func assertPhase1Tables(t *testing.T, ctx context.Context, pool *pgxpool.Pool, wantExist bool) {
	t.Helper()

	want := []string{
		"users",
		"refresh_tokens",
		"folders",
		"files",
		"albums",
		"album_items",
	}

	for _, table := range want {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT to_regclass($1) IS NOT NULL
		`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if exists != wantExist {
			t.Errorf("table %s exists=%v, want %v", table, exists, wantExist)
		}
	}
}

func openTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("FILVAULT_DATABASE_URL")
	if url == "" {
		t.Fatal("FILVAULT_DATABASE_URL is required")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("pgxpool.ParseConfig: %v", err)
	}

	// Isolate this test in its own schema so MigrateDown does not drop the
	// shared `public` tables that other packages rely on when tests run in
	// parallel against the same database.
	schema := "schema_test_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS `+schema); err != nil {
			return err
		}
		_, err := conn.Exec(ctx, `SET search_path TO `+schema)
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("pgxpool.NewWithConfig: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return pool
}
