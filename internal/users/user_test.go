package users_test

import (
	"strings"
	"testing"
	"time"

	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"
)

func TestUser_SetTimestamps(t *testing.T) {
	tests := []struct {
		name           string
		initialUser    *users.User
		expectedFields []string
	}{
		{
			name: "should set both timestamps for new user",
			initialUser: &users.User{
				Username: "Test User",
				Email:    "test@example.com",
			},
			expectedFields: []string{"CreatedAt", "UpdatedAt"},
		},
		{
			name: "should only update UpdatedAt for existing user",
			initialUser: &users.User{
				Username:  "Test User",
				Email:     "test@example.com",
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedFields: []string{"UpdatedAt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.initialUser
			originalCreatedAt := user.CreatedAt
			originalUpdatedAt := user.UpdatedAt

			// Capture time before setting timestamps
			beforeTime := time.Now().UTC()

			user.SetTimestamps()

			// Capture time after setting timestamps
			afterTime := time.Now().UTC()

			// Check that timestamps are within expected range
			if user.UpdatedAt == nil || user.UpdatedAt.Before(beforeTime) || user.UpdatedAt.After(afterTime) {
				t.Errorf("UpdatedAt should be between %v and %v, got %v", beforeTime, afterTime, user.UpdatedAt)
			}

			// Check specific field updates
			for _, field := range tt.expectedFields {
				switch field {
				case "CreatedAt":
					if user.CreatedAt.IsZero() {
						t.Errorf("CreatedAt should not be zero for new user")
					}
					if user.CreatedAt.Before(beforeTime) || user.CreatedAt.After(afterTime) {
						t.Errorf("CreatedAt should be between %v and %v, got %v", beforeTime, afterTime, user.CreatedAt)
					}
				case "UpdatedAt":
					if user.UpdatedAt == nil || user.UpdatedAt.IsZero() {
						t.Errorf("UpdatedAt should not be zero")
					}
				}
			}

			// For existing users, CreatedAt should remain unchanged
			if !originalCreatedAt.IsZero() && !user.CreatedAt.Equal(originalCreatedAt) {
				t.Errorf("CreatedAt should remain unchanged for existing user, expected %v, got %v", originalCreatedAt, user.CreatedAt)
			}

			// UpdatedAt should always be updated
			if originalUpdatedAt != nil && user.UpdatedAt != nil && user.UpdatedAt.Equal(*originalUpdatedAt) {
				t.Errorf("UpdatedAt should be updated, expected different from %v, got %v", originalUpdatedAt, user.UpdatedAt)
			}
		})
	}
}

func TestUser_SetDefaultRole(t *testing.T) {
	tests := []struct {
		name        string
		initialRole string
		expected    string
	}{
		{
			name:        "should set default role when empty",
			initialRole: "",
			expected:    "NON_ADMIN",
		},
		{
			name:        "should not change existing role",
			initialRole: "admin",
			expected:    "admin",
		},
		{
			name:        "should not change custom role",
			initialRole: "moderator",
			expected:    "moderator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Role:     tt.initialRole,
			}

			user.SetDefaultRole()

			if user.Role != tt.expected {
				t.Errorf("expected role %s, got %s", tt.expected, user.Role)
			}
		})
	}
}

func TestUser_SetDefaultStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus *string
		expected      string
	}{
		{
			name:          "should set default status when nil",
			initialStatus: nil,
			expected:      users.StatusPendingApproval,
		},
		{
			name:          "should not change existing status",
			initialStatus: testutils.StringPtr("inactive"),
			expected:      "inactive",
		},
		{
			name:          "should not change custom status",
			initialStatus: testutils.StringPtr("suspended"),
			expected:      "suspended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Status:   tt.initialStatus,
			}

			user.SetDefaultStatus()

			if user.Status == nil || *user.Status != tt.expected {
				t.Errorf("expected status %s, got %v", tt.expected, user.Status)
			}
		})
	}
}

func TestUser_SetDefaultApproval(t *testing.T) {
	tests := []struct {
		name             string
		initialUser      *users.User
		expectedApproval bool
	}{
		{
			name: "should set default approval for new user",
			initialUser: &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				// CreatedAt is zero (new user)
			},
			expectedApproval: false,
		},
		{
			name: "should not change approval for existing user",
			initialUser: &users.User{
				Username:  "Test User",
				Email:     "test@example.com",
				Approved:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), // existing user
			},
			expectedApproval: true, // should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.initialUser

			user.SetDefaultApproval()

			if user.Approved != tt.expectedApproval {
				t.Errorf("expected approval %v, got %v", tt.expectedApproval, user.Approved)
			}
		})
	}
}

func TestUser_SetDefaults(t *testing.T) {
	tests := []struct {
		name             string
		initialUser      *users.User
		expectedRole     string
		expectedStatus   string
		expectedApproval bool
	}{
		{
			name: "should set all defaults when empty for new user",
			initialUser: &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Role:     "",
				Status:   nil,
				// CreatedAt is zero (new user)
			},
			expectedRole:     "NON_ADMIN",
			expectedStatus:   users.StatusPendingApproval,
			expectedApproval: false,
		},
		{
			name: "should not change existing values",
			initialUser: &users.User{
				Username:  "Test User",
				Email:     "test@example.com",
				Role:      "admin",
				Status:    testutils.StringPtr("inactive"),
				Approved:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), // existing user
			},
			expectedRole:     "admin",
			expectedStatus:   "inactive",
			expectedApproval: true,
		},
		{
			name: "should set only empty values for new user",
			initialUser: &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Role:     "moderator",
				Status:   nil,
				// CreatedAt is zero (new user)
			},
			expectedRole:     "moderator",
			expectedStatus:   users.StatusPendingApproval,
			expectedApproval: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.initialUser

			user.SetDefaults()

			if user.Role != tt.expectedRole {
				t.Errorf("expected role %s, got %s", tt.expectedRole, user.Role)
			}

			if user.Status == nil || *user.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %v", tt.expectedStatus, user.Status)
			}

			if user.Approved != tt.expectedApproval {
				t.Errorf("expected approval %v, got %v", tt.expectedApproval, user.Approved)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "should hash password successfully",
			password:    "mypassword123",
			expectError: false,
		},
		{
			name:        "should hash complex password",
			password:    "P@ssw0rd!@#$%^&*()",
			expectError: false,
		},
		{
			name:        "should hash unicode password",
			password:    "密码123",
			expectError: false,
		},
		{
			name:        "should fail with empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
			}

			err := user.SetPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check that password and salt are set
			if user.Password == "" {
				t.Errorf("password should not be empty after hashing")
			}

			if user.Salt == "" {
				t.Errorf("salt should not be empty after hashing")
			}

			// Check that password is hashed (should start with bcrypt identifier)
			if !strings.HasPrefix(user.Password, "$2a$") && !strings.HasPrefix(user.Password, "$2b$") {
				t.Errorf("password should be hashed with bcrypt, got: %s", user.Password)
			}

			// Check salt format (should be hexadecimal)
			if len(user.Salt) != 32 {
				t.Errorf("salt should be 32 characters long, got %d", len(user.Salt))
			}

			// Verify the password can be verified
			if err := auth.VerifyPassword(tt.password, user.Password, user.Salt); err != nil {
				t.Errorf("password verification failed: %v", err)
			}
		})
	}
}

func TestUser_IsStatusActive(t *testing.T) {
	tests := []struct {
		name     string
		status   *string
		expected bool
	}{
		{
			name:     "should return true for approved status",
			status:   testutils.StringPtr(users.StatusApproved),
			expected: true,
		},
		{
			name:     "should return false for pending approval status",
			status:   testutils.StringPtr(users.StatusPendingApproval),
			expected: false,
		},
		{
			name:     "should return false for inactive status",
			status:   testutils.StringPtr(users.StatusInactive),
			expected: false,
		},
		{
			name:     "should return false for suspended status",
			status:   testutils.StringPtr(users.StatusSuspended),
			expected: false,
		},
		{
			name:     "should return false for nil status",
			status:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Status:   tt.status,
			}

			result := user.IsStatusActive()

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUser_SetPassword_Verification(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wrongPass   string
		expectError bool
	}{
		{
			name:        "should verify correct password",
			password:    "mypassword123",
			wrongPass:   "wrongpassword",
			expectError: true,
		},
		{
			name:        "should verify complex password",
			password:    "P@ssw0rd!@#$%^&*()",
			wrongPass:   "wrongpassword",
			expectError: true,
		},
		{
			name:        "should verify unicode password",
			password:    "密码123",
			wrongPass:   "wrongpassword",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
			}

			// Set the password
			err := user.SetPassword(tt.password)
			if err != nil {
				t.Errorf("failed to set password: %v", err)
				return
			}

			// Verify correct password works
			err = auth.VerifyPassword(tt.password, user.Password, user.Salt)
			if err != nil {
				t.Errorf("correct password verification failed: %v", err)
			}

			// Verify wrong password fails
			err = auth.VerifyPassword(tt.wrongPass, user.Password, user.Salt)
			if !tt.expectError || err == nil {
				t.Errorf("wrong password should fail verification")
			}
		})
	}
}
