package auth

import (
	"strings"
	"testing"
)

func TestPasswordHasher(t *testing.T) {
	h := NewPasswordHasher(10)
	hash, err := h.Hash("S3cretPass")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "S3cretPass" {
		t.Fatal("password was not hashed")
	}
	if !h.Verify(hash, "S3cretPass") {
		t.Fatal("expected password to verify")
	}
	if h.Verify(hash, "wrong") {
		t.Fatal("wrong password must not verify")
	}
	if h.Verify("", "S3cretPass") {
		t.Fatal("empty hash must not verify")
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "abcd1234", false},
		{"too short", "a1b2c3", true},
		{"no digit", "abcdefgh", true},
		{"no letter", "12345678", true},
		{"too long", string(make([]byte, 200)), true},
		{"accented letters are valid", "contraseña1", false},
		{"at byte limit", strings.Repeat("a", 71) + "1", false},
		{"over byte limit", strings.Repeat("a", 72) + "1", true},
		{"multibyte over byte limit", strings.Repeat("ñ", 36) + "1", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.password)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRandomTokenUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		tok, err := RandomToken(32)
		if err != nil {
			t.Fatalf("RandomToken: %v", err)
		}
		if len(tok) < 32 {
			t.Fatalf("token too short: %d", len(tok))
		}
		if seen[tok] {
			t.Fatal("duplicate token generated")
		}
		seen[tok] = true
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	if HashToken("abc") != HashToken("abc") {
		t.Fatal("HashToken must be deterministic")
	}
	if HashToken("abc") == HashToken("abd") {
		t.Fatal("different tokens must hash differently")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("got %q", got)
	}
}
