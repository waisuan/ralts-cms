// Package auth provides authentication and authorization functionality
// including password hashing, JWT token generation, and validation.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"strconv"
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

// VerifyPassword verifies a password against a stored hash and salt
func VerifyPassword(password, hash, salt string) error {
	// Combine password and salt and hash with SHA-256 to match the hashing process
	passwordWithSalt := password + salt
	hashed := sha256.Sum256([]byte(passwordWithSalt))

	// Convert to hex string for bcrypt comparison
	passwordForBcrypt := fmt.Sprintf("%x", hashed)

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordForBcrypt))
}

// GenerateJWTToken creates a new JWT token for the given entity ID with the provided secret
func GenerateJWTToken(entityID int, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"entity_id": strconv.Itoa(entityID),
		"exp":       time.Now().Add(24 * time.Hour).Unix(), // Token expires in 24 hours
		"iat":       time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
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
