package vault

import (
	"context"
	"time"
)

type Vault struct {
	UserID    string    `json:"userId"`
	PinHash   string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type VaultFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mimeType"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Status struct {
	Initialized bool `json:"initialized"`
	Unlocked    bool `json:"unlocked"`
}

type Session struct {
	Unlocked  bool      `json:"unlocked"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Repository interface {
	GetVault(ctx context.Context, userID string) (*Vault, error)
	CreateVault(ctx context.Context, userID, pinHash string, now time.Time) error
	UpdateVaultPin(ctx context.Context, userID, pinHash string, now time.Time) error
	ListVaultFiles(ctx context.Context, userID string) ([]VaultFile, error)
	MoveFilesToVault(ctx context.Context, userID string, fileIDs []string) error
	MoveFilesFromVault(ctx context.Context, userID string, fileIDs []string) error
}

type UserVerifier interface {
	VerifyPassword(ctx context.Context, userID, password string) (bool, error)
}
