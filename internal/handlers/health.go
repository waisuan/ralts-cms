// Package handlers provides HTTP request handlers for the Ralts-CMS API,
// including machine, maintenance, user, and health check endpoints.
package handlers

import (
	"encoding/json"
	"net/http"
	"ralts-cms/internal/deps"
	"time"
)

// HealthHandler handles HTTP requests for health check endpoints
type HealthHandler struct {
	deps *deps.Dependencies
}

// NewHealthHandler creates a new health handler instance with the given dependencies
func NewHealthHandler(deps *deps.Dependencies) *HealthHandler {
	return &HealthHandler{
		deps: deps,
	}
}

// HealthResponse represents the response structure for health check endpoints
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "ralts-cms",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
