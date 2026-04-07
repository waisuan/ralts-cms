// Package middlewares provides shared HTTP middleware. This file implements structured request logging without buffering response bodies.
package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code for logging.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP requests with response time and status codes using structured logging.
// It does not buffer response bodies, avoiding extra memory use on large payloads (e.g. CSV export).
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status_code", rw.statusCode),
			slog.Duration("duration", duration),
		}

		if r.URL.RawQuery != "" {
			attrs = append(attrs, slog.String("query", r.URL.RawQuery))
		}

		if userAgent := r.Header.Get("User-Agent"); userAgent != "" {
			attrs = append(attrs, slog.String("user_agent", userAgent))
		}

		ctx := context.Background()
		switch {
		case rw.statusCode >= 500:
			slog.LogAttrs(ctx, slog.LevelError, "HTTP request - server error", attrs...)
		case rw.statusCode >= 400:
			slog.LogAttrs(ctx, slog.LevelWarn, "HTTP request - client error", attrs...)
		default:
			slog.LogAttrs(ctx, slog.LevelInfo, "HTTP request", attrs...)
		}
	})
}
