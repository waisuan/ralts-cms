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

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
