package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			rr := httptest.NewRecorder()

			// Apply middleware
			CORSMiddleware(handler).ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check headers
			for header, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(header))
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	// Apply middleware
	LoggingMiddleware(handler).ServeHTTP(rr, req)

	// Check that the request was processed
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := "test-secret-key"

	tests := []struct {
		name           string
		jwtSecret      string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Valid JWT token",
			jwtSecret:      jwtSecret,
			authHeader:     "Bearer " + createValidJWT(jwtSecret),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing authorization header",
			jwtSecret:      jwtSecret,
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid authorization format",
			jwtSecret:      jwtSecret,
			authHeader:     "Basic dGVzdDp0ZXN0",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty bearer token",
			jwtSecret:      jwtSecret,
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid JWT token",
			jwtSecret:      jwtSecret,
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired JWT token",
			jwtSecret:      jwtSecret,
			authHeader:     "Bearer " + createExpiredJWT(jwtSecret),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := AuthMiddleware(tt.jwtSecret)

			// Create a simple handler that returns 200 OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply middleware
			middleware(handler).ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestAuthMiddlewareIntegration(t *testing.T) {
	// Test that middleware allows request to proceed with valid JWT
	jwtSecret := "integration-test-secret"
	middleware := AuthMiddleware(jwtSecret)

	// Create a handler that sets a custom header to verify it was called
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "success")
		w.WriteHeader(http.StatusOK)
	})

	// Create request with valid JWT token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT(jwtSecret))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Apply middleware
	middleware(handler).ServeHTTP(rr, req)

	// Check that request was allowed to proceed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("X-Test-Header") != "success" {
		t.Error("handler was not called, middleware blocked valid request")
	}
}

func TestAuthMiddlewareWithUserContext(t *testing.T) {
	jwtSecret := "context-test-secret"
	middleware := AuthMiddleware(jwtSecret)

	// Create a handler that checks for user context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		if user == nil {
			http.Error(w, "User context not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-User-Found", "true")
		w.WriteHeader(http.StatusOK)
	})

	// Create request with valid JWT token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+createValidJWT(jwtSecret))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Apply middleware
	middleware(handler).ServeHTTP(rr, req)

	// Check that user context was added
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("X-User-Found") != "true" {
		t.Error("user context was not added to request")
	}
}

// Helper functions to create JWT tokens for testing
func createValidJWT(secret string) string {
	claims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func createExpiredJWT(secret string) string {
	claims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}
