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
	assertChatAttachmentLifecycle(t, ctx, pool)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate second time: %v", err)
	}
	assertPhase1Tables(t, ctx, pool, true)
}

func assertChatAttachmentLifecycle(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var hasSource, hasCompletedMessage, hasDisplayName, hasMimeType, hasSizeBytes, hasNullableFileID, hasPurgeJobs bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'files' AND column_name = 'source'
		)
	`).Scan(&hasSource); err != nil {
		t.Fatalf("check files.source: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'files' AND column_name = 'completed_message_id'
		)
	`).Scan(&hasCompletedMessage); err != nil {
		t.Fatalf("check files.completed_message_id: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'message_attachments' AND column_name = 'display_name'
		)
	`).Scan(&hasDisplayName); err != nil {
		t.Fatalf("check attachment snapshot: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'message_attachments' AND column_name = 'mime_type'
		)
	`).Scan(&hasMimeType); err != nil {
		t.Fatalf("check attachment mime snapshot: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'message_attachments' AND column_name = 'size_bytes'
		)
	`).Scan(&hasSizeBytes); err != nil {
		t.Fatalf("check attachment size snapshot: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT is_nullable = 'YES'
		FROM information_schema.columns
		WHERE table_name = 'message_attachments' AND column_name = 'file_id'
	`).Scan(&hasNullableFileID); err != nil {
		t.Fatalf("check nullable attachment file: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT to_regclass('file_purge_jobs') IS NOT NULL`).Scan(&hasPurgeJobs); err != nil {
		t.Fatalf("check file purge jobs: %v", err)
	}
	if !hasSource || !hasCompletedMessage || !hasDisplayName || !hasMimeType || !hasSizeBytes || !hasNullableFileID || !hasPurgeJobs {
		t.Fatalf("chat attachment lifecycle schema incomplete: source=%v completedMessage=%v displayName=%v mimeType=%v sizeBytes=%v nullableFileID=%v purgeJobs=%v", hasSource, hasCompletedMessage, hasDisplayName, hasMimeType, hasSizeBytes, hasNullableFileID, hasPurgeJobs)
	}
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
		t.Fatalf("MigrateDown activity events: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown share links: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown favorites: %v", err)
	}
	if err := postgres.MigrateDown(ctx, pool); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}
	for _, name := range []string{
		"production direct chat",
		"pending upload cleanup",
		"pending upload cleanup job retention",
		"chat attachment lifecycle",
		"file name uniqueness",
		"file versions",
		"shares",
		"chat",
		"thumbnails",
	} {
		if err := postgres.MigrateDown(ctx, pool); err != nil {
			t.Fatalf("MigrateDown %s: %v", name, err)
		}
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
