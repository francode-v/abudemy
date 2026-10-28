// Package auth provides password hashing and JWT issuance/verification.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials indicates a failed password comparison.
var ErrInvalidCredentials = errors.New("invalid credentials")

// PasswordHasher hashes and verifies passwords with bcrypt.
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher returns a hasher using the provided bcrypt cost.
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{cost: cost}
}

// Hash returns a bcrypt hash of the plaintext password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify reports whether the plaintext password matches the stored hash.
func (h *PasswordHasher) Verify(hash, password string) bool {
	if hash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// PasswordPolicyError describes why a password was rejected.
type PasswordPolicyError struct{ Reason string }

func (e *PasswordPolicyError) Error() string { return e.Reason }

// ValidatePassword enforces the shared password policy: 8 to 72 characters,
// with at least one letter and one ASCII digit. The upper bound matches
// bcrypt's 72-byte hard limit, so a password that passes validation can always
// be hashed. Letters may be accented; the byte limit keeps multibyte input safe.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return &PasswordPolicyError{"password must be at least 8 characters"}
	}
	if len(password) > 72 {
		return &PasswordPolicyError{"password is too long"}
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return &PasswordPolicyError{"password must contain at least one letter and one number"}
	}
	return nil
}

// RandomToken returns a URL-safe random token. The token is used for refresh
// tokens and password reset links; only its SHA-256 hex digest is stored, never
// the raw value.
func RandomToken(byteLen int) (string, error) {
	if byteLen < 16 {
		byteLen = 32
	}
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ConstantTimeEqual compares two strings without leaking timing information.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// NormalizeEmail lowercases and trims an email address.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
