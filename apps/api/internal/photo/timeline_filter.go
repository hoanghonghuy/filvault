package photo

import (
	"context"
	"errors"
	"strings"
	"time"
)

type timelineTypeRepository interface {
	ListTimelineFilesByType(ctx context.Context, ownerID string, before *time.Time, limit int, mediaType string) ([]TimelineItem, error)
}

// TimelineByType filters before applying the cursor/limit so pagination remains
// truthful even when matching media is sparse among other timeline items.
func (s *Service) TimelineByType(ctx context.Context, ownerID string, before *time.Time, limit int, mediaType string) (Timeline, error) {
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))
	if mediaType == "" {
		return s.Timeline(ctx, ownerID, before, limit)
	}
	if mediaType != "image" && mediaType != "video" {
		return Timeline{}, errors.New("unsupported timeline media type")
	}
	if limit <= 0 {
		limit = DefaultTimelineLimit
	}
	if limit > 100 {
		limit = 100
	}

	repo, ok := s.repo.(timelineTypeRepository)
	if !ok {
		return Timeline{}, errors.New("timeline repository does not support media type filtering")
	}
	items, err := repo.ListTimelineFilesByType(ctx, ownerID, before, limit+1, mediaType)
	if err != nil {
		return Timeline{}, err
	}
	var nextBefore string
	if len(items) > limit {
		items = items[:limit]
		nextBefore = items[len(items)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}
	if err := s.attachThumbnails(ctx, ownerID, items); err != nil {
		return Timeline{}, err
	}
	return Timeline{Groups: groupByUTCDate(items), NextBefore: nextBefore}, nil
}
