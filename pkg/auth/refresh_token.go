package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenerateRawRefreshToken returns a URL-safe token for the client. Only a hash
// of this value should be stored in the database.
func GenerateRawRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefreshToken returns a non-reversible hash for storage and lookup. The
// pepper must be non-empty in production; callers typically pass
// fmt.Sprintf("%s:ralts-refresh", jwtSecret) or a dedicated env value.
func HashRefreshToken(raw, pepper string) []byte {
	h := sha256.New()
	_, _ = h.Write([]byte(raw))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(pepper))
	return h.Sum(nil)
}
