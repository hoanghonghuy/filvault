package activity

import (
	"context"
	"log/slog"
	"time"
)

// DefaultTickerInterval matches the trash cleanup cadence.
const DefaultTickerInterval = time.Hour

// StartTicker purges expired events on the given cadence until ctx ends.
func StartTicker(ctx context.Context, repo Repository, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultTickerInterval
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().UTC().AddDate(0, 0, -RetentionDays)
				n, err := repo.DeleteOlderThan(ctx, cutoff)
				if err != nil {
					slog.Warn("activity retention sweep", "err", err)
					continue
				}
				if n > 0 {
					slog.Info("activity retention sweep", "deleted", n)
				}
			}
		}
	}()
}
