package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"filvault/internal/activity"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("postgres ping", "err", err)
		os.Exit(1)
	}
	objects, err := objectstore.NewFromConfig(cfg)
	if err != nil {
		slog.Error("object store", "err", err)
		os.Exit(1)
	}
	store := postgres.NewStore(pool)
	trashWorker := trash.NewWorkerWithCleanup(
		postgres.NewTrashRepository(store),
		objects,
		"worker-"+os.Getenv("HOSTNAME"),
		newTrashService(pool, objects).RunAutoCleanup,
	)
	go trashWorker.Run(ctx, trash.DefaultWorkerInterval)
	activity.StartTicker(ctx, postgres.NewActivityRepository(store), activity.DefaultTickerInterval)
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := shutdownCtx.Err(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		slog.Warn("worker shutdown", "err", err)
	}
	slog.Info("worker stopped")
}

func newTrashService(pool *pgxpool.Pool, objects objectstore.ObjectStore) *trash.Service {
	store := postgres.NewStore(pool)
	return trash.NewService(
		postgres.NewTrashRepository(store),
		postgres.NewQuotaStore(store),
		objects,
		activity.NewRecorder(postgres.NewActivityRepository(store)),
	)
}
