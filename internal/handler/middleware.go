package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuthMiddleware creates middleware for Basic Authentication
func BasicAuthMiddleware(credentials string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Basic ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Decode the base64 credentials
			encodedCredentials := strings.TrimPrefix(auth, "Basic ")
			decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
			if err != nil {
				http.Error(w, "Invalid credentials format", http.StatusUnauthorized)
				return
			}

			decodedCredentials := string(decodedBytes)
			if decodedCredentials != credentials {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
