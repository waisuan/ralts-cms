package middlewares_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ralts-cms/internal/middlewares"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Do not run these in parallel: they replace slog.Default(), which is process-global.

func parseLastJSONLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	raw := bytes.TrimSpace(buf.Bytes())
	require.NotEmpty(t, raw, "expected log output")
	lines := bytes.Split(raw, []byte("\n"))
	var line map[string]any
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &line))
	return line
}

func TestLoggingMiddleware_StatusAndLevel(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	h := middlewares.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/gone", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	line := parseLastJSONLogLine(t, &buf)
	assert.Equal(t, float64(404), line["status_code"]) // JSON numbers decode as float64
	assert.Equal(t, "GET", line["method"])
	assert.Equal(t, "/gone", line["path"])
	assert.Equal(t, "HTTP request - client error", line["msg"])
	assert.Equal(t, "WARN", line["level"])
}

func TestLoggingMiddleware_SuccessIsInfo(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	h := middlewares.LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	line := parseLastJSONLogLine(t, &buf)
	assert.Equal(t, float64(200), line["status_code"])
	assert.Equal(t, "INFO", line["level"])
}
