package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"
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
