package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"filnest/internal/apperr"
	"filnest/internal/platform/mailer"
)

const verificationTTL = 15 * time.Minute

func (s *Service) ResendVerification(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if email == "" {
		return apperr.Validation
	}
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u == nil || u.EmailVerified() {
		return nil
	}

	code, err := newVerificationCode()
	if err != nil {
		return err
	}
	expires := s.now().UTC().Add(verificationTTL)
	if err := s.repo.SetVerificationCode(ctx, u.ID, hashVerificationCode(code), expires); err != nil {
		return err
	}
	return s.mailer.Send(ctx, mailer.Message{
		To:      u.Email,
		Subject: "Filnest verification code",
		Body:    code,
	})
}

func (s *Service) VerifyEmail(ctx context.Context, email, code string) error {
	email = normalizeEmail(email)
	if email == "" || code == "" {
		return apperr.Validation
	}
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u == nil {
		return apperr.Validation
	}
	if u.EmailVerified() {
		return nil
	}
	if u.VerificationCodeHash == "" || u.VerificationExpiresAt == nil || !s.now().Before(*u.VerificationExpiresAt) {
		return apperr.Validation
	}
	if !verifyVerificationCode(u.VerificationCodeHash, code) {
		return apperr.Validation
	}
	return s.repo.MarkEmailVerified(ctx, u.ID, s.now().UTC())
}

func newVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashVerificationCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func verifyVerificationCode(hash, code string) bool {
	got := hashVerificationCode(code)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(got)) == 1
}
