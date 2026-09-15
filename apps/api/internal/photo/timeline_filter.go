package photo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

func ParseTimelineCursor(raw string) (*TimelineCursor, error) {
	if raw == "" {
		return nil, nil
	}
	sep := strings.LastIndex(raw, "|")
	if sep <= 0 || sep == len(raw)-1 {
		return nil, errors.New("invalid timeline cursor")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, raw[:sep])
	if err != nil {
		return nil, errors.New("invalid timeline cursor")
	}
	id := raw[sep+1:]
	if _, err := ulid.ParseStrict(id); err != nil {
		return nil, errors.New("invalid timeline cursor")
	}
	return &TimelineCursor{CreatedAt: createdAt, ID: id}, nil
}

func FormatTimelineCursor(item TimelineItem) string {
	return item.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + item.ID
}

// TimelineByType filters before applying the composite cursor/limit so pagination
// remains truthful even when matching media is sparse or timestamps collide.
func (s *Service) TimelineByType(ctx context.Context, ownerID string, before *TimelineCursor, limit int, mediaType string) (Timeline, error) {
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))
	if mediaType != "image" && mediaType != "video" {
		return Timeline{}, errors.New("unsupported timeline media type")
	}
	if limit <= 0 {
		limit = DefaultTimelineLimit
	}
	if limit > 100 {
		limit = 100
	}

	items, err := s.repo.ListTimelineFilesByType(ctx, ownerID, before, limit+1, mediaType)
	if err != nil {
		return Timeline{}, err
	}
	var nextBefore string
	if len(items) > limit {
		items = items[:limit]
		nextBefore = FormatTimelineCursor(items[len(items)-1])
	}
	if err := s.attachThumbnails(ctx, ownerID, items); err != nil {
		return Timeline{}, err
	}
	return Timeline{Groups: groupByUTCDate(items), NextBefore: nextBefore}, nil
}
