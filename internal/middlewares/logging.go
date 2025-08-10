package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code and response body
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseBody *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Capture response body for error logging
	if rw.responseBody != nil {
		rw.responseBody.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs HTTP requests with response time, status codes, and error details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response buffer to capture error messages
		responseBuffer := &bytes.Buffer{}

		// Wrap the response writer to capture status code and response body
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default status code
			responseBody:   responseBuffer,
		}

		// Process the request
		next.ServeHTTP(rw, r)

		// Calculate response time
		duration := time.Since(start)

		// Determine log level and capture response body for errors
		var logLevel string
		var shouldLogBody bool
		switch {
		case rw.statusCode >= 500:
			logLevel = "ERROR"
			shouldLogBody = true
		case rw.statusCode >= 400:
			logLevel = "WARN"
			shouldLogBody = true
		case rw.statusCode >= 300:
			logLevel = "INFO"
		default:
			logLevel = "INFO"
		}

		// Base log message
		baseMsg := fmt.Sprintf("[%s] [%s] %s %s %s - %d - %v",
			time.Now().Format("2006-01-02 15:04:05"),
			logLevel,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)

		// For errors, capture and log the response body if it exists
		if shouldLogBody && rw.responseBody != nil && rw.responseBody.Len() > 0 {
			responseBody := rw.responseBody.String()
			// Truncate very long responses to avoid log spam
			if len(responseBody) > 500 {
				responseBody = responseBody[:500] + "..."
			}
			fmt.Printf("%s - Error: %s\n", baseMsg, responseBody)
		} else {
			fmt.Printf("%s\n", baseMsg)
		}
	})
}
