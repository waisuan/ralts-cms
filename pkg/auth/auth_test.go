package auth_test

import (
	"strconv"
	"strings"
	"testing"

	"ralts-cms/pkg/auth"
)

func TestGenerateSalt(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "should generate salt successfully",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, err := auth.GenerateSalt()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check salt length (16 bytes = 32 hex characters)
			if len(salt) != 32 {
				t.Errorf("expected salt length 32, got %d", len(salt))
			}

			// Check that salt is hexadecimal
			if !isHexString(salt) {
				t.Errorf("salt is not hexadecimal: %s", salt)
			}

			// Generate another salt and verify they're different
			salt2, err := auth.GenerateSalt()
			if err != nil {
				t.Errorf("failed to generate second salt: %v", err)
				return
			}

			if salt == salt2 {
				t.Errorf("generated salts should be different: %s", salt)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		salt        string
		expectError bool
	}{
		{
			name:        "should hash password successfully",
			password:    "mypassword123",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should hash empty password",
			password:    "",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should hash with empty salt",
			password:    "mypassword123",
			salt:        "",
			expectError: false,
		},
		{
			name:        "should hash with special characters",
			password:    "p@ssw0rd!@#$%",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should hash with unicode characters",
			password:    "password中文",
			salt:        "abcdef1234567890",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := auth.HashPassword(tt.password, tt.salt)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check that hash is not empty
			if hash == "" {
				t.Errorf("hash should not be empty")
			}

			// Check that hash starts with bcrypt identifier
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Errorf("hash should start with bcrypt identifier, got: %s", hash)
			}

			// Check that hash is different for same password with different salt
			if tt.salt != "" {
				hash2, err := auth.HashPassword(tt.password, tt.salt+"different")
				if err != nil {
					t.Errorf("failed to generate second hash: %v", err)
					return
				}

				if hash == hash2 {
					t.Errorf("hashes should be different for different salts")
				}
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		salt        string
		expectError bool
	}{
		{
			name:        "should verify correct password",
			password:    "mypassword123",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should verify empty password",
			password:    "",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should verify with empty salt",
			password:    "mypassword123",
			salt:        "",
			expectError: false,
		},
		{
			name:        "should verify with special characters",
			password:    "p@ssw0rd!@#$%",
			salt:        "abcdef1234567890",
			expectError: false,
		},
		{
			name:        "should verify with unicode characters",
			password:    "password中文",
			salt:        "abcdef1234567890",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, hash the password
			hash, err := auth.HashPassword(tt.password, tt.salt)
			if err != nil {
				t.Errorf("failed to hash password: %v", err)
				return
			}

			// Then verify it
			err = auth.VerifyPassword(tt.password, hash, tt.salt)
			if err != nil {
				t.Errorf("failed to verify password: %v", err)
			}
		})
	}
}

func TestVerifyPasswordWithWrongPassword(t *testing.T) {
	tests := []struct {
		name        string
		correctPass string
		wrongPass   string
		salt        string
		expectError bool
	}{
		{
			name:        "should fail with wrong password",
			correctPass: "mypassword123",
			wrongPass:   "wrongpassword",
			salt:        "abcdef1234567890",
			expectError: true,
		},
		{
			name:        "should fail with empty password when original was not empty",
			correctPass: "mypassword123",
			wrongPass:   "",
			salt:        "abcdef1234567890",
			expectError: true,
		},
		{
			name:        "should fail with non-empty password when original was empty",
			correctPass: "",
			wrongPass:   "somepassword",
			salt:        "abcdef1234567890",
			expectError: true,
		},
		{
			name:        "should fail with different case",
			correctPass: "MyPassword123",
			wrongPass:   "mypassword123",
			salt:        "abcdef1234567890",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, hash the correct password
			hash, err := auth.HashPassword(tt.correctPass, tt.salt)
			if err != nil {
				t.Errorf("failed to hash password: %v", err)
				return
			}

			// Then try to verify with wrong password
			err = auth.VerifyPassword(tt.wrongPass, hash, tt.salt)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyPasswordWithWrongSalt(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		correctSalt string
		wrongSalt   string
		expectError bool
	}{
		{
			name:        "should fail with wrong salt",
			password:    "mypassword123",
			correctSalt: "abcdef1234567890",
			wrongSalt:   "different1234567890",
			expectError: true,
		},
		{
			name:        "should fail with empty salt when original was not empty",
			password:    "mypassword123",
			correctSalt: "abcdef1234567890",
			wrongSalt:   "",
			expectError: true,
		},
		{
			name:        "should fail with non-empty salt when original was empty",
			password:    "mypassword123",
			correctSalt: "",
			wrongSalt:   "someothersalt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, hash the password with correct salt
			hash, err := auth.HashPassword(tt.password, tt.correctSalt)
			if err != nil {
				t.Errorf("failed to hash password: %v", err)
				return
			}

			// Then try to verify with wrong salt
			err = auth.VerifyPassword(tt.password, hash, tt.wrongSalt)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPasswordHashingRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
	}{
		{
			name:     "simple password",
			password: "password123",
			salt:     "salt1234567890",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!@#$%^&*()",
			salt:     "complexsalt123456",
		},
		{
			name:     "unicode password",
			password: "密码123",
			salt:     "unicodesalt123456",
		},
		{
			name:     "empty password",
			password: "",
			salt:     "emptysalt123456",
		},
		{
			name:     "empty salt",
			password: "password123",
			salt:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hash, err := auth.HashPassword(tt.password, tt.salt)
			if err != nil {
				t.Errorf("failed to hash password: %v", err)
				return
			}

			// Verify the password
			err = auth.VerifyPassword(tt.password, hash, tt.salt)
			if err != nil {
				t.Errorf("failed to verify password: %v", err)
			}

			// Verify that the same password with same salt can be verified
			// Note: bcrypt includes its own salt, so hashes will be different each time
			// but both should verify correctly
			hash2, err := auth.HashPassword(tt.password, tt.salt)
			if err != nil {
				t.Errorf("failed to hash password second time: %v", err)
				return
			}

			// Both hashes should verify correctly
			err = auth.VerifyPassword(tt.password, hash2, tt.salt)
			if err != nil {
				t.Errorf("second hash should verify correctly: %v", err)
			}

			// The hashes should be different (bcrypt includes random salt)
			if hash == hash2 {
				t.Errorf("bcrypt hashes should be different due to internal salt")
			}
		})
	}
}

func TestGenerateJWTToken(t *testing.T) {
	tests := []struct {
		name        string
		entityID    int
		role        string
		secret      string
		expectError bool
	}{
		{
			name:        "should generate token successfully with admin role",
			entityID:    1234567890,
			role:        "ADMIN",
			secret:      "your-jwt-secret-key",
			expectError: false,
		},
		{
			name:        "should generate token successfully with non-admin role",
			entityID:    9876543210,
			role:        "NON_ADMIN",
			secret:      "your-jwt-secret-key",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateJWTToken(tt.entityID, tt.role, tt.secret)
			if err != nil {
				t.Errorf("failed to generate token: %v", err)
				return
			}

			if token == "" {
				t.Errorf("token should not be empty")
			}
		})
	}
}

func TestValidateJWTToken(t *testing.T) {
	tests := []struct {
		name        string
		entityID    int
		role        string
		secret      string
		wrongSecret string
		expectError bool
	}{
		{
			name:        "should validate token successfully",
			entityID:    1234567890,
			role:        "ADMIN",
			secret:      "your-jwt-secret-key",
			wrongSecret: "wrong-secret",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a valid token
			token, err := auth.GenerateJWTToken(tt.entityID, tt.role, tt.secret)
			if err != nil {
				t.Errorf("failed to generate token: %v", err)
				return
			}

			// Validate the token with correct secret
			claims, err := auth.ValidateJWTToken(token, tt.secret)
			if err != nil {
				t.Errorf("failed to validate token: %v", err)
				return
			}

			// Check that claims contain expected data
			if claims == nil {
				t.Errorf("claims should not be nil")
				return
			}

			entityID, ok := claims["entity_id"].(string)
			if !ok {
				t.Errorf("entity_id should be present in claims")
				return
			}

			if entityID != strconv.Itoa(tt.entityID) {
				t.Errorf("expected entity_id %d, got %s", tt.entityID, entityID)
			}

			// Check that role is present in claims
			role, ok := claims["role"].(string)
			if !ok {
				t.Errorf("role should be present in claims")
				return
			}

			if role != tt.role {
				t.Errorf("expected role %s, got %s", tt.role, role)
			}

			// Test with wrong secret
			_, err = auth.ValidateJWTToken(token, tt.wrongSecret)
			if err == nil {
				t.Errorf("expected error when validating with wrong secret")
			}
		})
	}
}

func TestValidateJWTTokenWithInvalidToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		secret      string
		expectError bool
	}{
		{
			name:        "should reject invalid token format",
			token:       "invalid.token.format",
			secret:      "your-jwt-secret-key",
			expectError: true,
		},
		{
			name:        "should reject empty token",
			token:       "",
			secret:      "your-jwt-secret-key",
			expectError: true,
		},
		{
			name:        "should reject malformed token",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			secret:      "your-jwt-secret-key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := auth.ValidateJWTToken(tt.token, tt.secret)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function to check if a string is hexadecimal
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
