package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ralts-cms/internal/middlewares"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware_GETPassesThrough(t *testing.T) {
	t.Parallel()

	nextCalled := false
	h := middlewares.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.True(t, nextCalled)
	assert.Equal(t, http.StatusTeapot, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, rr.Header().Get("Access-Control-Expose-Headers"), "Content-Disposition")
}

func TestCORSMiddleware_OPTIONSShortCircuit(t *testing.T) {
	t.Parallel()

	nextCalled := false
	h := middlewares.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/machines", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.False(t, nextCalled, "OPTIONS must not invoke next handler")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}
