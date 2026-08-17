package handlers_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/notifications"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// NotificationsHandlerTestSuite exercises the notifications HTTP endpoints
// against a mocked repository. All endpoints must scope reads/writes to the
// authenticated user carried in the request context.
type NotificationsHandlerTestSuite struct {
	suite.Suite

	handler  *handlers.NotificationsHandler
	mockRepo *notifications.MockRepository
	ctrl     *gomock.Controller
}

func (suite *NotificationsHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = notifications.NewMockRepository(suite.ctrl)

	d := &deps.Dependencies{
		Config:                 &deps.Config{JWTSecret: "test-secret"},
		Logger:                 slog.New(slog.NewTextHandler(os.Stdout, nil)),
		NotificationRepository: suite.mockRepo,
	}
	suite.handler = handlers.NewNotificationsHandler(d)
}

func (suite *NotificationsHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// withUser attaches a JWT-like user context to the request.
func (suite *NotificationsHandlerTestSuite) withUser(r *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(r.Context(), appctx.UserContextKey, &appctx.UserContext{
		UserID: userID,
		Role:   "USER",
	})
	return r.WithContext(ctx)
}

func (suite *NotificationsHandlerTestSuite) TestListUnauthorized() {
	suite.Run("returns 401 without user context", func() {
		req := httptest.NewRequest("GET", "/api/v1/notifications", nil)
		w := httptest.NewRecorder()
		suite.handler.List(w, req)
		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
	})
}

func (suite *NotificationsHandlerTestSuite) TestList() {
	suite.Run("scopes results to the authenticated user and echoes pagination", func() {
		now := time.Now().UTC()
		items := []*notifications.Notification{
			{ID: "a", UserID: 42, Type: notifications.TypeAssigned, Title: "t1", CreatedAt: now},
			{ID: "b", UserID: 42, Type: notifications.TypeFlagged, Title: "t2", CreatedAt: now},
		}

		suite.mockRepo.EXPECT().
			ListForUser(gomock.Any(), int64(42), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int64, opts notifications.ListOptions) ([]*notifications.Notification, error) {
				suite.Assert().Equal(int32(10), opts.Limit)
				suite.Assert().Equal(int32(20), opts.Offset)
				suite.Assert().True(opts.UnreadOnly)
				return items, nil
			})
		suite.mockRepo.EXPECT().CountForUser(gomock.Any(), int64(42), gomock.Any()).Return(2, nil)
		suite.mockRepo.EXPECT().CountUnread(gomock.Any(), int64(42)).Return(1, nil)

		req := httptest.NewRequest("GET", "/api/v1/notifications?limit=10&offset=20&unread_only=true", nil)
		req = suite.withUser(req, 42)
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var resp handlers.ListNotificationsResponse
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
		suite.Assert().Len(resp.Notifications, 2)
		suite.Assert().Equal(int32(10), resp.Limit)
		suite.Assert().Equal(int32(20), resp.Offset)
		suite.Assert().Equal(2, resp.Count)
		suite.Assert().Equal(1, resp.UnreadCount)
	})

	suite.Run("rejects invalid limit", func() {
		req := httptest.NewRequest("GET", "/api/v1/notifications?limit=9999", nil)
		req = suite.withUser(req, 1)
		w := httptest.NewRecorder()
		suite.handler.List(w, req)
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
	})
}

func (suite *NotificationsHandlerTestSuite) TestUnreadCount() {
	suite.Run("returns unread_count scoped to the caller", func() {
		suite.mockRepo.EXPECT().CountUnread(gomock.Any(), int64(7)).Return(3, nil)

		req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
		req = suite.withUser(req, 7)
		w := httptest.NewRecorder()
		suite.handler.UnreadCount(w, req)

		suite.Require().Equal(http.StatusOK, w.Code)
		var out map[string]int
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &out))
		suite.Assert().Equal(3, out["unread_count"])
	})
}

func (suite *NotificationsHandlerTestSuite) TestMarkReadRoutesIDToRepo() {
	suite.Run("passes id and user id from context to repo", func() {
		suite.mockRepo.EXPECT().MarkRead(gomock.Any(), int64(5), "note-1").Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/notifications/note-1/read", nil)
		req = suite.withUser(req, 5)
		req = mux.SetURLVars(req, map[string]string{"id": "note-1"})
		w := httptest.NewRecorder()

		suite.handler.MarkRead(w, req)
		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("rejects missing id", func() {
		req := httptest.NewRequest("POST", "/api/v1/notifications//read", nil)
		req = suite.withUser(req, 5)
		w := httptest.NewRecorder()
		suite.handler.MarkRead(w, req)
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
	})
}

func (suite *NotificationsHandlerTestSuite) TestMarkAllRead() {
	suite.Run("scopes bulk update to caller and returns count", func() {
		suite.mockRepo.EXPECT().MarkAllRead(gomock.Any(), int64(11)).Return(4, nil)

		req := httptest.NewRequest("POST", "/api/v1/notifications/read-all", nil)
		req = suite.withUser(req, 11)
		w := httptest.NewRecorder()
		suite.handler.MarkAllRead(w, req)

		suite.Require().Equal(http.StatusOK, w.Code)
		var out map[string]int
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &out))
		suite.Assert().Equal(4, out["updated"])
	})
}

func TestNotificationsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationsHandlerTestSuite))
}
