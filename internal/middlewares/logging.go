package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs HTTP requests with response time and status codes
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default status code
		}

		// Process the request
		next.ServeHTTP(rw, r)

		// Calculate response time
		duration := time.Since(start)

		// Determine log level based on status code
		var logLevel string
		switch {
		case rw.statusCode >= 500:
			logLevel = "ERROR"
		case rw.statusCode >= 400:
			logLevel = "WARN"
		case rw.statusCode >= 300:
			logLevel = "INFO"
		default:
			logLevel = "INFO"
		}

		// Log the request with response time and status code
		fmt.Printf("[%s] [%s] %s %s %s - %d - %v\n",
			time.Now().Format("2006-01-02 15:04:05"),
			logLevel,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
	})
}
