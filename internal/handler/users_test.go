package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handler"
	"ralts-cms/internal/users"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// UsersHandlerTestSuite defines the test suite for user handler
type UsersHandlerTestSuite struct {
	suite.Suite

	handler  *handler.UsersHandler
	mockRepo *users.MockRepository
	ctrl     *gomock.Controller
}

// SetupTest sets up each test
func (suite *UsersHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = users.NewMockRepository(suite.ctrl)
	deps := &deps.Dependencies{
		UsersRepository: suite.mockRepo,
	}
	suite.handler = handler.NewUsersHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *UsersHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *UsersHandlerTestSuite) TestCreateUser() {
	suite.Run("should create user successfully", func() {
		userData := users.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "mypassword123",
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
		userData := users.User{
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

func TestUsersHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UsersHandlerTestSuite))
}
