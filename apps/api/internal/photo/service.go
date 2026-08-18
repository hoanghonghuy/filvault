package photo

import (
	"context"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/config"
	"filvault/internal/platform/objectstore"
)

type Service struct {
	repo    Repository
	objects objectstore.ObjectStore
	now     func() time.Time
}

func NewService(repo Repository, objects objectstore.ObjectStore) *Service {
	return &Service{repo: repo, objects: objects, now: time.Now}
}

func (s *Service) Timeline(ctx context.Context, ownerID string, before *time.Time, limit int) (Timeline, error) {
	if limit <= 0 {
		limit = DefaultTimelineLimit
	}
	if limit > 100 {
		limit = 100
	}
	items, err := s.repo.ListTimelineFiles(ctx, ownerID, before, limit+1)
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

func (s *Service) CreateAlbum(ctx context.Context, ownerID, name string) (Album, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Album{}, apperr.Validation
	}
	now := s.now().UTC()
	a := Album{
		ID:        auth.NewID(),
		OwnerID:   ownerID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateAlbum(ctx, a); err != nil {
		return Album{}, err
	}
	return a, nil
}

func (s *Service) ListAlbums(ctx context.Context, ownerID string) ([]Album, error) {
	return s.repo.ListAlbums(ctx, ownerID)
}

func (s *Service) GetAlbum(ctx context.Context, ownerID, albumID string) (AlbumDetail, error) {
	a, err := s.repo.GetAlbum(ctx, ownerID, albumID)
	if err != nil {
		return AlbumDetail{}, err
	}
	if a == nil {
		return AlbumDetail{}, apperr.NotFound
	}
	items, err := s.repo.ListAlbumItems(ctx, ownerID, albumID)
	if err != nil {
		return AlbumDetail{}, err
	}
	if err := s.attachThumbnails(ctx, ownerID, items); err != nil {
		return AlbumDetail{}, err
	}
	return AlbumDetail{Album: *a, Items: items}, nil
}

func (s *Service) attachThumbnails(ctx context.Context, ownerID string, items []TimelineItem) error {
	if s.objects == nil || len(items) == 0 {
		return nil
	}
	prefs, err := s.repo.GetThumbnailPrefs(ctx, ownerID)
	if err != nil {
		return err
	}
	for i := range items {
		isImage := strings.HasPrefix(items[i].MimeType, "image/")
		isVideo := strings.HasPrefix(items[i].MimeType, "video/")
		if (isImage && !prefs.ImageEnabled) || (isVideo && !prefs.VideoEnabled) {
			continue
		}
		url, err := s.objects.CreateDownloadURL(ctx, items[i].ObjectKey, objectstore.DownloadOptions{
			Expires: config.ThumbnailPresignTTL,
		})
		if err != nil {
			return err
		}
		items[i].ThumbnailURL = url.URL
	}
	return nil
}

func (s *Service) PatchAlbum(ctx context.Context, ownerID, albumID, name string) (Album, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Album{}, apperr.Validation
	}
	a, err := s.repo.GetAlbum(ctx, ownerID, albumID)
	if err != nil {
		return Album{}, err
	}
	if a == nil {
		return Album{}, apperr.NotFound
	}
	a.Name = name
	a.UpdatedAt = s.now().UTC()
	if err := s.repo.UpdateAlbum(ctx, *a); err != nil {
		return Album{}, err
	}
	return *a, nil
}

func (s *Service) DeleteAlbum(ctx context.Context, ownerID, albumID string) error {
	a, err := s.repo.GetAlbum(ctx, ownerID, albumID)
	if err != nil {
		return err
	}
	if a == nil {
		return apperr.NotFound
	}
	return s.repo.DeleteAlbum(ctx, ownerID, albumID)
}

func (s *Service) AddAlbumItems(ctx context.Context, ownerID, albumID string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return apperr.Validation
	}
	a, err := s.repo.GetAlbum(ctx, ownerID, albumID)
	if err != nil {
		return err
	}
	if a == nil {
		return apperr.NotFound
	}
	now := s.now().UTC()
	for i, fileID := range fileIDs {
		if _, err := s.repo.GetPhotoFile(ctx, ownerID, fileID); err != nil {
			return err
		}
		if err := s.repo.AddAlbumItem(ctx, albumID, fileID, i, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RemoveAlbumItem(ctx context.Context, ownerID, albumID, fileID string) error {
	a, err := s.repo.GetAlbum(ctx, ownerID, albumID)
	if err != nil {
		return err
	}
	if a == nil {
		return apperr.NotFound
	}
	return s.repo.RemoveAlbumItem(ctx, albumID, fileID)
}

func groupByUTCDate(items []TimelineItem) []TimelineGroup {
	if len(items) == 0 {
		return []TimelineGroup{}
	}
	var groups []TimelineGroup
	currentDate := items[0].CreatedAt.UTC().Format("2006-01-02")
	current := TimelineGroup{Date: currentDate, Items: []TimelineItem{items[0]}}
	for _, item := range items[1:] {
		d := item.CreatedAt.UTC().Format("2006-01-02")
		if d != currentDate {
			groups = append(groups, current)
			currentDate = d
			current = TimelineGroup{Date: d, Items: []TimelineItem{item}}
			continue
		}
		current.Items = append(current.Items, item)
	}
	groups = append(groups, current)
	return groups
}
