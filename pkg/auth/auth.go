package auth

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GenerateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	return fmt.Sprintf("%x", salt), nil
}

// hashPassword hashes a password with the given salt using bcrypt
func HashPassword(password, salt string) (string, error) {
	// Combine password and salt
	passwordWithSalt := password + salt

	// Hash using bcrypt with cost factor 12 (good balance of security and performance)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordWithSalt), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a stored hash and salt
func VerifyPassword(password, hash, salt string) error {
	passwordWithSalt := password + salt
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordWithSalt))
}
