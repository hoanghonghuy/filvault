package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"filvault/internal/app"
	"filvault/internal/platform/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	engine, err := app.New(cfg, pool)
	if err != nil {
		slog.Error("object store", "err", err)
		os.Exit(1)
	}

	slog.Info("api listening", "addr", cfg.HTTPAddr)
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		slog.Error("http", "err", err)
		os.Exit(1)
	}
}
