package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// UsersHandler handles HTTP requests for user-related operations
type UsersHandler struct {
	deps *deps.Dependencies
}

// CreateUserRequest represents the request structure for user creation
type CreateUserRequest struct {
	Username string  `json:"username" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Role     string  `json:"role,omitempty"`
	Approved bool    `json:"approved,omitempty"`
	Status   string  `json:"status,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
}

// LoginRequest represents the request structure for user authentication
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateUserStatusRequest represents the request structure for updating user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// BulkUpdateStatusRequest represents the request structure for bulk status updates
type BulkUpdateStatusRequest struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1"`
	Status  string  `json:"status" validate:"required"`
}

// ListUsersResponse represents the response structure for user listing
type ListUsersResponse struct {
	Users      []*users.User `json:"users"`
	TotalCount int           `json:"total_count"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// NewUsersHandler creates a new users handler instance with the given dependencies
func NewUsersHandler(deps *deps.Dependencies) *UsersHandler {
	return &UsersHandler{
		deps: deps,
	}
}

// CreateUser handles POST /api/v1/users for creating new users
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserRequest CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create user from request
	var status *string
	if createUserRequest.Status != "" {
		status = &createUserRequest.Status
	}

	user := &users.User{
		Username: createUserRequest.Username,
		Email:    createUserRequest.Email,
		Password: createUserRequest.Password,
		Role:     createUserRequest.Role,
		Approved: createUserRequest.Approved,
		Status:   status,
		Avatar:   createUserRequest.Avatar,
	}

	if err := h.deps.UsersRepository.Create(r.Context(), user); err != nil {
		// Check for duplicate key constraint violations
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "A user with this username or email already exists. Please try different credentials.", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login handles POST /api/v1/users/login for user authentication
func (h *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	user, err := h.deps.UsersRepository.Login(r.Context(), loginRequest.Username, loginRequest.Password)
	if err != nil {
		// Check if it's an approval-related error and provide specific message
		errMsg := err.Error()
		if errMsg == "account pending approval: your account is awaiting administrator approval" {
			http.Error(w, "Account Pending Approval: Your account is awaiting administrator approval. Please contact an administrator to activate your account.", http.StatusForbidden)
			return
		}
		if errMsg == "account not active: your account status does not allow login" {
			http.Error(w, "Account Not Active: Your account status does not allow login. Please contact an administrator.", http.StatusForbidden)
			return
		}

		// Generic login failure message for other errors (password, username not found, etc.)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token with user role
	tokenString, err := auth.GenerateJWTToken(int(user.ID), user.Role, h.deps.Config.JWTSecret)
	if err != nil {
		h.deps.Logger.Error("Failed to generate JWT token",
			"error", err,
			"user_id", user.ID)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Create response with user data and token
	response := struct {
		User  *users.User `json:"user"`
		Token string      `json:"token"`
	}{
		User:  user,
		Token: tokenString,
	}

	// Log successful login
	h.deps.Logger.Info("User login successful",
		"user_id", user.ID,
		"username", user.Username,
		"email", user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListUsers handles GET /api/v1/admin/users for listing all users with pagination
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 50
	offset := 0

	// Parse limit if provided
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			http.Error(w, "Invalid limit parameter: must be between 1 and 100", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	// Parse offset if provided
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			http.Error(w, "Invalid offset parameter: must be non-negative", http.StatusBadRequest)
			return
		}
		offset = parsedOffset
	}

	// Fetch users from repository
	userList, totalCount, err := h.deps.UsersRepository.ListUsers(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// Create response
	response := ListUsersResponse{
		Users:      userList,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateUserStatus handles PUT /api/v1/admin/users/{id}/status for updating a single user's status
func (h *UsersHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path variables
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing user ID in URL path", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID: must be a positive integer", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateRequest UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate status value
	if err := users.ValidateStatusValue(updateRequest.Status); err != nil {
		http.Error(w, fmt.Sprintf("Invalid status: %v", err), http.StatusBadRequest)
		return
	}

	// Update user status
	if err := h.deps.UsersRepository.UpdateStatus(r.Context(), userID, updateRequest.Status); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user status: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "User status updated successfully",
		"user_id": userID,
		"status":  updateRequest.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// BulkUpdateStatus handles PUT /api/v1/admin/users/bulk-status for updating multiple users' status
func (h *UsersHandler) BulkUpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var bulkRequest BulkUpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&bulkRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(bulkRequest.UserIDs) == 0 {
		http.Error(w, "User IDs list cannot be empty", http.StatusBadRequest)
		return
	}

	if len(bulkRequest.UserIDs) > 100 {
		http.Error(w, "Cannot update more than 100 users at once", http.StatusBadRequest)
		return
	}

	// Validate all user IDs
	for _, userID := range bulkRequest.UserIDs {
		if userID <= 0 {
			http.Error(w, "All user IDs must be positive integers", http.StatusBadRequest)
			return
		}
	}

	// Validate status value
	if err := users.ValidateStatusValue(bulkRequest.Status); err != nil {
		http.Error(w, fmt.Sprintf("Invalid status: %v", err), http.StatusBadRequest)
		return
	}

	// Update multiple user statuses
	if err := h.deps.UsersRepository.UpdateMultipleStatuses(r.Context(), bulkRequest.UserIDs, bulkRequest.Status); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user statuses: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message":        "User statuses updated successfully",
		"updated_count":  len(bulkRequest.UserIDs),
		"status":         bulkRequest.Status,
		"affected_users": bulkRequest.UserIDs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
