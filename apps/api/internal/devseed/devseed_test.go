package devseed_test

import (
	"context"
	"os"
	"testing"
	"time"

	"filvault/internal/activity"
	"filvault/internal/auth"
	"filvault/internal/devseed"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

// nopRepo swallows activity events in tests that do not assert on them.
type nopRepo struct{}

func (nopRepo) Append(context.Context, activity.Event) error { return nil }

func (nopRepo) List(context.Context, string, time.Time, int) ([]activity.Event, error) {
	return nil, nil
}

func (nopRepo) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

func TestEnsureDevUser_CreatesVerifiedUser(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	email := uniqueEmail()

	result, err := devseed.EnsureDevUser(ctx, pool, devseed.Options{
		Email:       email,
		Password:    "filvault-dev",
		DisplayName: "Filvault Dev",
		InviteCode:  "local-test-invite-code",
	})
	if err != nil {
		t.Fatalf("EnsureDevUser: %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true on first run")
	}
	if !result.Verified {
		t.Fatal("expected Verified=true")
	}

	store := postgres.NewStore(pool)
	u, err := store.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if u == nil || !u.EmailVerified() {
		t.Fatal("user should exist and be verified")
	}

	svc := auth.NewService(config.Config{
		InviteCode:                "local-test-invite-code",
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
	}, store, auth.NewTokens("test-jwt"), mailer.NewMemory(), activity.NewRecorder(nopRepo{}))

	session, err := svc.Login(ctx, email, "filvault-dev")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !session.User.EmailVerified() {
		t.Fatal("login session user should be verified")
	}
}

func TestEnsureDevUser_Idempotent(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	email := uniqueEmail()
	opts := devseed.Options{
		Email:       email,
		Password:    "filvault-dev",
		DisplayName: "Filvault Dev",
		InviteCode:  "local-test-invite-code",
	}

	first, err := devseed.EnsureDevUser(ctx, pool, opts)
	if err != nil {
		t.Fatalf("first EnsureDevUser: %v", err)
	}
	if !first.Created {
		t.Fatal("expected Created=true on first run")
	}

	second, err := devseed.EnsureDevUser(ctx, pool, opts)
	if err != nil {
		t.Fatalf("second EnsureDevUser: %v", err)
	}
	if second.Created {
		t.Fatal("expected Created=false on second run")
	}
	if !second.Verified {
		t.Fatal("expected Verified=true on second run")
	}
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	url := os.Getenv("FILVAULT_DATABASE_URL")
	if url == "" {
		t.Fatal("FILVAULT_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func uniqueEmail() string {
	return "devseed-" + time.Now().Format("20060102150405.000000") + "@filvault.local"
}
