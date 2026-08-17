package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"filnest/internal/devseed"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	url := os.Getenv("FILNEST_DATABASE_URL")
	if url == "" {
		fatal("FILNEST_DATABASE_URL is required")
	}

	opts := devseed.Options{
		Email:       os.Getenv("FILNEST_SEED_EMAIL"),
		Password:    os.Getenv("FILNEST_SEED_PASSWORD"),
		DisplayName: os.Getenv("FILNEST_SEED_DISPLAY_NAME"),
		InviteCode:  getenv("FILNEST_INVITE_CODE", devseed.DefaultInviteCode),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fatal("connect: %v", err)
	}
	defer pool.Close()

	result, err := devseed.EnsureDevUser(ctx, pool, opts)
	if err != nil {
		fatal("seed: %v", err)
	}

	status := "created"
	if !result.Created {
		status = "already exists (verified)"
	}

	fmt.Printf(`dev user ready (%s)
  email:       %s
  password:    %s
  displayName: %s
  inviteCode:  %s (chỉ cần khi đăng ký tay trên UI)
`, status, result.Email, result.Password, result.DisplayName, result.InviteCode)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
