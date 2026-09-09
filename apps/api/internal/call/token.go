package call

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type VideoGrant struct {
	Room           string `json:"room"`
	RoomJoin       bool   `json:"roomJoin"`
	CanPublish     bool   `json:"canPublish"`
	CanSubscribe   bool   `json:"canSubscribe"`
	CanPublishData bool   `json:"canPublishData"`
}

type LiveKitClaims struct {
	jwt.RegisteredClaims
	Name  string     `json:"name,omitempty"`
	Video VideoGrant `json:"video"`
}

type TokenGenerator struct {
	apiKey    string
	apiSecret string
	ttl       time.Duration
}

func NewTokenGenerator(apiKey, apiSecret string, ttl time.Duration) *TokenGenerator {
	if ttl <= 0 {
		ttl = 6 * time.Hour
	}
	return &TokenGenerator{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		ttl:       ttl,
	}
}

func (g *TokenGenerator) GenerateToken(room, identity, name string) (string, error) {
	now := time.Now().UTC()
	claims := LiveKitClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.apiKey,
			Subject:   identity,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(g.ttl)),
		},
		Name: name,
		Video: VideoGrant{
			Room:           room,
			RoomJoin:       true,
			CanPublish:     true,
			CanSubscribe:   true,
			CanPublishData: true,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.apiSecret))
}
