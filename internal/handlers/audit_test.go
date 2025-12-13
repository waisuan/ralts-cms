package handlers_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"ralts-cms/internal/audit"
	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

// AuditHandlerTestSuite defines the test suite for audit handler
type AuditHandlerTestSuite struct {
	suite.Suite

	handler          *handlers.AuditHandler
	mockRepo         *audit.MockRepository
	mockAuditService *audit.MockAuditService
	ctrl             *gomock.Controller
}

// SetupTest sets up each test
func (suite *AuditHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = audit.NewMockRepository(suite.ctrl)
	suite.mockAuditService = audit.NewMockAuditService(suite.ctrl)

	// Allow any audit events to be logged
	suite.mockAuditService.EXPECT().LogEvent(gomock.Any()).AnyTimes()

	deps := &deps.Dependencies{
		Config: &deps.Config{
			JWTSecret: "your-jwt-secret-key",
		},
		Logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
		AuditRepository: suite.mockRepo,
		AuditService:    suite.mockAuditService,
	}
	suite.handler = handlers.NewAuditHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *AuditHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// addUserContext adds a mock user context to the request
func (suite *AuditHandlerTestSuite) addUserContext(req *http.Request) *http.Request {
	userCtx := &appctx.UserContext{
		UserID: 1,
		Role:   "ADMIN",
	}
	ctx := context.WithValue(req.Context(), appctx.UserContextKey, userCtx)
	return req.WithContext(ctx)
}

func (suite *AuditHandlerTestSuite) TestListAuditEvents() {
	suite.Run("should list audit events successfully", func() {
		now := time.Now()
		events := []*audit.Event{
			{
				ID:           "event-1",
				Action:       audit.ActionCreated,
				ResourceType: audit.ResourceMachine,
				ResourceID:   "SN-001",
				CreatedAt:    now,
			},
			{
				ID:           "event-2",
				Action:       audit.ActionUpdated,
				ResourceType: audit.ResourceUser,
				ResourceID:   "1",
				CreatedAt:    now.Add(-time.Hour),
			},
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(events, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(2), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response handlers.ListAuditEventsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Len(response.Events, 2)
		suite.Assert().Equal(int32(2), response.TotalCount)
		suite.Assert().Equal(int32(50), response.Limit)
		suite.Assert().Equal(int32(0), response.Offset)
	})

	suite.Run("should apply pagination parameters", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, opts *audit.ListOptions) ([]*audit.Event, error) {
				suite.Assert().Equal(int32(10), opts.Limit)
				suite.Assert().Equal(int32(20), opts.Offset)
				return []*audit.Event{}, nil
			},
		)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events?limit=10&offset=20", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should apply resource_type filter", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, opts *audit.ListOptions) ([]*audit.Event, error) {
				suite.Assert().NotNil(opts.ResourceType)
				suite.Assert().Equal("machine", *opts.ResourceType)
				return []*audit.Event{}, nil
			},
		)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events?resource_type=machine", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should apply action filter", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, opts *audit.ListOptions) ([]*audit.Event, error) {
				suite.Assert().NotNil(opts.Action)
				suite.Assert().Equal("created", *opts.Action)
				return []*audit.Event{}, nil
			},
		)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events?action=created", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should apply date range filters", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, opts *audit.ListOptions) ([]*audit.Event, error) {
				suite.Assert().NotNil(opts.FromDate)
				suite.Assert().NotNil(opts.ToDate)
				return []*audit.Event{}, nil
			},
		)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events?from_date=2024-01-01T00:00:00Z&to_date=2024-12-31T23:59:59Z", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should enforce max limit of 100", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, opts *audit.ListOptions) ([]*audit.Event, error) {
				suite.Assert().Equal(int32(100), opts.Limit)
				return []*audit.Event{}, nil
			},
		)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events?limit=500", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should return empty list when no events", func() {
		suite.mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*audit.Event{}, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(int32(0), nil)

		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.ListAuditEvents(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListAuditEventsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Len(response.Events, 0)
		suite.Assert().Equal(int32(0), response.TotalCount)
	})
}

func (suite *AuditHandlerTestSuite) TestStreamAuditEvents() {
	suite.Run("should return 501 Not Implemented", func() {
		req := httptest.NewRequest("GET", "/api/v1/admin/audit/events/stream", nil)
		req = suite.addUserContext(req)
		w := httptest.NewRecorder()

		suite.handler.StreamAuditEvents(w, req)

		suite.Assert().Equal(http.StatusNotImplemented, w.Code)
	})
}

func TestAuditHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuditHandlerTestSuite))
}
