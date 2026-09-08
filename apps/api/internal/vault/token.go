package vault

import (
	"time"

	"filvault/internal/apperr"

	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct {
	secret []byte
	ttl    time.Duration
}

func NewTokens(secret string, ttl time.Duration) *Tokens {
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}
	return &Tokens{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (t *Tokens) Issue(userID string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(t.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "filvault-vault",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString(t.secret)
	return raw, expiresAt, err
}

func (t *Tokens) Parse(raw string) (string, error) {
	if raw == "" {
		return "", apperr.Unauthorized
	}
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
	if !ok || claims.Subject == "" || claims.Issuer != "filvault-vault" {
		return "", apperr.Unauthorized
	}
	return claims.Subject, nil
}
