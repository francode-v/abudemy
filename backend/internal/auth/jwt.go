package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/abudemy/backend/internal/models"
)

// ErrInvalidToken indicates a token that failed verification.
var ErrInvalidToken = errors.New("invalid token")

// Claims is the JWT payload issued to authenticated clients.
type Claims struct {
	UserID    string `json:"uid"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies access tokens.
type TokenManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

// NewTokenManager constructs a token manager.
func NewTokenManager(secret, issuer string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

// AccessTTL returns the configured access token lifetime.
func (m *TokenManager) AccessTTL() time.Duration { return m.accessTTL }

// GenerateAccess signs an access token for the given user and session.
func (m *TokenManager) GenerateAccess(user *models.User, sessionID string) (string, time.Time, error) {
	now := time.Now()
	expires := now.Add(m.accessTTL)
	claims := Claims{
		UserID:    user.ID,
		Role:      string(user.Role),
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

// Parse verifies a token and returns its claims.
func (m *TokenManager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
