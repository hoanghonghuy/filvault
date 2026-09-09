package postgres

import (
	"context"
	"errors"
	"time"

	"filvault/internal/vault"

	"github.com/jackc/pgx/v5"
)

type vaultRepo struct {
	store *Store
}

func NewVaultRepository(store *Store) vault.Repository {
	return &vaultRepo{store: store}
}

func (r *vaultRepo) GetVault(ctx context.Context, userID string) (*vault.Vault, error) {
	var v vault.Vault
	err := r.store.pool.QueryRow(ctx, `
		SELECT user_id, pin_hash, created_at, updated_at
		FROM user_vaults
		WHERE user_id = $1
	`, userID).Scan(&v.UserID, &v.PinHash, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *vaultRepo) CreateVault(ctx context.Context, userID, pinHash string, now time.Time) error {
	_, err := r.store.pool.Exec(ctx, `
		INSERT INTO user_vaults (user_id, pin_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, pinHash, now, now)
	return err
}

func (r *vaultRepo) UpdateVaultPin(ctx context.Context, userID, pinHash string, now time.Time) error {
	_, err := r.store.pool.Exec(ctx, `
		INSERT INTO user_vaults (user_id, pin_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET pin_hash = EXCLUDED.pin_hash, updated_at = EXCLUDED.updated_at
	`, userID, pinHash, now)
	return err
}

func (r *vaultRepo) ListVaultFiles(ctx context.Context, userID string) ([]vault.VaultFile, error) {
	rows, err := r.store.pool.Query(ctx, `
		SELECT id, name, mime_type, size_bytes, created_at, updated_at
		FROM files
		WHERE owner_id = $1 AND deleted_at IS NULL AND is_vault = TRUE AND status = 'READY'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []vault.VaultFile
	for rows.Next() {
		var f vault.VaultFile
		if err := rows.Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeBytes, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	if files == nil {
		files = []vault.VaultFile{}
	}
	return files, rows.Err()
}

func (r *vaultRepo) MoveFilesToVault(ctx context.Context, userID string, fileIDs []string) error {
	_, err := r.store.pool.Exec(ctx, `
		UPDATE files
		SET is_vault = TRUE, folder_id = NULL, updated_at = now()
		WHERE owner_id = $1 AND id = ANY($2) AND deleted_at IS NULL
	`, userID, fileIDs)
	if err != nil {
		return err
	}

	_, _ = r.store.pool.Exec(ctx, `
		DELETE FROM favorites WHERE owner_id = $1 AND file_id = ANY($2)
	`, userID, fileIDs)

	return nil
}

func (r *vaultRepo) MoveFilesFromVault(ctx context.Context, userID string, fileIDs []string) error {
	_, err := r.store.pool.Exec(ctx, `
		UPDATE files
		SET is_vault = FALSE, folder_id = NULL, updated_at = now()
		WHERE owner_id = $1 AND id = ANY($2) AND deleted_at IS NULL
	`, userID, fileIDs)
	return err
}
