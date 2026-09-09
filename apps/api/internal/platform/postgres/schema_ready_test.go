package postgres_test

import (
	"context"
	"testing"
	"time"

	"filvault/internal/platform/postgres"
)

func TestSchemaReady_FailsWhenMigrationsAreNotApplied(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	pool := openTestPool(t, ctx)
	t.Cleanup(pool.Close)

	if err := postgres.SchemaReady(ctx, pool); err == nil {
		t.Fatal("expected schema not ready before migrate")
	}
}

func TestSchemaReady_PassesAfterMigrate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	pool := openTestPool(t, ctx)
	t.Cleanup(pool.Close)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := postgres.SchemaReady(ctx, pool); err != nil {
		t.Fatalf("schema ready: %v", err)
	}
}
