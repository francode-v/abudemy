package auth

import (
	"testing"
	"time"

	"github.com/abudemy/backend/internal/models"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	m := NewTokenManager("a-very-secret-signing-key-of-length", "abudemy-test", 15*time.Minute)
	user := &models.User{ID: "11111111-1111-1111-1111-111111111111", Role: models.RoleMobileUser}

	token, expires, err := m.GenerateAccess(user, "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if time.Until(expires) <= 0 {
		t.Fatal("token should not already be expired")
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("user id mismatch: %s", claims.UserID)
	}
	if claims.Role != string(models.RoleMobileUser) {
		t.Fatalf("role mismatch: %s", claims.Role)
	}
	if claims.SessionID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("session id mismatch: %s", claims.SessionID)
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	m := NewTokenManager("a-very-secret-signing-key-of-length", "abudemy-test", 15*time.Minute)
	user := &models.User{ID: "11111111-1111-1111-1111-111111111111", Role: models.RoleAdmin}
	token, _, err := m.GenerateAccess(user, "sess")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.Parse(token + "x"); err == nil {
		t.Fatal("tampered token must fail to parse")
	}
}

func TestParseRejectsWrongIssuer(t *testing.T) {
	signer := NewTokenManager("a-very-secret-signing-key-of-length", "issuer-a", time.Minute)
	verifier := NewTokenManager("a-very-secret-signing-key-of-length", "issuer-b", time.Minute)
	user := &models.User{ID: "id", Role: models.RoleAdmin}
	token, _, err := signer.GenerateAccess(user, "sess")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := verifier.Parse(token); err == nil {
		t.Fatal("token from another issuer must be rejected")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	m := NewTokenManager("a-very-secret-signing-key-of-length", "issuer", -time.Minute)
	user := &models.User{ID: "id", Role: models.RoleAdmin}
	token, _, err := m.GenerateAccess(user, "sess")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := m.Parse(token); err == nil {
		t.Fatal("expired token must be rejected")
	}
}
