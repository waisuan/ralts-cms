package handler

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		credentials    string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid credentials",
			credentials:    "admin:password",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password")),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			credentials:    "admin:password",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			credentials:    "admin:password",
			authHeader:     "Bearer token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid credentials",
			credentials:    "admin:password",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrong")),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed base64",
			credentials:    "admin:password",
			authHeader:     "Basic invalid-base64",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := BasicAuthMiddleware(tt.credentials)

			// Create a simple handler that returns 200 OK
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply middleware
			middleware(handler).ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

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

func TestMiddlewareChain(t *testing.T) {
	// Test that multiple middleware can be chained together
	credentials := "admin:password"
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password"))

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", authHeader)
	rr := httptest.NewRecorder()

	// Apply multiple middleware
	chainedHandler := LoggingMiddleware(handler)
	chainedHandler = CORSMiddleware(chainedHandler)
	chainedHandler = BasicAuthMiddleware(credentials)(chainedHandler)

	// Serve the request
	chainedHandler.ServeHTTP(rr, req)

	// Check that all middleware worked
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestBasicAuthMiddleware_EmptyCredentials(t *testing.T) {
	// Test with empty credentials
	middleware := BasicAuthMiddleware("")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":")))
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	// Should return 401 Unauthorized for empty credentials
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
