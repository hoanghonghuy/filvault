package call

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenGenerator_GenerateToken(t *testing.T) {
	apiKey := "test-key"
	apiSecret := "test-secret"
	generator := NewTokenGenerator(apiKey, apiSecret, time.Hour)

	room := "conv-123"
	identity := "user-456"
	name := "Alice"

	tokenStr, err := generator.GenerateToken(room, identity, name)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	if tokenStr == "" {
		t.Fatalf("expected non-empty token")
	}

	// Parse and verify claims
	var claims LiveKitClaims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(apiSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatalf("token is not valid")
	}

	if claims.Issuer != apiKey {
		t.Errorf("expected issuer %q, got %q", apiKey, claims.Issuer)
	}
	if claims.Subject != identity {
		t.Errorf("expected subject %q, got %q", identity, claims.Subject)
	}
	if claims.Name != name {
		t.Errorf("expected name %q, got %q", name, claims.Name)
	}
	if claims.Video.Room != room {
		t.Errorf("expected room %q, got %q", room, claims.Video.Room)
	}
	if !claims.Video.RoomJoin {
		t.Errorf("expected RoomJoin to be true")
	}
	if !claims.Video.CanPublish {
		t.Errorf("expected CanPublish to be true")
	}
	if !claims.Video.CanSubscribe {
		t.Errorf("expected CanSubscribe to be true")
	}
	if !claims.Video.CanPublishData {
		t.Errorf("expected CanPublishData to be true")
	}
}
