package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
)

const passwordResetTTL = 30 * time.Minute
const passwordResetTokenBytes = 32

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") {
		return apperr.Validation
	}

	// Generate before the lookup so known/unknown requests share the crypto work.
	token, err := newPasswordResetToken()
	if err != nil {
		return err
	}

	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u == nil {
		return nil
	}

	expiresAt := s.now().UTC().Add(passwordResetTTL)
	if err := s.repo.SetPasswordResetToken(ctx, u.ID, hashPasswordResetToken(token), expiresAt); err != nil {
		return err
	}

	// Do not let SMTP/provider failures become an account-enumeration oracle. The
	// token remains short-lived and a later request replaces it.
	_ = s.mailer.Send(ctx, mailer.Message{
		To:      u.Email,
		Subject: "Filvault password reset",
		Body:    token,
	})
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if !validPasswordResetToken(token) || len(newPassword) < config.MinPasswordLength {
		return apperr.Validation
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = s.repo.ConsumePasswordReset(ctx, hashPasswordResetToken(token), passwordHash, s.now().UTC())
	return err
}

func newPasswordResetToken() (string, error) {
	buf := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func validPasswordResetToken(token string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(decoded) == passwordResetTokenBytes
}

func hashPasswordResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
