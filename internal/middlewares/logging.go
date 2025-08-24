package middlewares

import (
	"bytes"
	"context"
	"log/slog"
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

// LoggingMiddleware logs HTTP requests with response time, status codes, and error details using structured logging
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

		// Base attributes for all log entries
		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status_code", rw.statusCode),
			slog.Duration("duration", duration),
		}

		// Add query parameters if present
		if r.URL.RawQuery != "" {
			attrs = append(attrs, slog.String("query", r.URL.RawQuery))
		}

		// Add user agent
		if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		// Determine log level and message based on status code
		ctx := context.Background()
		switch {
		case rw.statusCode >= 500:
			// Include error response body for server errors
			if rw.responseBody != nil && rw.responseBody.Len() > 0 {
				responseBody := rw.responseBody.String()
				// Truncate very long responses to avoid log spam
				if len(responseBody) > 500 {
					responseBody = responseBody[:500] + "..."
				}
				attrs = append(attrs, slog.String("error_response", responseBody))
			}
			slog.LogAttrs(ctx, slog.LevelError, "HTTP request - server error", attrs...)

		case rw.statusCode >= 400:
			// Include error response body for client errors
			if rw.responseBody != nil && rw.responseBody.Len() > 0 {
				responseBody := rw.responseBody.String()
				if len(responseBody) > 500 {
					responseBody = responseBody[:500] + "..."
				}
				attrs = append(attrs, slog.String("error_response", responseBody))
			}
			slog.LogAttrs(ctx, slog.LevelWarn, "HTTP request - client error", attrs...)

		default:
			// Successful requests at INFO level
			slog.LogAttrs(ctx, slog.LevelInfo, "HTTP request", attrs...)
		}
	})
}
