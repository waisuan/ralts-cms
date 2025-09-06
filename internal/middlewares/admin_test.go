package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ralts-cms/internal/middlewares"
	"ralts-cms/internal/users"

	"github.com/stretchr/testify/assert"
)

func TestAdminOnlyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		userContext    *middlewares.UserContext
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "admin user allowed",
			userContext: &middlewares.UserContext{
				EntityID: "1",
				UserID:   1,
				Role:     users.RoleAdmin,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "non-admin user forbidden",
			userContext: &middlewares.UserContext{
				EntityID: "2",
				UserID:   2,
				Role:     users.RoleNonAdmin,
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Admin access required\n",
		},
		{
			name: "user with empty role forbidden",
			userContext: &middlewares.UserContext{
				EntityID: "3",
				UserID:   3,
				Role:     "",
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Admin access required\n",
		},
		{
			name: "user with invalid role forbidden",
			userContext: &middlewares.UserContext{
				EntityID: "4",
				UserID:   4,
				Role:     "INVALID_ROLE",
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Admin access required\n",
		},
		{
			name:           "no user context unauthorized",
			userContext:    nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that returns success
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Create middleware
			middleware := middlewares.AdminOnlyMiddleware()
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Add user context if provided
			if tt.userContext != nil {
				ctx := context.WithValue(req.Context(), middlewares.UserContextKey, tt.userContext)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
