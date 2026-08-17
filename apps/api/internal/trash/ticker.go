package trash

import (
	"context"
	"time"
)

const DefaultTickerInterval = time.Hour

func StartTicker(ctx context.Context, svc *Service, interval time.Duration) {
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
				_ = svc.RunAutoCleanup(ctx)
			}
		}
	}()
}
