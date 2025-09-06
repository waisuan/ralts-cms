package users_test

import (
	"strings"
	"testing"
	"time"

	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"

	"github.com/stretchr/testify/suite"
)

// UserTestSuite defines the test suite for user model
type UserTestSuite struct {
	suite.Suite
}

func (suite *UserTestSuite) TestSetTimestamps() {
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
		suite.Run(tt.name, func() {
			user := tt.initialUser
			originalCreatedAt := user.CreatedAt
			originalUpdatedAt := user.UpdatedAt

			// Capture time before setting timestamps
			beforeTime := time.Now().UTC()

			user.SetTimestamps()

			// Capture time after setting timestamps
			afterTime := time.Now().UTC()

			// Check that timestamps are within expected range
			suite.Require().NotNil(user.UpdatedAt)
			suite.Assert().True(user.UpdatedAt.After(beforeTime) || user.UpdatedAt.Equal(beforeTime))
			suite.Assert().True(user.UpdatedAt.Before(afterTime) || user.UpdatedAt.Equal(afterTime))

			// Check specific field updates
			for _, field := range tt.expectedFields {
				switch field {
				case "CreatedAt":
					suite.Assert().False(user.CreatedAt.IsZero(), "CreatedAt should not be zero for new user")
					suite.Assert().True(user.CreatedAt.After(beforeTime) || user.CreatedAt.Equal(beforeTime))
					suite.Assert().True(user.CreatedAt.Before(afterTime) || user.CreatedAt.Equal(afterTime))
				case "UpdatedAt":
					suite.Require().NotNil(user.UpdatedAt)
					suite.Assert().False(user.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
				}
			}

			// For existing users, CreatedAt should remain unchanged
			if !originalCreatedAt.IsZero() {
				suite.Assert().True(user.CreatedAt.Equal(originalCreatedAt), "CreatedAt should remain unchanged for existing user")
			}

			// UpdatedAt should always be updated
			if originalUpdatedAt != nil {
				suite.Assert().False(user.UpdatedAt.Equal(*originalUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func (suite *UserTestSuite) TestSetDefaultRole() {
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
		suite.Run(tt.name, func() {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Role:     tt.initialRole,
			}

			user.SetDefaultRole()

			suite.Assert().Equal(tt.expected, user.Role)
		})
	}
}

func (suite *UserTestSuite) TestSetDefaultStatus() {
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
		suite.Run(tt.name, func() {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Status:   tt.initialStatus,
			}

			user.SetDefaultStatus()

			suite.Require().NotNil(user.Status)
			suite.Assert().Equal(tt.expected, *user.Status)
		})
	}
}

func (suite *UserTestSuite) TestSetDefaultApproval() {
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
		suite.Run(tt.name, func() {
			user := tt.initialUser

			user.SetDefaultApproval()

			suite.Assert().Equal(tt.expectedApproval, user.Approved)
		})
	}
}

func (suite *UserTestSuite) TestSetDefaults() {
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
		suite.Run(tt.name, func() {
			user := tt.initialUser

			user.SetDefaults()

			suite.Assert().Equal(tt.expectedRole, user.Role)
			suite.Require().NotNil(user.Status)
			suite.Assert().Equal(tt.expectedStatus, *user.Status)
			suite.Assert().Equal(tt.expectedApproval, user.Approved)
		})
	}
}

func (suite *UserTestSuite) TestSetPassword() {
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
		suite.Run(tt.name, func() {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
			}

			err := user.SetPassword(tt.password)

			if tt.expectError {
				suite.Require().Error(err)
				suite.Assert().Contains(err.Error(), tt.errorMsg)
				return
			}

			suite.Require().NoError(err)

			// Check that password and salt are set
			suite.Assert().NotEmpty(user.Password, "password should not be empty after hashing")
			suite.Assert().NotEmpty(user.Salt, "salt should not be empty after hashing")

			// Check that password is hashed (should start with bcrypt identifier)
			suite.Assert().True(
				strings.HasPrefix(user.Password, "$2a$") || strings.HasPrefix(user.Password, "$2b$"),
				"password should be hashed with bcrypt, got: %s", user.Password,
			)

			// Check salt format (should be hexadecimal)
			suite.Assert().Len(user.Salt, 32, "salt should be 32 characters long")

			// Verify the password can be verified
			err = auth.VerifyPassword(tt.password, user.Password, user.Salt)
			suite.Assert().NoError(err, "password verification failed")
		})
	}
}

func (suite *UserTestSuite) TestIsStatusActive() {
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
		suite.Run(tt.name, func() {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
				Status:   tt.status,
			}

			result := user.IsStatusActive()

			suite.Assert().Equal(tt.expected, result)
		})
	}
}

func (suite *UserTestSuite) TestSetPasswordVerification() {
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
		suite.Run(tt.name, func() {
			user := &users.User{
				Username: "Test User",
				Email:    "test@example.com",
			}

			// Set the password
			err := user.SetPassword(tt.password)
			suite.Require().NoError(err, "failed to set password")

			// Verify correct password works
			err = auth.VerifyPassword(tt.password, user.Password, user.Salt)
			suite.Assert().NoError(err, "correct password verification failed")

			// Verify wrong password fails
			err = auth.VerifyPassword(tt.wrongPass, user.Password, user.Salt)
			if tt.expectError {
				suite.Assert().Error(err, "wrong password should fail verification")
			} else {
				suite.Assert().NoError(err, "wrong password verification should succeed")
			}
		})
	}
}

func (suite *UserTestSuite) TestValidateStatusValue() {
	tests := []struct {
		name        string
		status      string
		expectError bool
	}{
		{
			name:        "should accept pending_approval status",
			status:      users.StatusPendingApproval,
			expectError: false,
		},
		{
			name:        "should accept approved status",
			status:      users.StatusApproved,
			expectError: false,
		},
		{
			name:        "should accept suspended status",
			status:      users.StatusSuspended,
			expectError: false,
		},
		{
			name:        "should accept inactive status",
			status:      users.StatusInactive,
			expectError: false,
		},
		{
			name:        "should reject invalid status",
			status:      "invalid_status",
			expectError: true,
		},
		{
			name:        "should reject empty status",
			status:      "",
			expectError: true,
		},
		{
			name:        "should reject uppercase status",
			status:      "APPROVED",
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := users.ValidateStatusValue(tt.status)

			if tt.expectError {
				suite.Require().Error(err)
				suite.Assert().Contains(err.Error(), "invalid status")
			} else {
				suite.Assert().NoError(err)
			}
		})
	}
}

// TestUserTestSuite runs the user test suite
func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
