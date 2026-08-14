package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"filnest/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	down := flag.Bool("down", false, "roll back the latest migration")
	flag.Parse()

	url := os.Getenv("FILNEST_DATABASE_URL")
	if url == "" {
		fatal("FILNEST_DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fatal("connect: %v", err)
	}
	defer pool.Close()

	if *down {
		if err := postgres.MigrateDown(ctx, pool); err != nil {
			fatal("migrate down: %v", err)
		}
		fmt.Println("migrate down: ok")
		return
	}

	if err := postgres.Migrate(ctx, pool); err != nil {
		fatal("migrate: %v", err)
	}
	fmt.Println("migrate: ok")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
