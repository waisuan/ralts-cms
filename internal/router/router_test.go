package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/machine"
)

type mockMachineRepo struct{ machine.Repository }

func (m *mockMachineRepo) GetBySerialNumber(_ context.Context, _ string) (*machine.Machine, error) {
	return &machine.Machine{}, nil
}
func (m *mockMachineRepo) Create(_ context.Context, _ *machine.Machine) error { return nil }
func (m *mockMachineRepo) Update(_ context.Context, _ *machine.Machine) error { return nil }
func (m *mockMachineRepo) Delete(_ context.Context, _ string) error           { return nil }

func TestNewRouter(t *testing.T) {
	// Create mock dependencies
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
	}

	// Create router
	router := NewRouter(mockDeps)
	require.NotNil(t, router)

	// Test that router is an http.Handler
	_, ok := router.(http.Handler)
	assert.True(t, ok)
}

func TestRouter_HealthEndpoint(t *testing.T) {
	// Create mock dependencies
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test health endpoint (should not require auth)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Health endpoint should return 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

func TestRouter_ProtectedEndpoints(t *testing.T) {
	// Create mock dependencies with a mock MachineRepository
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
		MachineRepository: &mockMachineRepo{},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test protected endpoints without auth (should return 401)
	protectedEndpoints := []string{
		"/machines/MACHINE123",
		"/machines/MACHINE123/maintenance",
		"/machines/MACHINE123/maintenance/WO123",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Should return 401 Unauthorized without auth
			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestRouter_ProtectedEndpointsWithAuth(t *testing.T) {
	// Create mock dependencies with a mock MachineRepository
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
		MachineRepository: &mockMachineRepo{},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test protected endpoints with auth
	req := httptest.NewRequest(http.MethodGet, "/machines/MACHINE123", nil)
	req.Header.Set("Authorization", "Basic YWRtaW46cGFzc3dvcmQ=") // admin:password in base64
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should not return 401 (auth passed), but might return 404 or other status
	assert.NotEqual(t, http.StatusUnauthorized, rr.Code)
}

func TestRouter_InvalidAuth(t *testing.T) {
	// Create mock dependencies with a mock MachineRepository
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
		MachineRepository: &mockMachineRepo{},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test with invalid auth
	req := httptest.NewRequest(http.MethodGet, "/machines/MACHINE123", nil)
	req.Header.Set("Authorization", "Basic d3Jvbmc6Y3JlZGVudGlhbHM=") // wrong:credentials in base64
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRouter_InvalidEndpoint(t *testing.T) {
	// Create mock dependencies
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test invalid endpoint
	req := httptest.NewRequest(http.MethodGet, "/invalid/endpoint", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should return 404 Not Found
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRouter_OptionsRequest(t *testing.T) {
	// Create mock dependencies with a mock MachineRepository
	mockDeps := &deps.Dependencies{
		Config: &deps.Config{
			Credentials: "admin:password",
		},
		MachineRepository: &mockMachineRepo{},
	}

	// Create router
	router := NewRouter(mockDeps)

	// Test OPTIONS request (CORS preflight)
	req := httptest.NewRequest(http.MethodOptions, "/machines/MACHINE123", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should return 200 OK for OPTIONS requests
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}
