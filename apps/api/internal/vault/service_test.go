package vault_test

import (
	"context"
	"testing"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/vault"
)

type mockRepo struct {
	vault *vault.Vault
	files []vault.VaultFile
}

func (m *mockRepo) GetVault(ctx context.Context, userID string) (*vault.Vault, error) {
	if m.vault != nil && m.vault.UserID == userID {
		return m.vault, nil
	}
	return nil, nil
}

func (m *mockRepo) CreateVault(ctx context.Context, userID, pinHash string, now time.Time) error {
	m.vault = &vault.Vault{
		UserID:    userID,
		PinHash:   pinHash,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return nil
}

func (m *mockRepo) UpdateVaultPin(ctx context.Context, userID, pinHash string, now time.Time) error {
	if m.vault == nil {
		m.vault = &vault.Vault{
			UserID:    userID,
			PinHash:   pinHash,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return nil
	}
	m.vault.PinHash = pinHash
	m.vault.UpdatedAt = now
	return nil
}

func (m *mockRepo) ListVaultFiles(ctx context.Context, userID string) ([]vault.VaultFile, error) {
	return m.files, nil
}

func (m *mockRepo) MoveFilesToVault(ctx context.Context, userID string, fileIDs []string) error {
	for _, id := range fileIDs {
		m.files = append(m.files, vault.VaultFile{ID: id, Name: "file-" + id})
	}
	return nil
}

func (m *mockRepo) MoveFilesFromVault(ctx context.Context, userID string, fileIDs []string) error {
	filtered := make([]vault.VaultFile, 0, len(m.files))
	for _, f := range m.files {
		found := false
		for _, id := range fileIDs {
			if f.ID == id {
				found = true
				break
			}
		}
		if !found {
			filtered = append(filtered, f)
		}
	}
	m.files = filtered
	return nil
}

type mockVerifier struct {
	validPassword string
}

func (m *mockVerifier) VerifyPassword(ctx context.Context, userID, password string) (bool, error) {
	return password == m.validPassword, nil
}

func TestVaultService_FullFlow(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	verifier := &mockVerifier{validPassword: "SecretAccountPassword123"}
	tokens := vault.NewTokens("test-secret-key-1234567890123456", 1*time.Hour)
	svc := vault.NewService(repo, verifier, tokens)

	userID := "01M1X1Q2YTJ4RBHWK82VE2YCCN"

	// 1. Status: Not initialized
	st, err := svc.Status(ctx, userID, "")
	if err != nil {
		t.Fatalf("unexpected error on status: %v", err)
	}
	if st.Initialized || st.Unlocked {
		t.Fatalf("expected initialized=false, unlocked=false; got %+v", st)
	}

	// 2. Setup with short PIN (< 4) should fail
	_, err = svc.Setup(ctx, userID, "12")
	if err != apperr.Validation {
		t.Fatalf("expected validation error on short PIN, got %v", err)
	}

	// 3. Setup with valid PIN
	session, err := svc.Setup(ctx, userID, "123456")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if !session.Unlocked || session.Token == "" {
		t.Fatalf("expected valid session, got %+v", session)
	}

	// 4. Status with issued token should be unlocked
	st, err = svc.Status(ctx, userID, session.Token)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !st.Initialized || !st.Unlocked {
		t.Fatalf("expected initialized=true, unlocked=true; got %+v", st)
	}

	// 5. Setup again should fail with Conflict
	_, err = svc.Setup(ctx, userID, "654321")
	if err != apperr.Conflict {
		t.Fatalf("expected Conflict on duplicate setup, got %v", err)
	}

	// 6. Unlock with wrong PIN should fail
	_, err = svc.Unlock(ctx, userID, "wrongpin")
	if err == nil {
		t.Fatal("expected error on wrong PIN, got nil")
	}

	// 7. Unlock with correct PIN
	unlockedSession, err := svc.Unlock(ctx, userID, "123456")
	if err != nil {
		t.Fatalf("unlock failed: %v", err)
	}
	if !unlockedSession.Unlocked || unlockedSession.Token == "" {
		t.Fatalf("expected unlocked session, got %+v", unlockedSession)
	}

	// 8. Change PIN
	err = svc.ChangePin(ctx, userID, "wrong", "999999")
	if err == nil {
		t.Fatal("expected error on wrong current PIN, got nil")
	}
	err = svc.ChangePin(ctx, userID, "123456", "999999")
	if err != nil {
		t.Fatalf("change pin failed: %v", err)
	}

	// 9. Unlock with old PIN should fail, new PIN should succeed
	_, err = svc.Unlock(ctx, userID, "123456")
	if err == nil {
		t.Fatal("expected old PIN to fail, got nil")
	}
	_, err = svc.Unlock(ctx, userID, "999999")
	if err != nil {
		t.Fatalf("unlock with new PIN failed: %v", err)
	}

	// 10. Reset PIN with account password
	_, err = svc.ResetPin(ctx, userID, "WrongAccountPassword", "888888")
	if err == nil {
		t.Fatal("expected error on wrong account password, got nil")
	}
	resetSession, err := svc.ResetPin(ctx, userID, "SecretAccountPassword123", "888888")
	if err != nil {
		t.Fatalf("reset pin failed: %v", err)
	}
	if !resetSession.Unlocked {
		t.Fatal("expected unlocked session after reset")
	}

	// 11. Move files to vault and list
	err = svc.MoveToVault(ctx, userID, []string{"file-1", "file-2"})
	if err != nil {
		t.Fatalf("move to vault failed: %v", err)
	}
	files, err := svc.ListFiles(ctx, userID)
	if err != nil {
		t.Fatalf("list files failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	// 12. Move file out of vault
	err = svc.MoveFromVault(ctx, userID, []string{"file-1"})
	if err != nil {
		t.Fatalf("move from vault failed: %v", err)
	}
	files, err = svc.ListFiles(ctx, userID)
	if err != nil {
		t.Fatalf("list files failed: %v", err)
	}
	if len(files) != 1 || files[0].ID != "file-2" {
		t.Fatalf("expected 1 file (file-2), got %+v", files)
	}
}
