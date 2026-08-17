package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"filvault/internal/platform/postgres"

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
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
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

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return pool
}
