package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"filnest/internal/apperr"
	"filnest/internal/platform/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type Tokens struct {
	secret []byte
}

func NewTokens(secret string) *Tokens {
	return &Tokens{secret: []byte(secret)}
}

func NewID() string {
	return ulid.Make().String()
}

func (t *Tokens) Access(userID string, now time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "filnest",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessTokenTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

func (t *Tokens) ParseAccess(raw string) (string, error) {
	parsed, err := jwt.ParseWithClaims(
		raw,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return t.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !parsed.Valid {
		return "", apperr.Unauthorized
	}
	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return "", apperr.Unauthorized
	}
	return claims.Subject, nil
}

func newRefreshToken() (plaintext string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	plaintext = hex.EncodeToString(buf)
	return plaintext, hashRefresh(plaintext), nil
}

func hashRefresh(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

func bearer(header string) (string, error) {
	if header == "" {
		return "", apperr.Unauthorized
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", apperr.Unauthorized
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", apperr.Unauthorized
	}
	return token, nil
}

var dummyHashOnce sync.Once
var dummyHash string

func dummyPasswordHash() string {
	dummyHashOnce.Do(func() {
		hash, err := hashPassword("dummy-password-not-used")
		if err != nil {
			panic(fmt.Sprintf("dummy hash: %v", err))
		}
		dummyHash = hash
	})
	return dummyHash
}
