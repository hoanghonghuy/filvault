package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"filvault/internal/activity"
	"filvault/internal/app"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"
	"filvault/internal/trash"

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

	if err := postgres.Migrate(ctx, pool); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	engine, err := app.New(cfg, pool)
	if err != nil {
		slog.Error("object store", "err", err)
		os.Exit(1)
	}

	tickerCtx, tickerCancel := context.WithCancel(context.Background())
	defer tickerCancel()
	trashSvc := app.NewTrashService(pool, mustObjectStore(cfg))
	trash.StartTicker(tickerCtx, trashSvc, trash.DefaultTickerInterval)
	activity.StartTicker(tickerCtx, app.NewActivityRepository(pool), activity.DefaultTickerInterval)

	slog.Info("api listening", "addr", cfg.HTTPAddr)
	if err := engine.Run(cfg.HTTPAddr); err != nil {
		slog.Error("http", "err", err)
		os.Exit(1)
	}
}

func mustObjectStore(cfg config.Config) objectstore.ObjectStore {
	obj, err := objectstore.NewFromConfig(cfg)
	if err != nil {
		slog.Error("object store", "err", err)
		os.Exit(1)
	}
	return obj
}
