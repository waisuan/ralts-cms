package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/users"
)

type UsersHandler struct {
	deps *deps.Dependencies
}

func NewUsersHandler(deps *deps.Dependencies) *UsersHandler {
	return &UsersHandler{
		deps: deps,
	}
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.deps.UsersRepository.Create(r.Context(), &user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
