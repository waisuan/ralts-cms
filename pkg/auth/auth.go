// Package auth provides authentication and authorization functionality
// including password hashing, JWT token generation, and validation.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GenerateSalt creates a cryptographically secure random salt for password hashing
func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	return fmt.Sprintf("%x", salt), nil
}

// HashPassword hashes a password with the given salt using bcrypt
func HashPassword(password, salt string) (string, error) {
	// Combine password and salt and hash with SHA-256 to avoid bcrypt's 72-byte limit
	passwordWithSalt := password + salt
	hashed := sha256.Sum256([]byte(passwordWithSalt))

	// Convert to hex string for bcrypt
	passwordForBcrypt := fmt.Sprintf("%x", hashed)

	// Hash using bcrypt with cost factor 12 (good balance of security and performance)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordForBcrypt), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a stored hash and salt.
// It supports two hashing schemes:
//   - Ralts (current): SHA-256(password + hex_salt) -> bcrypt
//   - Legacy roti: plain bcrypt(password) with a bcrypt-generated salt
//
// The legacy path is only attempted when the stored salt looks like a bcrypt
// salt string (starts with "$2a$" or "$2b$"), which distinguishes migrated
// roti users from native Ralts users whose salts are random hex strings.
func VerifyPassword(password, hash, salt string) error {
	// Try current scheme first: SHA-256(password + salt) -> bcrypt compare
	passwordWithSalt := password + salt
	hashed := sha256.Sum256([]byte(passwordWithSalt))
	passwordForBcrypt := fmt.Sprintf("%x", hashed)

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordForBcrypt)); err == nil {
		return nil
	}

	// Fallback: legacy roti plain bcrypt (salt is a bcrypt salt string)
	if isLegacyBcryptSalt(salt) {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	}

	return bcrypt.ErrMismatchedHashAndPassword
}

// NeedsPasswordUpgrade reports whether the user's password was hashed with the
// legacy roti scheme and should be re-hashed on next successful login.
func NeedsPasswordUpgrade(salt string) bool {
	return isLegacyBcryptSalt(salt)
}

// isLegacyBcryptSalt detects bcrypt-generated salt strings (e.g. "$2a$10$...")
// used by the legacy roti application. Ralts salts are 32-char hex strings and
// will never match this pattern.
func isLegacyBcryptSalt(salt string) bool {
	return strings.HasPrefix(salt, "$2a$") || strings.HasPrefix(salt, "$2b$")
}

// GenerateAccessToken creates a new JWT (access token) for the given entity,
// with an explicit time-to-live.
func GenerateAccessToken(entityID int, role, secret string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be positive")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"entity_id": strconv.Itoa(entityID),
		"role":      role,
		"exp":       now.Add(ttl).Unix(),
		"iat":       now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// GenerateJWTToken creates a new JWT token for the given entity ID and role
// with a 24-hour lifetime. Tests and call sites that do not use refresh
// flows may use this; application login should prefer GenerateAccessToken with
// a short TTL from configuration.
func GenerateJWTToken(entityID int, role, secret string) (string, error) {
	return GenerateAccessToken(entityID, role, secret, 24*time.Hour)
}

// ValidateJWTToken validates a JWT token and returns the claims
func ValidateJWTToken(tokenString, secret string) (jwt.MapClaims, error) {
	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	return claims, nil
}
