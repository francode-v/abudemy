package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken returns a deterministic SHA-256 hex digest of a high-entropy
// token. Tokens (refresh tokens, reset tokens) are random 256-bit values, so a
// fast digest is appropriate and lets the store look them up by indexed hash.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
