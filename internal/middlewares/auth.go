package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"ralts-cms/pkg/auth"
	"strings"
)

// UserContext represents the user information extracted from JWT token
type UserContext struct {
	EntityID string `json:"entity_id"`
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key used to store user context in request context
	UserContextKey contextKey = "user"
)

// AuthenticationMiddleware validates JWT tokens and adds user context to requests
func AuthenticationMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Extract token from Authorization header
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Invalid token: missing token", http.StatusUnauthorized)
				return
			}

			// Validate JWT token
			claims, err := auth.ValidateJWTToken(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Extract user information from claims
			entityID, ok := claims["entity_id"].(string)
			if !ok {
				http.Error(w, "Invalid token: missing entity_id", http.StatusUnauthorized)
				return
			}

			// Create user context
			user := &UserContext{
				EntityID: entityID,
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
