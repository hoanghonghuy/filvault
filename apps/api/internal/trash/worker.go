package trash

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"filvault/internal/platform/objectstore"
)

const (
	DefaultWorkerInterval = time.Minute
	DefaultWorkerLease    = 2 * time.Minute
	maxPurgeAttempts      = 8
	maxBackoff            = 6
	maxJobsPerRun         = 100
)

// Worker runs durable purge jobs outside the HTTP request lifecycle. The
// repository owns claiming and retry state; multiple workers may run safely.
type Worker struct {
	repo     Repository
	objects  ObjectStore
	workerID string
	cleanup  func(context.Context) error
}

func NewWorker(repo Repository, objects ObjectStore, workerID string) *Worker {
	return &Worker{repo: repo, objects: objects, workerID: workerID}
}

func NewWorkerWithCleanup(repo Repository, objects ObjectStore, workerID string, cleanup func(context.Context) error) *Worker {
	worker := NewWorker(repo, objects, workerID)
	worker.cleanup = cleanup
	return worker
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultWorkerInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if err := w.RunOnce(ctx); err != nil {
			slog.Error("file lifecycle worker", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	if w.repo == nil || w.objects == nil {
		return fmt.Errorf("purge worker dependencies are required")
	}
	if w.cleanup != nil {
		if err := w.cleanup(ctx); err != nil {
			return fmt.Errorf("auto cleanup: %w", err)
		}
	}
	if err := w.repo.QueueExpiredUploadCleanupJobs(ctx, time.Now().UTC()); err != nil {
		return err
	}
	for i := 0; i < maxJobsPerRun; i++ {
		job, err := w.repo.ClaimPurgeJob(ctx, w.workerID, DefaultWorkerLease)
		if err != nil {
			return err
		}
		if job == nil {
			break
		}
		if err := deleteObject(ctx, w.objects, job.ObjectKey); err != nil {
			if err := w.failPurge(ctx, job, err); err != nil {
				return err
			}
			continue
		}
		if err := w.repo.CompletePurgeJob(ctx, job.ID, w.workerID); err != nil {
			return err
		}
	}
	for i := 0; i < maxJobsPerRun; i++ {
		processed, err := w.runUploadCleanupOnce(ctx)
		if err != nil {
			return err
		}
		if !processed {
			break
		}
	}
	return nil
}

func (w *Worker) failPurge(ctx context.Context, job *PurgeJob, cause error) error {
	dead := job.Attempts >= maxPurgeAttempts
	retryAt := time.Now().UTC().Add(backoff(job.Attempts))
	if err := w.repo.FailPurgeJob(ctx, job.ID, w.workerID, cause.Error(), retryAt, dead); err != nil {
		return fmt.Errorf("record purge failure: %w (original: %v)", err, cause)
	}
	if dead {
		slog.Error("purge job dead", "jobId", job.ID, "fileId", job.FileID, "err", cause)
	}
	return nil
}

func (w *Worker) runUploadCleanupOnce(ctx context.Context) (bool, error) {
	job, err := w.repo.ClaimUploadCleanupJob(ctx, w.workerID, DefaultWorkerLease)
	if err != nil || job == nil {
		return job != nil, err
	}
	if err := deleteObject(ctx, w.objects, job.ObjectKey); err != nil {
		dead := job.Attempts >= maxPurgeAttempts
		if failErr := w.repo.FailUploadCleanupJob(ctx, job.ID, w.workerID, err.Error(), time.Now().UTC().Add(backoff(job.Attempts)), dead); failErr != nil {
			return true, fmt.Errorf("record upload cleanup failure: %w (original: %v)", failErr, err)
		}
		if dead {
			slog.Error("upload cleanup job dead", "jobId", job.ID, "fileId", job.FileID, "err", err)
		}
		return true, nil
	}
	return true, w.repo.CompleteUploadCleanupJob(ctx, job.ID, w.workerID, job.OwnerID, job.FileID)
}

func deleteObject(ctx context.Context, objects ObjectStore, key string) error {
	err := objects.Delete(ctx, key)
	if errors.Is(err, objectstore.ErrObjectNotFound) {
		return nil
	}
	return err
}

func backoff(attempts int) time.Duration {
	base := time.Duration(1<<min(attempts, maxBackoff)) * time.Minute
	return base + time.Duration(rand.Int63n(int64(30*time.Second)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
