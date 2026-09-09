package vault

import (
	"context"
	"strings"
	"time"

	"filvault/internal/apperr"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo     Repository
	verifier UserVerifier
	tokens   *Tokens
	now      func() time.Time
}

func NewService(repo Repository, verifier UserVerifier, tokens *Tokens) *Service {
	return &Service{
		repo:     repo,
		verifier: verifier,
		tokens:   tokens,
		now:      time.Now,
	}
}

func (s *Service) Status(ctx context.Context, userID string, rawVaultToken string) (Status, error) {
	v, err := s.repo.GetVault(ctx, userID)
	if err != nil {
		return Status{}, err
	}
	if v == nil {
		return Status{Initialized: false, Unlocked: false}, nil
	}

	unlocked := false
	if strings.TrimSpace(rawVaultToken) != "" {
		sub, err := s.tokens.Parse(strings.TrimSpace(rawVaultToken))
		if err == nil && sub == userID {
			unlocked = true
		}
	}

	return Status{
		Initialized: true,
		Unlocked:    unlocked,
	}, nil
}

func (s *Service) Setup(ctx context.Context, userID string, pin string) (Session, error) {
	pin = strings.TrimSpace(pin)
	if len(pin) < 4 || len(pin) > 64 {
		return Session{}, apperr.Validation
	}

	existing, err := s.repo.GetVault(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if existing != nil {
		return Session{}, apperr.Conflict
	}

	hash, err := hashPin(pin)
	if err != nil {
		return Session{}, err
	}

	now := s.now().UTC()
	if err := s.repo.CreateVault(ctx, userID, hash, now); err != nil {
		return Session{}, err
	}

	token, expiresAt, err := s.tokens.Issue(userID, now)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Unlocked:  true,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Unlock(ctx context.Context, userID string, pin string) (Session, error) {
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return Session{}, apperr.Validation
	}

	v, err := s.repo.GetVault(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if v == nil {
		return Session{}, apperr.NotFound
	}

	if !verifyPin(v.PinHash, pin) {
		return Session{}, apperr.New(401, "INVALID_VAULT_PIN", "Mật khẩu kho không chính xác")
	}

	now := s.now().UTC()
	token, expiresAt, err := s.tokens.Issue(userID, now)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Unlocked:  true,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) ChangePin(ctx context.Context, userID string, currentPin string, newPin string) error {
	currentPin = strings.TrimSpace(currentPin)
	newPin = strings.TrimSpace(newPin)
	if len(newPin) < 4 || len(newPin) > 64 {
		return apperr.Validation
	}

	v, err := s.repo.GetVault(ctx, userID)
	if err != nil {
		return err
	}
	if v == nil {
		return apperr.NotFound
	}

	if !verifyPin(v.PinHash, currentPin) {
		return apperr.New(401, "INVALID_VAULT_PIN", "Mật khẩu kho hiện tại không chính xác")
	}

	newHash, err := hashPin(newPin)
	if err != nil {
		return err
	}

	return s.repo.UpdateVaultPin(ctx, userID, newHash, s.now().UTC())
}

func (s *Service) ResetPin(ctx context.Context, userID string, accountPassword string, newPin string) (Session, error) {
	newPin = strings.TrimSpace(newPin)
	if len(newPin) < 4 || len(newPin) > 64 || accountPassword == "" {
		return Session{}, apperr.Validation
	}

	if s.verifier != nil {
		ok, err := s.verifier.VerifyPassword(ctx, userID, accountPassword)
		if err != nil {
			return Session{}, err
		}
		if !ok {
			return Session{}, apperr.New(401, "INVALID_ACCOUNT_PASSWORD", "Mật khẩu tài khoản không chính xác")
		}
	}

	newHash, err := hashPin(newPin)
	if err != nil {
		return Session{}, err
	}

	now := s.now().UTC()
	v, err := s.repo.GetVault(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if v == nil {
		if err := s.repo.CreateVault(ctx, userID, newHash, now); err != nil {
			return Session{}, err
		}
	} else {
		if err := s.repo.UpdateVaultPin(ctx, userID, newHash, now); err != nil {
			return Session{}, err
		}
	}

	token, expiresAt, err := s.tokens.Issue(userID, now)
	if err != nil {
		return Session{}, err
	}

	return Session{
		Unlocked:  true,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) ListFiles(ctx context.Context, userID string) ([]VaultFile, error) {
	return s.repo.ListVaultFiles(ctx, userID)
}

func (s *Service) MoveToVault(ctx context.Context, userID string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return apperr.Validation
	}
	return s.repo.MoveFilesToVault(ctx, userID, fileIDs)
}

func (s *Service) MoveFromVault(ctx context.Context, userID string, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return apperr.Validation
	}
	return s.repo.MoveFilesFromVault(ctx, userID, fileIDs)
}

func (s *Service) ValidateToken(token string) (string, error) {
	return s.tokens.Parse(token)
}

func hashPin(pin string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	return string(b), err
}

func verifyPin(hash, pin string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil
}
