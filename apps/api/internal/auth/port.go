package auth

import (
	"context"
	"time"

	"filnest/internal/user"
)

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Repository interface {
	CreateUser(ctx context.Context, u user.User) error
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
	SetVerificationCode(ctx context.Context, userID, codeHash string, expiresAt time.Time) error
	MarkEmailVerified(ctx context.Context, userID string, at time.Time) error
	InsertRefreshToken(ctx context.Context, tok RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, hash string, at time.Time) error
}
