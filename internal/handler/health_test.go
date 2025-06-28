package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handler"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Health(t *testing.T) {
	// Create mock dependencies
	mockDeps := &deps.Dependencies{}

	// Create handler
	healthHandler := handler.NewHealthHandler(mockDeps)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	healthHandler.Health(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse response body
	var response handler.HealthResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check response fields
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "ralts-cms", response.Service)
	assert.NotEmpty(t, response.Timestamp)
}

func TestHealthHandler_NewHealthHandler(t *testing.T) {
	// Create mock dependencies
	mockDeps := &deps.Dependencies{}

	// Create handler
	healthHandler := handler.NewHealthHandler(mockDeps)

	// Check handler is not nil
	assert.NotNil(t, healthHandler)
}

func TestHealthResponse_Structure(t *testing.T) {
	// Test the HealthResponse struct can be marshaled and unmarshaled
	response := handler.HealthResponse{
		Status:    "healthy",
		Service:   "ralts-cms",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled handler.HealthResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Check fields match
	assert.Equal(t, response.Status, unmarshaled.Status)
	assert.Equal(t, response.Service, unmarshaled.Service)
	assert.Equal(t, response.Timestamp, unmarshaled.Timestamp)
}
