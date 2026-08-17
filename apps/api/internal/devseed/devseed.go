package devseed

import (
	"context"
	"fmt"
	"time"

	"filnest/internal/auth"
	"filnest/internal/platform/config"
	"filnest/internal/platform/mailer"
	"filnest/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	DefaultEmail       = "dev@filnest.com"
	DefaultPassword    = "Dev1234@"
	DefaultDisplayName = "Filnest Dev"
	DefaultInviteCode  = "dev-invite"
)

type Options struct {
	Email       string
	Password    string
	DisplayName string
	InviteCode  string
}

type Result struct {
	Email       string
	Password    string
	DisplayName string
	InviteCode  string
	Created     bool
	Verified    bool
}

func (o Options) withDefaults() Options {
	if o.Email == "" {
		o.Email = DefaultEmail
	}
	if o.Password == "" {
		o.Password = DefaultPassword
	}
	if o.DisplayName == "" {
		o.DisplayName = DefaultDisplayName
	}
	if o.InviteCode == "" {
		o.InviteCode = DefaultInviteCode
	}
	return o
}

// EnsureDevUser creates a verified dev account or ensures an existing one is verified.
func EnsureDevUser(ctx context.Context, pool *pgxpool.Pool, opts Options) (Result, error) {
	opts = opts.withDefaults()
	store := postgres.NewStore(pool)

	existing, err := store.GetUserByEmail(ctx, opts.Email)
	if err != nil {
		return Result{}, err
	}

	now := time.Now().UTC()
	if existing != nil {
		if !existing.EmailVerified() {
			if err := store.MarkEmailVerified(ctx, existing.ID, now); err != nil {
				return Result{}, err
			}
		}
		return Result{
			Email:       opts.Email,
			Password:    opts.Password,
			DisplayName: existing.DisplayName,
			InviteCode:  opts.InviteCode,
			Created:     false,
			Verified:    true,
		}, nil
	}

	cfg := config.Config{
		InviteCode:                opts.InviteCode,
		DefaultTrashAutoDelete:    false,
		DefaultTrashRetentionDays: config.DefaultTrashRetentionDays,
	}
	mem := mailer.NewMemory()
	svc := auth.NewService(cfg, store, auth.NewTokens("seed-unused"), mem)

	if _, err := svc.Register(ctx, opts.Email, opts.Password, opts.DisplayName, opts.InviteCode); err != nil {
		return Result{}, fmt.Errorf("register: %w", err)
	}
	code := mem.LastCode(opts.Email)
	if code == "" {
		return Result{}, fmt.Errorf("verification code not sent")
	}
	if err := svc.VerifyEmail(ctx, opts.Email, code); err != nil {
		return Result{}, fmt.Errorf("verify: %w", err)
	}

	return Result{
		Email:       opts.Email,
		Password:    opts.Password,
		DisplayName: opts.DisplayName,
		InviteCode:  opts.InviteCode,
		Created:     true,
		Verified:    true,
	}, nil
}
