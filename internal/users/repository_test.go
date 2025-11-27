package users_test

import (
	"context"
	"testing"

	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryTestSuite defines the test suite for user repository
type UserRepositoryTestSuite struct {
	suite.Suite

	db   *testutils.TestDatabase
	repo users.Repository
}

// SetupTest sets up each test
func (suite *UserRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = users.NewRepository(db.PostgresClient)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *UserRepositoryTestSuite) TearDownSubTest() {
	// Clear the users table for PostgreSQL
	ctx := context.Background()
	err := suite.db.CleanupAllTables(ctx)
	require.NoError(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create user successfully", func() {
		user := testutils.CreateUser("testuser", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(user.CreatedAt)
		suite.Assert().NotNil(user.UpdatedAt)
		suite.Assert().Equal(user.CreatedAt, *user.UpdatedAt)
		suite.Assert().Greater(user.ID, int64(0))    // PostgreSQL should return an ID
		suite.Assert().Equal("NON_ADMIN", user.Role) // Default role
		suite.Assert().False(user.Approved)          // Default approval status
		suite.Assert().NotNil(user.Status)
		suite.Assert().Equal(users.StatusPendingApproval, *user.Status) // Default status
		suite.Assert().NotEmpty(user.Password)                          // Should be hashed
		suite.Assert().NotEmpty(user.Salt)                              // Should be set
	})

	suite.Run("should fail when creating duplicate user", func() {
		user := testutils.CreateUser("duplicate", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to create the same user again
		duplicateUser := testutils.CreateUser("duplicate", "differentpassword")
		err = suite.repo.Create(ctx, duplicateUser)
		suite.Require().Error(err)
		// PostgreSQL will return a unique constraint violation error
		suite.Assert().Contains(err.Error(), "duplicate key")
	})

	suite.Run("should create user with custom role and status", func() {
		user := testutils.CreateUserWithCustomFields("adminuser", "adminpass", "admin", "inactive")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().Equal("admin", user.Role)
		suite.Assert().NotNil(user.Status)
		suite.Assert().Equal("inactive", *user.Status)
		suite.Assert().NotEmpty(user.CreatedAt)
		suite.Assert().NotNil(user.UpdatedAt)
	})

	suite.Run("should create user with avatar", func() {
		avatar := "https://example.com/avatar.jpg"
		user := testutils.CreateUserWithAvatar("avataruser", "mypassword123", &avatar)

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().NotNil(user.Avatar)
		suite.Assert().Equal(avatar, *user.Avatar)
	})

	suite.Run("should create user with nil avatar", func() {
		user := testutils.CreateUserWithAvatar("noavataruser", "mypassword123", nil)

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().Nil(user.Avatar)
	})

	suite.Run("should hash password correctly", func() {
		user := testutils.CreateUser("passworduser", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Password should be hashed (not plain text)
		suite.Assert().NotEqual("mypassword123", user.Password)
		suite.Assert().NotEmpty(user.Password)
		suite.Assert().NotEmpty(user.Salt)
		// bcrypt hashes are long
		suite.Assert().GreaterOrEqual(len(user.Password), 60)
	})
}

func (suite *UserRepositoryTestSuite) TestCreate_Validation() {
	ctx := context.Background()

	suite.Run("should fail with missing username", func() {
		user := &users.User{
			Email:    "test@example.com",
			Password: "mypassword123",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "username is required")
	})

	suite.Run("should fail with missing email", func() {
		user := &users.User{
			Username: "Test User",
			Password: "mypassword123",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "email is required")
	})

	suite.Run("should fail with missing password", func() {
		user := &users.User{
			Username: "Test User",
			Email:    "test@example.com",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "password is required")
	})

	suite.Run("should fail with empty password", func() {
		user := &users.User{
			Username: "Test User",
			Email:    "test@example.com",
			Password: "",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "password is required")
	})
}

func (suite *UserRepositoryTestSuite) TestLogin() {
	ctx := context.Background()

	suite.Run("should login successfully with approved user and correct credentials", func() {
		// Create an approved user first
		user := testutils.CreateUser("loginuser", "mypassword123")
		user.Approved = true // Explicitly approve the user before creation
		status := users.StatusApproved
		user.Status = &status
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to login with correct credentials
		loggedInUser, err := suite.repo.Login(ctx, "loginuser", "mypassword123")
		suite.Require().NoError(err)
		suite.Assert().NotNil(loggedInUser)
		suite.Assert().Equal("loginuser@example.com", loggedInUser.Email)
		suite.Assert().Equal("loginuser", loggedInUser.Username)
		suite.Assert().True(loggedInUser.Approved)
	})

	suite.Run("should fail login for unapproved user", func() {
		// Create an unapproved user (default state)
		user := testutils.CreateUser("unapproveduser", "mypassword123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to login - should fail due to approval status
		loggedInUser, err := suite.repo.Login(ctx, "unapproveduser", "mypassword123")
		suite.Require().Error(err)
		suite.Assert().Nil(loggedInUser)
		suite.Assert().Contains(err.Error(), "account pending approval")
	})

	suite.Run("should fail login for user with inactive status", func() {
		// Create an approved user but with inactive status
		user := testutils.CreateUser("inactiveuser", "mypassword123")
		user.Approved = true // Explicitly approve the user before creation
		inactiveStatus := "inactive"
		user.Status = &inactiveStatus
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to login - should fail due to status
		loggedInUser, err := suite.repo.Login(ctx, "inactiveuser", "mypassword123")
		suite.Require().Error(err)
		suite.Assert().Nil(loggedInUser)
		suite.Assert().Contains(err.Error(), "account not active")
	})

	suite.Run("should fail with incorrect password", func() {
		// Create a user first
		user := testutils.CreateUser("wrongpassuser", "correctpassword")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to login with wrong password
		loggedInUser, err := suite.repo.Login(ctx, "wrongpassuser", "wrongpassword")
		suite.Require().Error(err)
		suite.Assert().Nil(loggedInUser)
		suite.Assert().Contains(err.Error(), "invalid password")
	})

	suite.Run("should fail with non-existent username", func() {
		// Try to login with non-existent username
		loggedInUser, err := suite.repo.Login(ctx, "nonexistentuser", "anypassword")
		suite.Require().Error(err)
		suite.Assert().Nil(loggedInUser)
		suite.Assert().Contains(err.Error(), "failed to get user by username")
	})

}

func (suite *UserRepositoryTestSuite) TestListUsers() {
	ctx := context.Background()

	suite.Run("should return empty list when no users exist", func() {
		users, totalCount, err := suite.repo.ListUsers(ctx, 10, 0)
		suite.Require().NoError(err)
		suite.Assert().Empty(users)
		suite.Assert().Equal(0, totalCount)
	})

	suite.Run("should list users with pagination", func() {
		// Create test users
		user1 := testutils.CreateUser("user1", "password123")
		user2 := testutils.CreateUser("user2", "password123")
		user3 := testutils.CreateUser("user3", "password123")

		err := suite.repo.Create(ctx, user1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, user2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, user3)
		suite.Require().NoError(err)

		// Test first page
		users, totalCount, err := suite.repo.ListUsers(ctx, 2, 0)
		suite.Require().NoError(err)
		suite.Assert().Len(users, 2)
		suite.Assert().Equal(3, totalCount)
		// Should be ordered by created_at DESC (newest first)
		suite.Assert().Equal(user3.Username, users[0].Username)
		suite.Assert().Equal(user2.Username, users[1].Username)

		// Test second page
		users, totalCount, err = suite.repo.ListUsers(ctx, 2, 2)
		suite.Require().NoError(err)
		suite.Assert().Len(users, 1)
		suite.Assert().Equal(3, totalCount)
		suite.Assert().Equal(user1.Username, users[0].Username)
	})

	suite.Run("should validate pagination parameters", func() {
		// Test invalid limit
		_, _, err := suite.repo.ListUsers(ctx, 0, 0)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "limit must be between 1 and 100")

		_, _, err = suite.repo.ListUsers(ctx, 101, 0)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "limit must be between 1 and 100")

		// Test invalid offset
		_, _, err = suite.repo.ListUsers(ctx, 10, -1)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "offset must be non-negative")
	})
}

func (suite *UserRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	suite.Run("should get user by ID successfully", func() {
		user := testutils.CreateUser("testuser", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		retrievedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().Equal(user.ID, retrievedUser.ID)
		suite.Assert().Equal(user.Username, retrievedUser.Username)
		suite.Assert().Equal(user.Email, retrievedUser.Email)
		suite.Assert().Equal(user.Role, retrievedUser.Role)
	})

	suite.Run("should return error for non-existent user", func() {
		_, err := suite.repo.GetByID(ctx, 999999)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "user with ID 999999 not found")
	})
}

func (suite *UserRepositoryTestSuite) TestUpdateStatus() {
	ctx := context.Background()

	suite.Run("should update user status successfully", func() {
		user := testutils.CreateUser("testuser", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Update to approved status
		err = suite.repo.UpdateStatus(ctx, user.ID, users.StatusApproved)
		suite.Require().NoError(err)

		// Verify status was updated
		updatedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().Equal(users.StatusApproved, *updatedUser.Status)
		suite.Assert().True(updatedUser.Approved)
	})

	suite.Run("should update to suspended status", func() {
		user := testutils.CreateUser("testuser2", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Update to suspended status
		err = suite.repo.UpdateStatus(ctx, user.ID, users.StatusSuspended)
		suite.Require().NoError(err)

		// Verify status was updated but approved remains unchanged
		updatedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().Equal(users.StatusSuspended, *updatedUser.Status)
		suite.Assert().False(updatedUser.Approved) // Should remain false
	})

	suite.Run("should validate status value", func() {
		user := testutils.CreateUser("testuser3", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		err = suite.repo.UpdateStatus(ctx, user.ID, "invalid_status")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "invalid status: invalid_status")
	})

	suite.Run("should return error for non-existent user", func() {
		err := suite.repo.UpdateStatus(ctx, 999999, users.StatusApproved)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "user with ID 999999 not found")
	})
}

func (suite *UserRepositoryTestSuite) TestUpdateMultipleStatuses() {
	ctx := context.Background()

	suite.Run("should update multiple users successfully", func() {
		// Create test users
		user1 := testutils.CreateUser("user1", "password123")
		user2 := testutils.CreateUser("user2", "password123")
		user3 := testutils.CreateUser("user3", "password123")

		err := suite.repo.Create(ctx, user1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, user2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, user3)
		suite.Require().NoError(err)

		userIDs := []int64{user1.ID, user2.ID, user3.ID}

		// Update all to approved status
		err = suite.repo.UpdateMultipleStatuses(ctx, userIDs, users.StatusApproved)
		suite.Require().NoError(err)

		// Verify all users were updated
		for _, userID := range userIDs {
			updatedUser, err := suite.repo.GetByID(ctx, userID)
			suite.Require().NoError(err)
			suite.Assert().Equal(users.StatusApproved, *updatedUser.Status)
			suite.Assert().True(updatedUser.Approved)
		}
	})

	suite.Run("should update multiple users to suspended status", func() {
		// Create test users
		user1 := testutils.CreateUser("user4", "password123")
		user2 := testutils.CreateUser("user5", "password123")

		err := suite.repo.Create(ctx, user1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, user2)
		suite.Require().NoError(err)

		userIDs := []int64{user1.ID, user2.ID}

		// Update all to suspended status
		err = suite.repo.UpdateMultipleStatuses(ctx, userIDs, users.StatusSuspended)
		suite.Require().NoError(err)

		// Verify all users were updated
		for _, userID := range userIDs {
			updatedUser, err := suite.repo.GetByID(ctx, userID)
			suite.Require().NoError(err)
			suite.Assert().Equal(users.StatusSuspended, *updatedUser.Status)
			suite.Assert().False(updatedUser.Approved) // Should remain false
		}
	})

	suite.Run("should validate user IDs", func() {
		err := suite.repo.UpdateMultipleStatuses(ctx, []int64{}, users.StatusApproved)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "userIDs cannot be empty")

		err = suite.repo.UpdateMultipleStatuses(ctx, []int64{1, 0, 3}, users.StatusApproved)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "all user IDs must be positive")

		err = suite.repo.UpdateMultipleStatuses(ctx, []int64{1, -2, 3}, users.StatusApproved)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "all user IDs must be positive")
	})

	suite.Run("should validate status value", func() {
		user := testutils.CreateUser("user6", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		err = suite.repo.UpdateMultipleStatuses(ctx, []int64{user.ID}, "invalid_status")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "invalid status: invalid_status")
	})

	suite.Run("should handle partial failure", func() {
		user := testutils.CreateUser("user7", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Mix of valid and invalid user IDs
		userIDs := []int64{user.ID, 999999}

		err = suite.repo.UpdateMultipleStatuses(ctx, userIDs, users.StatusApproved)
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "expected to update 2 users, but updated 1")
	})

	suite.Run("should handle single user update", func() {
		user := testutils.CreateUser("user8", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		err = suite.repo.UpdateMultipleStatuses(ctx, []int64{user.ID}, users.StatusInactive)
		suite.Require().NoError(err)

		updatedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().Equal(users.StatusInactive, *updatedUser.Status)
	})
}

func (suite *UserRepositoryTestSuite) TestUpdatePassword() {
	ctx := context.Background()

	suite.Run("should update password successfully", func() {
		// Create a user with initial password
		user := testutils.CreateUser("passworduser", "oldpassword123")
		user.Approved = true
		status := users.StatusApproved
		user.Status = &status
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Update password
		err = suite.repo.UpdatePassword(ctx, user.ID, "newpassword456")
		suite.Require().NoError(err)

		// Verify new password works by logging in
		loggedInUser, err := suite.repo.Login(ctx, "passworduser", "newpassword456")
		suite.Require().NoError(err)
		suite.Assert().NotNil(loggedInUser)
		suite.Assert().Equal(user.ID, loggedInUser.ID)
	})

	suite.Run("should fail login with old password after update", func() {
		// Create a user with initial password
		user := testutils.CreateUser("passworduser2", "oldpassword123")
		user.Approved = true
		status := users.StatusApproved
		user.Status = &status
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Update password
		err = suite.repo.UpdatePassword(ctx, user.ID, "newpassword456")
		suite.Require().NoError(err)

		// Verify old password no longer works
		_, err = suite.repo.Login(ctx, "passworduser2", "oldpassword123")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "invalid password")
	})

	suite.Run("should update updated_at timestamp", func() {
		// Create a user
		user := testutils.CreateUser("timestampuser", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		originalUpdatedAt := user.UpdatedAt

		// Update password
		err = suite.repo.UpdatePassword(ctx, user.ID, "newpassword456")
		suite.Require().NoError(err)

		// Get updated user and verify timestamp changed
		updatedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().NotNil(updatedUser.UpdatedAt)
		suite.Assert().True(updatedUser.UpdatedAt.After(*originalUpdatedAt) || updatedUser.UpdatedAt.Equal(*originalUpdatedAt))
	})

	suite.Run("should generate new salt on password update", func() {
		// Create a user
		user := testutils.CreateUser("saltuser", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		originalSalt := user.Salt
		originalPassword := user.Password

		// Update password
		err = suite.repo.UpdatePassword(ctx, user.ID, "newpassword456")
		suite.Require().NoError(err)

		// Get updated user and verify salt and password hash changed
		updatedUser, err := suite.repo.GetByID(ctx, user.ID)
		suite.Require().NoError(err)
		suite.Assert().NotEqual(originalSalt, updatedUser.Salt)
		suite.Assert().NotEqual(originalPassword, updatedUser.Password)
	})

	suite.Run("should fail with invalid user ID", func() {
		err := suite.repo.UpdatePassword(ctx, 0, "newpassword")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "user ID must be positive")

		err = suite.repo.UpdatePassword(ctx, -1, "newpassword")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "user ID must be positive")
	})

	suite.Run("should fail with empty password", func() {
		user := testutils.CreateUser("emptypassuser", "password123")
		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		err = suite.repo.UpdatePassword(ctx, user.ID, "")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "password cannot be empty")
	})

	suite.Run("should fail for non-existent user", func() {
		err := suite.repo.UpdatePassword(ctx, 999999, "newpassword")
		suite.Assert().Error(err)
		suite.Assert().Contains(err.Error(), "user with ID 999999 not found")
	})
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
