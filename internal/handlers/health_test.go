package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"testing"

	"github.com/stretchr/testify/suite"
)

// HealthHandlerTestSuite defines the test suite for health handler
type HealthHandlerTestSuite struct {
	suite.Suite

	handler *handlers.HealthHandler
	deps    *deps.Dependencies
}

// SetupTest sets up each test
func (suite *HealthHandlerTestSuite) SetupTest() {
	suite.deps = &deps.Dependencies{}
	suite.handler = handlers.NewHealthHandler(suite.deps)
}

func (suite *HealthHandlerTestSuite) TestHealth() {
	suite.Run("should return healthy status", func() {
		// Create request
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		// Call the handler
		suite.handler.Health(rr, req)

		// Check status code
		suite.Assert().Equal(http.StatusOK, rr.Code)

		// Check content type
		suite.Assert().Equal("application/json", rr.Header().Get("Content-Type"))

		// Parse response body
		var response handlers.HealthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		suite.Require().NoError(err)

		// Check response fields
		suite.Assert().Equal("healthy", response.Status)
		suite.Assert().Equal("ralts-cms", response.Service)
		suite.Assert().NotEmpty(response.Timestamp)
	})
}

// TestHealthHandlerTestSuite runs the test suite
func TestHealthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}
