package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"filvault/internal/apperr"
	"filvault/internal/platform/config"
	"filvault/internal/platform/mailer"
	"filvault/internal/user"
)

type Service struct {
	cfg    config.Config
	repo   Repository
	tokens *Tokens
	mailer mailer.Mailer
	now    func() time.Time
	dummy  string
}

type Session struct {
	User         user.User
	AccessToken  string
	RefreshToken string
}

func NewService(cfg config.Config, repo Repository, tokens *Tokens, mailer mailer.Mailer) *Service {
	return &Service{
		cfg:    cfg,
		repo:   repo,
		tokens: tokens,
		mailer: mailer,
		now:    time.Now,
		dummy:  dummyPasswordHash(),
	}
}

func (s *Service) Register(ctx context.Context, email, password, displayName, inviteCode string) (Session, error) {
	if err := checkInvite(inviteCode, s.cfg.InviteCode); err != nil {
		return Session{}, err
	}
	email = normalizeEmail(email)
	displayName = strings.TrimSpace(displayName)
	if email == "" || !strings.Contains(email, "@") || displayName == "" || len(password) < config.MinPasswordLength {
		return Session{}, apperr.Validation
	}

	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return Session{}, err
	}
	if existing != nil {
		return Session{}, apperr.Conflict
	}

	hash, err := hashPassword(password)
	if err != nil {
		return Session{}, err
	}
	now := s.now().UTC()
	u := user.User{
		ID:                     NewID(),
		Email:                  email,
		DisplayName:            displayName,
		PasswordHash:           hash,
		StorageUsed:            0,
		StorageQuota:           config.DefaultStorageQuotaBytes,
		TrashAutoDeleteEnabled: s.cfg.DefaultTrashAutoDelete,
		TrashRetentionDays:     s.cfg.DefaultTrashRetentionDays,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if u.TrashRetentionDays < 1 {
		u.TrashRetentionDays = config.DefaultTrashRetentionDays
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return Session{}, err
	}
	if err := s.ResendVerification(ctx, u.Email); err != nil {
		return Session{}, err
	}
	return s.issueSession(ctx, u, now)
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return Session{}, apperr.Validation
	}
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return Session{}, err
	}
	if u == nil {
		_ = verifyPassword(s.dummy, password)
		return Session{}, apperr.Unauthorized
	}
	if !verifyPassword(u.PasswordHash, password) {
		return Session{}, apperr.Unauthorized
	}
	return s.issueSession(ctx, *u, s.now().UTC())
}

func (s *Service) Refresh(ctx context.Context, refreshPlain string) (Session, error) {
	tok, err := s.loadActiveRefresh(ctx, refreshPlain)
	if err != nil {
		return Session{}, err
	}
	u, err := s.repo.GetUserByID(ctx, tok.UserID)
	if err != nil {
		return Session{}, err
	}
	if u == nil {
		return Session{}, apperr.Unauthorized
	}
	access, err := s.tokens.Access(u.ID, s.now().UTC())
	if err != nil {
		return Session{}, err
	}
	return Session{User: *u, AccessToken: access, RefreshToken: refreshPlain}, nil
}

func (s *Service) Logout(ctx context.Context, userID, refreshPlain string) error {
	tok, err := s.loadActiveRefresh(ctx, refreshPlain)
	if err != nil {
		return err
	}
	if tok.UserID != userID {
		return apperr.Unauthorized
	}
	return s.repo.RevokeRefreshToken(ctx, tok.TokenHash, s.now().UTC())
}

func (s *Service) Me(ctx context.Context, userID string) (user.User, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return user.User{}, err
	}
	if u == nil {
		return user.User{}, apperr.Unauthorized
	}
	return *u, nil
}

func (s *Service) PatchMe(ctx context.Context, userID string, displayName *string, autoDelete *bool, retentionDays *int) (user.User, error) {
	u, err := s.Me(ctx, userID)
	if err != nil {
		return user.User{}, err
	}
	if displayName != nil {
		name := strings.TrimSpace(*displayName)
		if name == "" {
			return user.User{}, apperr.Validation
		}
		if err := s.repo.UpdateDisplayName(ctx, userID, name); err != nil {
			return user.User{}, err
		}
	}
	if autoDelete != nil || retentionDays != nil {
		enabled := u.TrashAutoDeleteEnabled
		days := u.TrashRetentionDays
		if autoDelete != nil {
			enabled = *autoDelete
		}
		if retentionDays != nil {
			if *retentionDays < 1 {
				return user.User{}, apperr.Validation
			}
			days = *retentionDays
		}
		if err := s.repo.UpdateTrashSettings(ctx, userID, enabled, days); err != nil {
			return user.User{}, err
		}
	}
	return s.Me(ctx, userID)
}

func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) (Session, error) {
	if currentPassword == "" || len(newPassword) < config.MinPasswordLength {
		return Session{}, apperr.Validation
	}
	u, err := s.Me(ctx, userID)
	if err != nil {
		return Session{}, err
	}
	if !verifyPassword(u.PasswordHash, currentPassword) {
		return Session{}, apperr.Unauthorized
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return Session{}, err
	}
	now := s.now().UTC()
	if err := s.repo.UpdatePasswordHash(ctx, userID, hash); err != nil {
		return Session{}, err
	}
	if err := s.repo.RevokeAllRefreshTokensForUser(ctx, userID, now); err != nil {
		return Session{}, err
	}
	u.PasswordHash = hash
	return s.issueSession(ctx, u, now)
}

func (s *Service) ParseAccess(raw string) (string, error) {
	return s.tokens.ParseAccess(raw)
}

func (s *Service) issueSession(ctx context.Context, u user.User, now time.Time) (Session, error) {
	access, err := s.tokens.Access(u.ID, now)
	if err != nil {
		return Session{}, err
	}
	plain, hash, err := newRefreshToken()
	if err != nil {
		return Session{}, err
	}
	rt := RefreshToken{
		ID:        NewID(),
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: now.Add(config.RefreshTokenTTL),
		CreatedAt: now,
	}
	if err := s.repo.InsertRefreshToken(ctx, rt); err != nil {
		return Session{}, err
	}
	return Session{User: u, AccessToken: access, RefreshToken: plain}, nil
}

func (s *Service) loadActiveRefresh(ctx context.Context, plain string) (*RefreshToken, error) {
	if plain == "" {
		return nil, apperr.Validation
	}
	tok, err := s.repo.GetRefreshTokenByHash(ctx, hashRefresh(plain))
	if err != nil {
		return nil, err
	}
	if tok == nil || tok.RevokedAt != nil || !s.now().Before(tok.ExpiresAt) {
		return nil, apperr.Unauthorized
	}
	return tok, nil
}

func checkInvite(got, want string) error {
	if want == "" {
		return apperr.RegisterDisabled
	}
	sumGot := sha256.Sum256([]byte(got))
	sumWant := sha256.Sum256([]byte(want))
	if subtle.ConstantTimeCompare(sumGot[:], sumWant[:]) != 1 {
		return apperr.Forbidden
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
