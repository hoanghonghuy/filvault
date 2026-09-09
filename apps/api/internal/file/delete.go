package file

import (
	"context"

	"filvault/internal/activity"
	"filvault/internal/apperr"
)

func (s *Service) Delete(ctx context.Context, ownerID, fileID string) error {
	f, err := s.repo.GetByID(ctx, ownerID, fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return apperr.NotFound
	}
	if f.Status == StatusPending {
		_ = s.objects.Delete(ctx, f.ObjectKey)
		return s.repo.DeleteRow(ctx, ownerID, fileID)
	}
	if f.Status != StatusReady {
		return apperr.InvalidState
	}
	s.activity.Record(ctx, ownerID, activity.TypeFileTrashed, f.Name)
	return s.repo.SoftDelete(ctx, ownerID, fileID, s.now().UTC())
}
