package handlers

import (
	"encoding/json"
	"net/http"
	"ralts-cms/internal/deps"
	"time"
)

type HealthHandler struct {
	deps *deps.Dependencies
}

func NewHealthHandler(deps *deps.Dependencies) *HealthHandler {
	return &HealthHandler{
		deps: deps,
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "ralts-cms",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
