package trash_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"filvault/internal/file"
	"filvault/internal/platform/objectstore"
	"filvault/internal/platform/postgres"
	"filvault/internal/trash"
)

func TestWorker_CompletesPurgeJobAfterDatabasePurge(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "worker.jpg")

	if code, body := deleteAuth(t, engine, "/api/v1/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("trash status=%d body=%s", code, body)
	}
	if code, body := deleteAuth(t, engine, "/api/v1/trash/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("purge status=%d body=%s", code, body)
	}

	repo := postgres.NewTrashRepository(postgres.NewStore(pool))
	worker := trash.NewWorker(repo, objs, "test-worker")
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("worker run: %v", err)
	}

	var completed bool
	if err := pool.QueryRow(context.Background(), `
		SELECT completed_at IS NOT NULL
		FROM file_purge_jobs
		WHERE file_id = $1
	`, fileID).Scan(&completed); err != nil {
		t.Fatalf("purge job: %v", err)
	}
	if !completed {
		t.Fatal("purge job was not completed")
	}
}

func TestWorker_CleansExpiredPendingUploads(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	_, body := postAuth(t, engine, "/api/v1/chat/conversations", token, map[string]any{"title": "Uploads"})
	var conversation struct {
		ID string `json:"id"`
	}
	decodeJSON(t, body, &conversation)

	code, body := postAuth(t, engine, "/api/v1/chat/attachments/upload-sessions", token, map[string]any{
		"conversationId": conversation.ID,
		"name":           "expired.pdf",
		"size":           8,
		"contentType":    "application/pdf",
	})
	if code != http.StatusCreated {
		t.Fatalf("session status=%d body=%s", code, body)
	}
	var session struct {
		FileID string `json:"fileId"`
	}
	decodeJSON(t, body, &session)

	if _, err := pool.Exec(context.Background(), `
		UPDATE files SET upload_expires_at = $2 WHERE id = $1
	`, session.FileID, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatalf("expire upload: %v", err)
	}

	repo := postgres.NewTrashRepository(postgres.NewStore(pool))
	worker := trash.NewWorker(repo, objs, "test-worker")
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("worker run: %v", err)
	}

	var fileCount, completedCount int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM files WHERE id = $1`, session.FileID).Scan(&fileCount); err != nil {
		t.Fatalf("pending file: %v", err)
	}
	if err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM file_upload_cleanup_jobs
		WHERE file_id = $1 AND completed_at IS NOT NULL
	`, session.FileID).Scan(&completedCount); err != nil {
		t.Fatalf("cleanup job: %v", err)
	}
	if fileCount != 0 || completedCount != 1 {
		t.Fatalf("cleanup fileCount=%d completedJobs=%d", fileCount, completedCount)
	}
}

func TestWorker_DoesNotCompleteLostPurgeLease(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "lease.jpg")

	if code, body := deleteAuth(t, engine, "/api/v1/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("trash status=%d body=%s", code, body)
	}
	if code, body := deleteAuth(t, engine, "/api/v1/trash/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("purge status=%d body=%s", code, body)
	}

	var jobID string
	if err := pool.QueryRow(context.Background(), `
		SELECT id FROM file_purge_jobs WHERE file_id = $1
	`, fileID).Scan(&jobID); err != nil {
		t.Fatalf("purge job: %v", err)
	}
	repo := postgres.NewTrashRepository(postgres.NewStore(pool))
	if err := repo.CompletePurgeJob(context.Background(), jobID, "another-worker"); err == nil {
		t.Fatal("lost lease should not complete purge job")
	}
}

func TestWorker_CompletesPurgeWhenObjectIsAlreadyMissing(t *testing.T) {
	engine, mem, objs, pool := newEngine(t)
	token := registerVerified(t, engine, mem, uniqueEmail())
	fileID := uploadReady(t, engine, objs, token, "missing.jpg")

	if code, body := deleteAuth(t, engine, "/api/v1/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("trash status=%d body=%s", code, body)
	}
	if code, body := deleteAuth(t, engine, "/api/v1/trash/files/"+fileID, token); code != http.StatusNoContent {
		t.Fatalf("purge status=%d body=%s", code, body)
	}

	var me struct {
		ID string `json:"id"`
	}
	code, body := getAuth(t, engine, "/api/v1/users/me", token)
	if code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", code, body)
	}
	decodeJSON(t, body, &me)
	_ = objs.Delete(context.Background(), file.ObjectKey(me.ID, fileID))
	repo := postgres.NewTrashRepository(postgres.NewStore(pool))
	worker := trash.NewWorker(repo, objs, "test-worker")
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("worker run: %v", err)
	}

	var completed bool
	if err := pool.QueryRow(context.Background(), `
		SELECT completed_at IS NOT NULL
		FROM file_purge_jobs
		WHERE file_id = $1
	`, fileID).Scan(&completed); err != nil {
		t.Fatalf("purge job: %v", err)
	}
	if !completed {
		t.Fatal("missing object should be treated as successful purge")
	}
}

var _ objectstore.ObjectStore = (*objectstore.Memory)(nil)
