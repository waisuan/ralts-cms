package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/users"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// UsersHandlerTestSuite defines the test suite for user handler
type UsersHandlerTestSuite struct {
	suite.Suite

	handler  *handlers.UsersHandler
	mockRepo *users.MockRepository
	ctrl     *gomock.Controller
}

// SetupTest sets up each test
func (suite *UsersHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = users.NewMockRepository(suite.ctrl)
	deps := &deps.Dependencies{
		Config: &deps.Config{
			JWTSecret: "your-jwt-secret-key",
		},
		UsersRepository: suite.mockRepo,
	}
	suite.handler = handlers.NewUsersHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *UsersHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *UsersHandlerTestSuite) TestCreateUser() {
	suite.Run("should create user successfully", func() {
		userData := handlers.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "mypassword123",
			Role:     "user",
			Status:   "active",
			Avatar:   nil,
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u *users.User) error {
			u.ID = 1
			return nil
		})

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateUser(w, req)

		suite.Assert().Equal(http.StatusCreated, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response users.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("Test User", response.Name)
		suite.Assert().Equal("test@example.com", response.Email)
		suite.Assert().Empty(response.Password)
		suite.Assert().Empty(response.Salt)
		suite.Assert().Equal(1, response.ID)
	})

	suite.Run("should return 400 when request body is invalid", func() {
		req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateUser(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})

	suite.Run("should return 500 on repository error", func() {
		userData := handlers.CreateUserRequest{
			Name:     "Error User",
			Email:    "error@example.com",
			Password: "mypassword123",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database connection failed"))

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateUser(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to create user")
		suite.Assert().Contains(w.Body.String(), "database connection failed")
	})

	suite.Run("should handle empty request body", func() {
		req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(""))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateUser(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})
}

func (suite *UsersHandlerTestSuite) TestLogin() {
	suite.Run("should login successfully with valid credentials", func() {
		loginRequest := handlers.LoginRequest{
			Email:    "test@example.com",
			Password: "mypassword123",
		}

		expectedUser := &users.User{
			ID:       1,
			Name:     "Test User",
			Email:    "test@example.com",
			Role:     "user",
			Status:   "active",
			Password: "hashedpassword",
			Salt:     "salt123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "test@example.com", "mypassword123").Return(expectedUser, nil)

		body, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response struct {
			User  *users.User `json:"user"`
			Token string      `json:"token"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		// Verify user data
		suite.Assert().NotNil(response.User)
		suite.Assert().Equal(1, response.User.ID)
		suite.Assert().Equal("Test User", response.User.Name)
		suite.Assert().Equal("test@example.com", response.User.Email)
		suite.Assert().Equal("user", response.User.Role)
		suite.Assert().Equal("active", response.User.Status)
		// Password and Salt should be empty in response (json:"-")
		suite.Assert().Empty(response.User.Password)
		suite.Assert().Empty(response.User.Salt)

		// Verify token is present
		suite.Assert().NotEmpty(response.Token)
	})

	suite.Run("should return 400 when request body is invalid", func() {
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})

	suite.Run("should return 401 when user cannot be authenticated", func() {
		loginRequest := handlers.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "mypassword123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "nonexistent@example.com", "mypassword123").Return(nil, fmt.Errorf("invalid password"))

		body, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to login")
	})
}

func TestUsersHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UsersHandlerTestSuite))
}
