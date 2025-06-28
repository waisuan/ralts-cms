// main.go - JWT Token Generator for Local Testing
//
// This script generates a valid JWT token for testing the Ralts-CMS API locally.
// The token uses the default JWT secret from the application configuration.
//
// Usage:
//
//	go run scripts/jwt/main.go
//
// The generated token will be valid for 24 hours and can be used with all API endpoints.
package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Use the default JWT secret from your config
	jwtSecret := "your-jwt-secret-key"

	// Create claims with the same structure as your test files
	claims := jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(24 * time.Hour).Unix(), // Valid for 24 hours
		"iat": time.Now().Unix(),
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("Error creating JWT token: %v\n", err)
		return
	}

	fmt.Println("=== Valid JWT Token for Local Testing ===")
	fmt.Println()
	fmt.Printf("JWT Secret: %s\n", jwtSecret)
	fmt.Println()
	fmt.Printf("Authorization Header: Bearer %s\n", tokenString)
	fmt.Println()
	fmt.Printf("Full Token: %s\n", tokenString)
	fmt.Println()
	fmt.Println("=== Usage Examples ===")
	fmt.Println()
	fmt.Println("curl -H \"Authorization: Bearer " + tokenString + "\" \\")
	fmt.Println("     http://localhost:8080/api/v1/machines")
	fmt.Println()
	fmt.Println("curl -H \"Authorization: Bearer " + tokenString + "\" \\")
	fmt.Println("     -H \"Content-Type: application/json\" \\")
	fmt.Println("     -d '{\"serial_number\":\"TEST123\",\"customer\":\"Test Customer\"}' \\")
	fmt.Println("     http://localhost:8080/api/v1/machines")
	fmt.Println()
	fmt.Println("=== Token Details ===")
	fmt.Printf("Subject (sub): test-user\n")
	fmt.Printf("Issued At (iat): %s\n", time.Unix(claims["iat"].(int64), 0).Format(time.RFC3339))
	fmt.Printf("Expires At (exp): %s\n", time.Unix(claims["exp"].(int64), 0).Format(time.RFC3339))
	fmt.Printf("Valid Until: %s\n", time.Unix(claims["exp"].(int64), 0).Format(time.RFC3339))
}
