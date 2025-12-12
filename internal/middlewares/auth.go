// Package middlewares provides HTTP middleware functionality for the Ralts-CMS application,
// including authentication, CORS, and logging middleware.
package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"ralts-cms/pkg/auth"
	"strconv"
	"strings"

	appctx "ralts-cms/internal/context"
)

// UserContext is an alias to the shared context.UserContext for backward compatibility
type UserContext = appctx.UserContext

// UserContextKey is the key used to store user context in request context
// Re-exported from context package for backward compatibility
var UserContextKey = appctx.UserContextKey

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

			// Extract role from claims
			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, "Invalid token: missing role", http.StatusUnauthorized)
				return
			}

			// Convert entity_id to int64
			userID, err := strconv.ParseInt(entityID, 10, 64)
			if err != nil {
				http.Error(w, "Invalid token: invalid entity_id format", http.StatusUnauthorized)
				return
			}

			// Create user context with information from JWT token
			userCtx := &UserContext{
				EntityID: entityID,
				UserID:   userID,
				Role:     role,
			}

			// Add user context to request using the shared context package
			ctx := appctx.WithUser(r.Context(), userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user information from request context
// This is a convenience wrapper around context.GetUserFromContext
func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	return appctx.GetUserFromContext(ctx)
}

// ExtractUserIDFromJWT extracts the user ID from a JWT token string
func ExtractUserIDFromJWT(tokenString, jwtSecret string) (int64, error) {
	// Validate JWT token
	claims, err := auth.ValidateJWTToken(tokenString, jwtSecret)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Extract entity_id from claims
	entityID, ok := claims["entity_id"].(string)
	if !ok {
		return 0, fmt.Errorf("invalid token: missing entity_id")
	}

	// Convert to int64
	userID, err := strconv.ParseInt(entityID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid token: invalid entity_id format: %w", err)
	}

	return userID, nil
}
