// Package context provides shared context utilities for the Ralts-CMS application,
// including user context extraction from requests.
package context

import (
	"context"
	"fmt"
)

// UserContext represents the user information extracted from JWT token
type UserContext struct {
	EntityID string `json:"entity_id"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key used to store user context in request context
	UserContextKey contextKey = "user"
)

// WithUser adds a UserContext to the given context
func WithUser(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (*UserContext, error) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

// GetUserIDFromContext extracts the user ID as a string pointer from the request context.
// Returns nil if the user is not authenticated (anonymous access).
func GetUserIDFromContext(ctx context.Context) *string {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	if !ok || user == nil {
		return nil
	}
	return &user.EntityID
}
