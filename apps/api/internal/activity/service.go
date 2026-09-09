package activity

import (
	"context"
	"log/slog"
	"time"

	"filvault/internal/ids"
)

// Recorder writes events. Safe to call from anywhere; errors are logged and
// swallowed so activity logging never breaks the main flow (fail-open).
type Recorder struct {
	repo Repository
	now  func() time.Time
}

func NewRecorder(repo Repository) *Recorder {
	return &Recorder{repo: repo, now: time.Now}
}

// Record stores one event; errors are logged and swallowed so activity
// logging never breaks the main flow (fail-open).
func (r *Recorder) Record(ctx context.Context, ownerID, eventType, targetName string) {
	if ownerID == "" {
		return
	}
	err := r.repo.Append(ctx, Event{
		ID:         ids.New(),
		OwnerID:    ownerID,
		Type:       eventType,
		TargetName: targetName,
		CreatedAt:  r.now().UTC(),
	})
	if err != nil {
		slog.Warn("activity record", "type", eventType, "err", err)
	}
}

// Service serves the owner's activity feed.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

type Page struct {
	Events     []Event
	NextBefore *time.Time
}

// List returns one page of the feed. limit is clamped to [1, MaxListLimit].
func (s *Service) List(ctx context.Context, ownerID, before string, limit int) (*Page, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	var beforeAt time.Time
	if before != "" {
		parsed, err := parseCursor(before)
		if err != nil {
			return nil, errInvalidBefore
		}
		beforeAt = parsed
	} else {
		beforeAt = s.now().UTC().Add(time.Millisecond)
	}

	events, err := s.repo.List(ctx, ownerID, beforeAt, limit+1)
	if err != nil {
		return nil, err
	}
	page := &Page{Events: events}
	if len(events) > limit {
		last := events[limit-1]
		page.Events = events[:limit]
		next := last.CreatedAt
		page.NextBefore = &next
	}
	return page, nil
}
