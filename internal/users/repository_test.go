package users_test

import (
	"context"
	"testing"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/testutils"
	"ralts-cms/internal/users"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryTestSuite defines the test suite for user repository
type UserRepositoryTestSuite struct {
	suite.Suite

	deps *deps.Dependencies
	repo users.Repository
}

// SetupTest sets up each test
func (suite *UserRepositoryTestSuite) SetupTest() {
	deps := deps.Initialise()
	suite.deps = deps
	suite.repo = users.NewRepository(deps.PostgresClient)
}

func (suite *UserRepositoryTestSuite) TearDownSubTest() {
	// Clear the users table for PostgreSQL
	ctx := context.Background()
	_, err := suite.deps.PostgresClient.Exec(ctx, "DELETE FROM users")
	require.NoError(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create user successfully", func() {
		user := testutils.CreateUser("test@example.com", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(user.CreatedAt)
		suite.Assert().NotEmpty(user.UpdatedAt)
		suite.Assert().Equal(user.CreatedAt, user.UpdatedAt)
		suite.Assert().Greater(user.ID, 0)          // PostgreSQL should return an ID
		suite.Assert().Equal("user", user.Role)     // Default role
		suite.Assert().Equal("active", user.Status) // Default status
		suite.Assert().NotEmpty(user.Password)      // Should be hashed
		suite.Assert().NotEmpty(user.Salt)          // Should be set
	})

	suite.Run("should fail when creating duplicate user", func() {
		user := testutils.CreateUser("duplicate@example.com", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Try to create the same user again
		duplicateUser := testutils.CreateUser("duplicate@example.com", "differentpassword")
		err = suite.repo.Create(ctx, duplicateUser)
		suite.Require().Error(err)
		// PostgreSQL will return a unique constraint violation error
		suite.Assert().Contains(err.Error(), "duplicate key")
	})

	suite.Run("should create user with custom role and status", func() {
		user := testutils.CreateUserWithCustomFields("admin@example.com", "adminpass", "admin", "inactive")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().Equal("admin", user.Role)
		suite.Assert().Equal("inactive", user.Status)
		suite.Assert().NotEmpty(user.CreatedAt)
		suite.Assert().NotEmpty(user.UpdatedAt)
	})

	suite.Run("should create user with avatar", func() {
		avatar := "https://example.com/avatar.jpg"
		user := testutils.CreateUserWithAvatar("avatar@example.com", "mypassword123", &avatar)

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().NotNil(user.Avatar)
		suite.Assert().Equal(avatar, *user.Avatar)
	})

	suite.Run("should create user with nil avatar", func() {
		user := testutils.CreateUserWithAvatar("noavatar@example.com", "mypassword123", nil)

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().Nil(user.Avatar)
	})

	suite.Run("should hash password correctly", func() {
		user := testutils.CreateUser("password@example.com", "mypassword123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		// Password should be hashed (not plain text)
		suite.Assert().NotEqual("mypassword123", user.Password)
		suite.Assert().NotEmpty(user.Password)
		suite.Assert().NotEmpty(user.Salt)
		// bcrypt hashes are long
		suite.Assert().GreaterOrEqual(len(user.Password), 60)
	})

	suite.Run("should handle complex passwords", func() {
		user := testutils.CreateUser("complex@example.com", "P@ssw0rd!@#$%^&*()")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().NotEqual("P@ssw0rd!@#$%^&*()", user.Password)
		suite.Assert().NotEmpty(user.Password)
		suite.Assert().NotEmpty(user.Salt)
	})

	suite.Run("should handle unicode passwords", func() {
		user := testutils.CreateUser("unicode@example.com", "密码123")

		err := suite.repo.Create(ctx, user)
		suite.Require().NoError(err)

		suite.Assert().NotEqual("密码123", user.Password)
		suite.Assert().NotEmpty(user.Password)
		suite.Assert().NotEmpty(user.Salt)
	})
}

func (suite *UserRepositoryTestSuite) TestCreate_Validation() {
	ctx := context.Background()

	suite.Run("should fail with missing name", func() {
		user := &users.User{
			Email:    "test@example.com",
			Password: "mypassword123",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "name is required")
	})

	suite.Run("should fail with missing email", func() {
		user := &users.User{
			Name:     "Test User",
			Password: "mypassword123",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "email is required")
	})

	suite.Run("should fail with missing password", func() {
		user := &users.User{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "password is required")
	})

	suite.Run("should fail with empty password", func() {
		user := &users.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "",
		}

		err := suite.repo.Create(ctx, user)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "password is required")
	})
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
