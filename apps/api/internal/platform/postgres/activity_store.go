package postgres

import (
	"context"
	"time"

	"filvault/internal/activity"
)

func (s *Store) AppendActivityEvent(ctx context.Context, event activity.Event) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO activity_events (id, owner_id, type, target_name, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, event.ID, event.OwnerID, event.Type, event.TargetName, event.CreatedAt)
	return err
}

func (s *Store) ListActivityEvents(ctx context.Context, ownerID string, before time.Time, limit int) ([]activity.Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, type, target_name, created_at
		FROM activity_events
		WHERE owner_id = $1 AND created_at < $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`, ownerID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]activity.Event, 0, limit)
	for rows.Next() {
		var ev activity.Event
		if err := rows.Scan(&ev.ID, &ev.Type, &ev.TargetName, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (s *Store) DeleteActivityEventsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM activity_events WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

type activityRepo struct {
	store *Store
}

func NewActivityRepository(store *Store) activity.Repository {
	return activityRepo{store: store}
}

func (r activityRepo) Append(ctx context.Context, event activity.Event) error {
	return r.store.AppendActivityEvent(ctx, event)
}

func (r activityRepo) List(ctx context.Context, ownerID string, before time.Time, limit int) ([]activity.Event, error) {
	return r.store.ListActivityEvents(ctx, ownerID, before, limit)
}

func (r activityRepo) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.store.DeleteActivityEventsOlderThan(ctx, cutoff)
}

var _ activity.Repository = activityRepo{}
