package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/flags"
	"ralts-cms/internal/handlers"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// FlagsHandlerTestSuite covers the flag HTTP endpoints and their role checks.
type FlagsHandlerTestSuite struct {
	suite.Suite

	handler     *handlers.FlagsHandler
	mockRepo    *flags.MockRepository
	mockService *flags.MockService
	ctrl        *gomock.Controller
}

func (suite *FlagsHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = flags.NewMockRepository(suite.ctrl)
	suite.mockService = flags.NewMockService(suite.ctrl)

	d := &deps.Dependencies{
		Config:          &deps.Config{JWTSecret: "test-secret"},
		Logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
		FlagsRepository: suite.mockRepo,
		FlagsService:    suite.mockService,
	}
	suite.handler = handlers.NewFlagsHandler(d)
}

func (suite *FlagsHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *FlagsHandlerTestSuite) withUser(r *http.Request, userID int64, role string) *http.Request {
	ctx := context.WithValue(r.Context(), appctx.UserContextKey, &appctx.UserContext{
		UserID: userID,
		Role:   role,
	})
	return r.WithContext(ctx)
}

func (suite *FlagsHandlerTestSuite) TestListForMachine() {
	suite.Run("returns flags for the machine", func() {
		suite.mockRepo.EXPECT().
			ListForMachine(gomock.Any(), "SN-1", false).
			Return([]*flags.Flag{{ID: "f-1", MachineSerialNumber: "SN-1", Reason: flags.ReasonOther}}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/SN-1/flags", nil)
		req = suite.withUser(req, 1, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1"})
		w := httptest.NewRecorder()

		suite.handler.ListForMachine(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var body map[string][]*flags.Flag
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &body))
		suite.Require().Len(body["flags"], 1)
		suite.Assert().Equal("f-1", body["flags"][0].ID)
	})

	suite.Run("passes include_resolved=true through to repo", func() {
		suite.mockRepo.EXPECT().
			ListForMachine(gomock.Any(), "SN-2", true).
			Return([]*flags.Flag{}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/SN-2/flags?include_resolved=true", nil)
		req = suite.withUser(req, 1, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-2"})
		w := httptest.NewRecorder()

		suite.handler.ListForMachine(w, req)
		suite.Assert().Equal(http.StatusOK, w.Code)
	})
}

func (suite *FlagsHandlerTestSuite) TestListOpenByMachine() {
	suite.Run("splits comma-separated serials and returns a map", func() {
		suite.mockRepo.EXPECT().
			ListOpenBySerials(gomock.Any(), []string{"SN-1", "SN-2"}).
			Return(map[string][]*flags.Flag{
				"SN-1": {{ID: "f-1", MachineSerialNumber: "SN-1"}},
			}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/flags/open-by-machine?serials=SN-1,SN-2", nil)
		req = suite.withUser(req, 1, "USER")
		w := httptest.NewRecorder()

		suite.handler.ListOpenByMachine(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var resp handlers.BatchOpenByMachineResponse
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
		suite.Assert().Len(resp.Flags["SN-1"], 1)
		suite.Assert().Empty(resp.Flags["SN-2"])
	})

	suite.Run("empty serials query returns an empty map without calling the repo", func() {
		req := httptest.NewRequest("GET", "/api/v1/machines/flags/open-by-machine", nil)
		req = suite.withUser(req, 1, "USER")
		w := httptest.NewRecorder()
		suite.handler.ListOpenByMachine(w, req)
		suite.Assert().Equal(http.StatusOK, w.Code)
	})
}

func (suite *FlagsHandlerTestSuite) TestList() {
	suite.Run("defaults to open flags with a page size of 50", func() {
		opts := flags.ListOptions{Status: flags.StatusOpen, Limit: 50, Offset: 0}
		suite.mockRepo.EXPECT().
			List(gomock.Any(), opts).
			Return([]*flags.Flag{{ID: "f-1", MachineSerialNumber: "SN-1", Status: flags.StatusOpen}}, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), opts).Return(1, nil)

		req := httptest.NewRequest("GET", "/api/v1/flags", nil)
		req = suite.withUser(req, 9, "ADMIN")
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var resp handlers.ListFlagsResponse
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
		suite.Require().Len(resp.Flags, 1)
		suite.Assert().Equal("f-1", resp.Flags[0].ID)
		suite.Assert().Equal(1, resp.Count)
		suite.Assert().Equal(int32(50), resp.Limit)
		suite.Assert().Equal(handlers.FlagScopeAll, resp.Scope)
	})

	suite.Run("scopes non-admins to their machines and their own resolutions", func() {
		callerID := int64(7)
		opts := flags.ListOptions{
			Status:               flags.StatusOpen,
			Limit:                50,
			AssignedUserID:       &callerID,
			OwnResolutionsUserID: &callerID,
		}
		suite.mockRepo.EXPECT().List(gomock.Any(), opts).Return([]*flags.Flag{}, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), opts).Return(0, nil)

		req := httptest.NewRequest("GET", "/api/v1/flags", nil)
		req = suite.withUser(req, callerID, "USER")
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var resp handlers.ListFlagsResponse
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
		suite.Assert().Equal(handlers.FlagScopeAssigned, resp.Scope)
	})

	suite.Run("rejects unauthenticated callers", func() {
		req := httptest.NewRequest("GET", "/api/v1/flags", nil)
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
	})

	suite.Run("status=all drops the status filter", func() {
		opts := flags.ListOptions{Status: "", Limit: 50}
		suite.mockRepo.EXPECT().List(gomock.Any(), opts).Return(nil, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), opts).Return(0, nil)

		req := httptest.NewRequest("GET", "/api/v1/flags?status=all", nil)
		req = suite.withUser(req, 9, "ADMIN")
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)
		// A nil slice from the repo must still serialise as [] for the frontend.
		suite.Assert().Contains(w.Body.String(), `"flags":[]`)
	})

	suite.Run("status=resolved filters on resolved flags", func() {
		opts := flags.ListOptions{Status: flags.StatusResolved, Limit: 10, Offset: 20}
		suite.mockRepo.EXPECT().List(gomock.Any(), opts).Return([]*flags.Flag{}, nil)
		suite.mockRepo.EXPECT().Count(gomock.Any(), opts).Return(25, nil)

		req := httptest.NewRequest("GET", "/api/v1/flags?status=resolved&limit=10&offset=20", nil)
		req = suite.withUser(req, 9, "ADMIN")
		w := httptest.NewRecorder()

		suite.handler.List(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var resp handlers.ListFlagsResponse
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
		suite.Assert().Equal(25, resp.Count)
		suite.Assert().Equal(int32(20), resp.Offset)
	})

	suite.Run("rejects invalid status, limit and offset without hitting the repo", func() {
		for _, query := range []string{"?status=bogus", "?limit=0", "?limit=201", "?limit=abc", "?offset=-1"} {
			req := httptest.NewRequest("GET", "/api/v1/flags"+query, nil)
			req = suite.withUser(req, 9, "ADMIN")
			w := httptest.NewRecorder()

			suite.handler.List(w, req)
			suite.Assert().Equal(http.StatusBadRequest, w.Code, "query %s", query)
		}
	})
}

func (suite *FlagsHandlerTestSuite) TestCreateAdminOnly() {
	body, _ := json.Marshal(handlers.CreateFlagRequest{
		Reason: flags.ReasonMissingValues,
	})

	suite.Run("rejects non-admin callers", func() {
		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags", bytes.NewReader(body))
		req = suite.withUser(req, 1, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1"})
		w := httptest.NewRecorder()

		suite.handler.Create(w, req)
		suite.Assert().Equal(http.StatusForbidden, w.Code)
	})

	suite.Run("returns 401 without user context", func() {
		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1"})
		w := httptest.NewRecorder()

		suite.handler.Create(w, req)
		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
	})

	suite.Run("replacing an existing open flag answers 200", func() {
		suite.mockService.EXPECT().
			CreateManualFlag(gomock.Any(), int64(9), "SN-1", flags.ReasonMissingValues, "").
			Return(&flags.Flag{ID: "existing", MachineSerialNumber: "SN-1"}, false, nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags", bytes.NewReader(body))
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1"})
		w := httptest.NewRecorder()

		suite.handler.Create(w, req)
		suite.Require().Equal(http.StatusOK, w.Code)

		var f flags.Flag
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &f))
		suite.Assert().Equal("existing", f.ID)
	})

	suite.Run("admin can create a flag; service is invoked with actor + serial", func() {
		suite.mockService.EXPECT().
			CreateManualFlag(gomock.Any(), int64(9), "SN-1", flags.ReasonMissingValues, "").
			Return(&flags.Flag{ID: "created", MachineSerialNumber: "SN-1"}, true, nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags", bytes.NewReader(body))
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1"})
		w := httptest.NewRecorder()

		suite.handler.Create(w, req)
		suite.Require().Equal(http.StatusCreated, w.Code)

		var f flags.Flag
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &f))
		suite.Assert().Equal("created", f.ID)
	})
}

func (suite *FlagsHandlerTestSuite) TestResolve() {
	suite.Run("returns 401 without user context", func() {
		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", nil)
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusUnauthorized, w.Code)
	})

	suite.Run("rejects a non-admin who is not the assignee", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(1), false, "f-1").
			Return(false, nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", nil)
		req = suite.withUser(req, 1, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusForbidden, w.Code)
	})

	suite.Run("allows a non-admin assignee to resolve from their inbox", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(7), false, "f-1").
			Return(true, nil)
		suite.mockService.EXPECT().
			Resolve(gomock.Any(), int64(7), false, "f-1", "").
			Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", nil)
		req = suite.withUser(req, 7, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("passes an optional resolution note from the body", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(9), true, "f-1").
			Return(true, nil)
		suite.mockService.EXPECT().
			Resolve(gomock.Any(), int64(9), true, "f-1", "swapped the sensor").
			Return(nil)

		body := strings.NewReader(`{"note":"swapped the sensor"}`)
		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", body)
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("rejects a malformed body", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(9), true, "f-1").
			Return(true, nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", strings.NewReader("{oops"))
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
	})

	suite.Run("surfaces an over-long note as 400", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(9), true, "f-1").
			Return(true, nil)
		suite.mockService.EXPECT().
			Resolve(gomock.Any(), int64(9), true, "f-1", gomock.Any()).
			Return(flags.ErrNoteTooLong)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", strings.NewReader(`{"note":"long"}`))
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
	})

	suite.Run("admin resolve returns 204 on success", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(9), true, "f-1").
			Return(true, nil)
		suite.mockService.EXPECT().
			Resolve(gomock.Any(), int64(9), true, "f-1", "").
			Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/f-1/resolve", nil)
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "f-1"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("unknown flag id returns 404", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(9), true, "missing").
			Return(true, nil)
		suite.mockService.EXPECT().
			Resolve(gomock.Any(), int64(9), true, "missing", "").
			Return(flags.ErrNotFound)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/missing/resolve", nil)
		req = suite.withUser(req, 9, "ADMIN")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "missing"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusNotFound, w.Code)
	})

	suite.Run("permission check reporting an unknown flag returns 404", func() {
		suite.mockService.EXPECT().
			CanResolve(gomock.Any(), int64(1), false, "gone").
			Return(false, flags.ErrNotFound)

		req := httptest.NewRequest("POST", "/api/v1/machines/SN-1/flags/gone/resolve", nil)
		req = suite.withUser(req, 1, "USER")
		req = mux.SetURLVars(req, map[string]string{"serial_number": "SN-1", "id": "gone"})
		w := httptest.NewRecorder()

		suite.handler.Resolve(w, req)
		suite.Assert().Equal(http.StatusNotFound, w.Code)
	})
}

func TestFlagsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FlagsHandlerTestSuite))
}
