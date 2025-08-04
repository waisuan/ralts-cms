package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Role     string  `json:"role,omitempty"`
	Status   string  `json:"status,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
}

// LoginRequest represents the request structure for user authentication
type LoginRequest struct {
	Email    string `json:"email"`
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
	user := &users.User{
		Name:     createUserRequest.Name,
		Email:    createUserRequest.Email,
		Password: createUserRequest.Password,
		Role:     createUserRequest.Role,
		Status:   createUserRequest.Status,
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

	user, err := h.deps.UsersRepository.Login(r.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		http.Error(w, "Failed to login", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	tokenString, err := auth.GenerateJWTToken(user.ID, h.deps.Config.JWTSecret)
	if err != nil {
		log.Println("Failed to generate token: ", err)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
