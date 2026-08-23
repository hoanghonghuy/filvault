package activity

import (
	"context"
	"time"
)

// Repository stores and queries activity events.
type Repository interface {
	// Append inserts one event. Failures are logged, never fatal (fail-open).
	Append(ctx context.Context, event Event) error
	// List returns the owner's events newest-first. before=="" starts at now.
	List(ctx context.Context, ownerID string, before time.Time, limit int) ([]Event, error)
	// DeleteOlderThan purges expired events (retention sweep).
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}
