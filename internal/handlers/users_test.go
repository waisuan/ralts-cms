package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/middlewares"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/suite"
)

// UsersHandlerTestSuite defines the test suite for user handler
type UsersHandlerTestSuite struct {
	suite.Suite

	handler          *handlers.UsersHandler
	mockRepo         *users.MockRepository
	mockAuditService *audit.MockAuditService
	ctrl             *gomock.Controller
}

// SetupTest sets up each test
func (suite *UsersHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = users.NewMockRepository(suite.ctrl)
	suite.mockAuditService = audit.NewMockAuditService(suite.ctrl)

	// Allow any audit events to be logged
	suite.mockAuditService.EXPECT().LogEvent(gomock.Any()).AnyTimes()

	deps := &deps.Dependencies{
		Config: &deps.Config{
			JWTSecret: "your-jwt-secret-key",
		},
		Logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
		UsersRepository: suite.mockRepo,
		AuditService:    suite.mockAuditService,
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
			Username: "Test User",
			Email:    "test@example.com",
			Password: "mypassword123",
			Role:     "user",
			Approved: false,
			Status:   "active",
			Avatar:   nil,
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *users.User) error {
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
		suite.Assert().Equal("Test User", response.Username)
		suite.Assert().Equal("test@example.com", response.Email)
		suite.Assert().Empty(response.Password)
		suite.Assert().Empty(response.Salt)
		suite.Assert().Equal(int64(1), response.ID)
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
			Username: "Error User",
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

	suite.Run("should return 409 when user already exists", func() {
		userData := handlers.CreateUserRequest{
			Username: "duplicate_user",
			Email:    "duplicate@example.com",
			Password: "mypassword123",
		}

		// Mock repository to return a PostgreSQL unique_violation (23505)
		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(
			&pgconn.PgError{Code: "23505"})

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateUser(w, req)

		suite.Assert().Equal(http.StatusConflict, w.Code)
		suite.Assert().Contains(w.Body.String(), "A user with this username or email already exists. Please try different credentials.")
	})
}

func (suite *UsersHandlerTestSuite) TestLogin() {
	suite.Run("should login successfully with valid credentials", func() {
		loginRequest := handlers.LoginRequest{
			Username: "test",
			Password: "mypassword123",
		}

		status := "active"
		expectedUser := &users.User{
			ID:       1,
			Username: "test",
			Email:    "test@example.com",
			Role:     "user",
			Status:   &status,
			Password: "hashedpassword",
			Salt:     "salt123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "test", "mypassword123").Return(expectedUser, nil)

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
		suite.Assert().Equal(int64(1), response.User.ID)
		suite.Assert().Equal("test", response.User.Username)
		suite.Assert().Equal("test@example.com", response.User.Email)
		suite.Assert().Equal("user", response.User.Role)
		suite.Assert().NotNil(response.User.Status)
		suite.Assert().Equal("active", *response.User.Status)
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
			Username: "nonexistent",
			Password: "mypassword123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "nonexistent", "mypassword123").Return(nil, fmt.Errorf("invalid password"))

		body, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid username or password")
	})

	suite.Run("should return 403 when user is not approved", func() {
		loginRequest := handlers.LoginRequest{
			Username: "unapproved",
			Password: "mypassword123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "unapproved", "mypassword123").Return(nil, fmt.Errorf("account pending approval: your account is awaiting administrator approval"))

		body, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusForbidden, w.Code)
		suite.Assert().Contains(w.Body.String(), "Account Pending Approval")
	})

	suite.Run("should return 403 when user account is not active", func() {
		loginRequest := handlers.LoginRequest{
			Username: "inactive",
			Password: "mypassword123",
		}

		suite.mockRepo.EXPECT().Login(gomock.Any(), "inactive", "mypassword123").Return(nil, fmt.Errorf("account not active: your account status does not allow login"))

		body, _ := json.Marshal(loginRequest)
		req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.Login(w, req)

		suite.Assert().Equal(http.StatusForbidden, w.Code)
		suite.Assert().Contains(w.Body.String(), "Account Not Active")
	})
}

// TestListUsers tests the admin ListUsers endpoint
func (suite *UsersHandlerTestSuite) TestListUsers() {
	suite.Run("should list users successfully with default pagination", func() {
		// Create test users
		testUsers := []*users.User{
			{
				ID:       1,
				Username: "user1",
				Email:    "user1@example.com",
				Role:     users.RoleNonAdmin,
				Status:   &[]string{users.StatusApproved}[0],
			},
			{
				ID:       2,
				Username: "user2",
				Email:    "user2@example.com",
				Role:     users.RoleNonAdmin,
				Status:   &[]string{users.StatusPendingApproval}[0],
			},
		}

		// Setup mock expectations
		suite.mockRepo.EXPECT().
			ListUsers(gomock.Any(), 50, 0).
			Return(testUsers, 2, nil)

		// Create request with admin user context
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		suite.handler.ListUsers(rr, req)

		// Assertions
		suite.Equal(http.StatusOK, rr.Code)

		var response handlers.ListUsersResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		suite.NoError(err)

		suite.Equal(2, response.TotalCount)
		suite.Equal(50, response.Limit)
		suite.Equal(0, response.Offset)
		suite.Len(response.Users, 2)
		suite.Equal("user1", response.Users[0].Username)
		suite.Equal("user2", response.Users[1].Username)
	})

	suite.Run("should handle custom pagination parameters", func() {
		testUsers := []*users.User{
			{ID: 3, Username: "user3", Email: "user3@example.com"},
		}

		suite.mockRepo.EXPECT().
			ListUsers(gomock.Any(), 25, 10).
			Return(testUsers, 1, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?limit=25&offset=10", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.ListUsers(rr, req)

		suite.Equal(http.StatusOK, rr.Code)

		var response handlers.ListUsersResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		suite.NoError(err)

		suite.Equal(25, response.Limit)
		suite.Equal(10, response.Offset)
	})

	suite.Run("should return 400 for invalid limit", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?limit=150", nil)
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.ListUsers(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockRepo.EXPECT().
			ListUsers(gomock.Any(), 50, 0).
			Return(nil, 0, fmt.Errorf("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		rr := httptest.NewRecorder()

		suite.handler.ListUsers(rr, req)

		suite.Equal(http.StatusInternalServerError, rr.Code)
		suite.Contains(rr.Body.String(), "Failed to retrieve users")
	})
}

// TestUpdateUserStatus tests the admin UpdateUserStatus endpoint
func (suite *UsersHandlerTestSuite) TestUpdateUserStatus() {
	suite.Run("should update user status successfully", func() {
		// Setup mock expectations
		suite.mockRepo.EXPECT().
			UpdateStatus(gomock.Any(), int64(2), users.StatusApproved).
			Return(nil)

		// Create request body
		requestBody := handlers.UpdateUserStatusRequest{
			Status: users.StatusApproved,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Create request with admin user context and mux vars
		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2/status", bytes.NewReader(bodyBytes))
		req = mux.SetURLVars(req, map[string]string{"id": "2"})
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		suite.handler.UpdateUserStatus(rr, req)

		// Assertions
		suite.Equal(http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		suite.NoError(err)

		suite.Equal("User status updated successfully", response["message"])
		suite.Equal(float64(2), response["user_id"])
		suite.Equal(users.StatusApproved, response["status"])
	})

	suite.Run("should return 400 for invalid user ID", func() {
		requestBody := handlers.UpdateUserStatusRequest{
			Status: users.StatusApproved,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/invalid/status", bytes.NewReader(bodyBytes))
		req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.UpdateUserStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Invalid user ID")
	})

	suite.Run("should return 400 for invalid status", func() {
		requestBody := handlers.UpdateUserStatusRequest{
			Status: "invalid_status",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2/status", bytes.NewReader(bodyBytes))
		req = mux.SetURLVars(req, map[string]string{"id": "2"})
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.UpdateUserStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Invalid status")
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockRepo.EXPECT().
			UpdateStatus(gomock.Any(), int64(999), users.StatusApproved).
			Return(fmt.Errorf("user with ID 999 not found"))

		requestBody := handlers.UpdateUserStatusRequest{
			Status: users.StatusApproved,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/999/status", bytes.NewReader(bodyBytes))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.UpdateUserStatus(rr, req)

		suite.Equal(http.StatusInternalServerError, rr.Code)
		suite.Contains(rr.Body.String(), "Failed to update user status")
	})
}

// TestBulkUpdateStatus tests the admin BulkUpdateStatus endpoint
func (suite *UsersHandlerTestSuite) TestBulkUpdateStatus() {
	suite.Run("should update multiple user statuses successfully", func() {
		userIDs := []int64{2, 3, 4}

		// Setup mock expectations
		suite.mockRepo.EXPECT().
			UpdateMultipleStatuses(gomock.Any(), userIDs, users.StatusSuspended).
			Return(nil)

		// Create request body
		requestBody := handlers.BulkUpdateStatusRequest{
			UserIDs: userIDs,
			Status:  users.StatusSuspended,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Create request with admin user context
		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/bulk-status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		suite.handler.BulkUpdateStatus(rr, req)

		// Assertions
		suite.Equal(http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		suite.NoError(err)

		suite.Equal("User statuses updated successfully", response["message"])
		suite.Equal(float64(3), response["updated_count"])
		suite.Equal(users.StatusSuspended, response["status"])
	})

	suite.Run("should return 400 for empty user IDs", func() {
		requestBody := handlers.BulkUpdateStatusRequest{
			UserIDs: []int64{},
			Status:  users.StatusSuspended,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/bulk-status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.BulkUpdateStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "User IDs list cannot be empty")
	})

	suite.Run("should return 400 for too many user IDs", func() {
		// Create 101 user IDs
		userIDs := make([]int64, 101)
		for i := 0; i < 101; i++ {
			userIDs[i] = int64(i + 1)
		}

		requestBody := handlers.BulkUpdateStatusRequest{
			UserIDs: userIDs,
			Status:  users.StatusSuspended,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/bulk-status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.BulkUpdateStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Cannot update more than 100 users at once")
	})

	suite.Run("should return 400 for invalid user IDs", func() {
		requestBody := handlers.BulkUpdateStatusRequest{
			UserIDs: []int64{1, -1, 3}, // -1 is invalid
			Status:  users.StatusSuspended,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/bulk-status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.BulkUpdateStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "All user IDs must be positive integers")
	})

	suite.Run("should return 400 for invalid status", func() {
		requestBody := handlers.BulkUpdateStatusRequest{
			UserIDs: []int64{1, 2, 3},
			Status:  "invalid_status",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/bulk-status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleAdmin,
		}))

		rr := httptest.NewRecorder()
		suite.handler.BulkUpdateStatus(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Invalid status")
	})
}

// TestUpdatePassword tests the UpdatePassword endpoint
func (suite *UsersHandlerTestSuite) TestUpdatePassword() {
	suite.Run("should update password successfully", func() {
		// Generate real password hash for testing
		salt, err := auth.GenerateSalt()
		suite.Require().NoError(err)
		hashedPassword, err := auth.HashPassword("oldpassword123", salt)
		suite.Require().NoError(err)

		// Create mock user data for GetByID with real hash
		existingUser := &users.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Password: hashedPassword,
			Salt:     salt,
		}

		// Setup mock expectations
		suite.mockRepo.EXPECT().
			GetByID(gomock.Any(), int64(1)).
			Return(existingUser, nil)
		suite.mockRepo.EXPECT().
			UpdatePassword(gomock.Any(), int64(1), "newpassword123").
			Return(nil)

		// Create request body
		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Create request with authenticated user context
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusOK, rr.Code)

		var response map[string]string
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		suite.NoError(err)
		suite.Equal("Password updated successfully", response["message"])
	})

	suite.Run("should return 401 when user context is missing", func() {
		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		// Create request without user context
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusUnauthorized, rr.Code)
		suite.Contains(rr.Body.String(), "Unauthorized")
	})

	suite.Run("should return 400 for invalid request body", func() {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewBufferString("invalid json"))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Invalid request body")
	})

	suite.Run("should return 400 when current password is empty", func() {
		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "Current password is required")
	})

	suite.Run("should return 400 when new password is too short", func() {
		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "short",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusBadRequest, rr.Code)
		suite.Contains(rr.Body.String(), "New password must be at least 6 characters")
	})

	suite.Run("should return 404 when user not found", func() {
		suite.mockRepo.EXPECT().
			GetByID(gomock.Any(), int64(999)).
			Return(nil, fmt.Errorf("user with ID 999 not found"))

		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 999,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusNotFound, rr.Code)
		suite.Contains(rr.Body.String(), "User not found")
	})

	suite.Run("should return 401 when current password is incorrect", func() {
		// Create mock user with specific password/salt that will fail verification
		existingUser := &users.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Password: "correcthashedpassword",
			Salt:     "salt123",
		}

		suite.mockRepo.EXPECT().
			GetByID(gomock.Any(), int64(1)).
			Return(existingUser, nil)

		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusUnauthorized, rr.Code)
		suite.Contains(rr.Body.String(), "Current password is incorrect")
	})

	suite.Run("should return 500 when repository update fails", func() {
		// Generate real password hash for testing
		salt, err := auth.GenerateSalt()
		suite.Require().NoError(err)
		hashedPassword, err := auth.HashPassword("oldpassword123", salt)
		suite.Require().NoError(err)

		// Create mock user data for GetByID with real hash
		existingUser := &users.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Password: hashedPassword,
			Salt:     salt,
		}

		suite.mockRepo.EXPECT().
			GetByID(gomock.Any(), int64(1)).
			Return(existingUser, nil)
		suite.mockRepo.EXPECT().
			UpdatePassword(gomock.Any(), int64(1), "newpassword123").
			Return(fmt.Errorf("database error"))

		requestBody := handlers.UpdatePasswordRequest{
			CurrentPassword: "oldpassword123",
			NewPassword:     "newpassword123",
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/password", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, &middlewares.UserContext{
			UserID: 1,
			Role:   users.RoleNonAdmin,
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.handler.UpdatePassword(rr, req)

		suite.Equal(http.StatusInternalServerError, rr.Code)
		suite.Contains(rr.Body.String(), "Failed to update password")
	})
}

func TestUsersHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UsersHandlerTestSuite))
}
