// Package middlewares provides HTTP middleware functionality for the Ralts-CMS application,
// including admin authorization middleware.
package middlewares

import (
	"net/http"
	"ralts-cms/internal/users"
)

// AdminOnlyMiddleware ensures that only users with ADMIN role can access the protected endpoints.
// This middleware must be chained after AuthenticationMiddleware to work properly.
func AdminOnlyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user context from the request (set by AuthenticationMiddleware)
			userCtx, err := GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has admin role
			if userCtx.Role != users.RoleAdmin {
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			// User is authenticated and has admin role, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}
