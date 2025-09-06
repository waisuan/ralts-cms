package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ralts-cms/internal/middlewares"
	"ralts-cms/internal/users"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationMiddleware(t *testing.T) {
	jwtSecret := "test-secret"

	// Helper function to create a JWT token
	createToken := func(entityID, role string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"entity_id": entityID,
			"role":      role,
			"exp":       time.Now().Add(24 * time.Hour).Unix(),
			"iat":       time.Now().Unix(),
		})
		tokenString, _ := token.SignedString([]byte(jwtSecret))
		return tokenString
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		checkContext   bool
	}{
		{
			name:           "valid token with admin user",
			authHeader:     "Bearer " + createToken("1", users.RoleAdmin),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext:   true,
		},
		{
			name:           "valid token with non-admin user",
			authHeader:     "Bearer " + createToken("2", users.RoleNonAdmin),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext:   true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header is required\n",
		},
		{
			name:           "invalid token format",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token: failed to parse token: token is malformed: token contains an invalid number of segments\n",
		},
		{
			name:           "missing bearer prefix",
			authHeader:     createToken("1", users.RoleAdmin),
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			checkContext:   true,
		},
		{
			name: "missing role in token",
			authHeader: "Bearer " + func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"entity_id": "1",
					"exp":       time.Now().Add(24 * time.Hour).Unix(),
					"iat":       time.Now().Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return tokenString
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token: missing role\n",
		},
		{
			name: "missing entity_id in token",
			authHeader: "Bearer " + func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"role": users.RoleAdmin,
					"exp":  time.Now().Add(24 * time.Hour).Unix(),
					"iat":  time.Now().Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return tokenString
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token: missing entity_id\n",
		},
		{
			name: "invalid entity_id format",
			authHeader: "Bearer " + func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"entity_id": "invalid",
					"role":      users.RoleAdmin,
					"exp":       time.Now().Add(24 * time.Hour).Unix(),
					"iat":       time.Now().Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return tokenString
			}(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token: invalid entity_id format\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that returns success and checks context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					userCtx, err := middlewares.GetUserFromContext(r.Context())
					assert.NoError(t, err)
					assert.NotNil(t, userCtx)
					assert.NotEmpty(t, userCtx.EntityID)
					assert.NotZero(t, userCtx.UserID)
					assert.NotEmpty(t, userCtx.Role)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Create middleware
			middleware := middlewares.AuthenticationMiddleware(jwtSecret)
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
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

func TestExtractUserIDFromJWT(t *testing.T) {
	jwtSecret := "test-secret"

	// Helper function to create a JWT token
	createToken := func(entityID, role string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"entity_id": entityID,
			"role":      role,
			"exp":       time.Now().Add(24 * time.Hour).Unix(),
			"iat":       time.Now().Unix(),
		})
		tokenString, _ := token.SignedString([]byte(jwtSecret))
		return tokenString
	}

	tests := []struct {
		name        string
		token       string
		expectedID  int64
		expectError bool
	}{
		{
			name:        "valid token with numeric entity_id",
			token:       createToken("123", users.RoleAdmin),
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectedID:  0,
			expectError: true,
		},
		{
			name: "token with non-numeric entity_id",
			token: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"entity_id": "abc",
					"role":      users.RoleAdmin,
					"exp":       time.Now().Add(24 * time.Hour).Unix(),
					"iat":       time.Now().Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return tokenString
			}(),
			expectedID:  0,
			expectError: true,
		},
		{
			name: "token missing entity_id",
			token: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"role": users.RoleAdmin,
					"exp":  time.Now().Add(24 * time.Hour).Unix(),
					"iat":  time.Now().Unix(),
				})
				tokenString, _ := token.SignedString([]byte(jwtSecret))
				return tokenString
			}(),
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := middlewares.ExtractUserIDFromJWT(tt.token, jwtSecret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Zero(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
