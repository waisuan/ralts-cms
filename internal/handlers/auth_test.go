package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/refreshtokens"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	handler       *handlers.AuthHandler
	mockRefresh   *refreshtokens.MockRepository
	mockUsers     *users.MockRepository
	ctrl          *gomock.Controller
	refreshPepper string
}

func (s *AuthHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRefresh = refreshtokens.NewMockRepository(s.ctrl)
	s.mockUsers = users.NewMockRepository(s.ctrl)
	s.refreshPepper = "test-jwt" + ":ralts-refresh"
	d := &deps.Dependencies{
		Config: &deps.Config{
			JWTSecret:            "test-jwt",
			AccessTokenLifetime:  20 * time.Minute,
			RefreshTokenLifetime: 24 * time.Hour,
			RefreshTokenPepper:   "",
		},
		Logger:                 slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		RefreshTokenRepository: s.mockRefresh,
		UsersRepository:        s.mockUsers,
	}
	// use derived pepper like production when REFRESH_TOKEN_PEPPER is empty
	s.Equal(s.refreshPepper, d.Config.RefreshPepper())
	s.handler = handlers.NewAuthHandler(d)
}

func (s *AuthHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *AuthHandlerTestSuite) TestRefresh_Success() {
	ctx := gomock.Any()
	raw := "client-refresh-raw-token"
	uid := uuid.New()
	tok := &refreshtokens.Token{ID: uid, UserID: 7}
	st := users.StatusApproved
	u := &users.User{
		ID:       7,
		Username: "u",
		Email:    "u@e.com",
		Role:     users.RoleAdmin,
		Approved: true,
		Status:   &st,
	}
	hash := auth.HashRefreshToken(raw, s.refreshPepper)

	s.mockRefresh.EXPECT().GetActiveByHash(ctx, hash).Return(tok, nil)
	s.mockUsers.EXPECT().GetByID(ctx, int64(7)).Return(u, nil)
	s.mockRefresh.EXPECT().Rotate(ctx, uid, int64(7), gomock.Any(), gomock.Any()).
		Return(uuid.New(), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(`{"refresh_token":"`+raw+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)

	s.Equal(http.StatusOK, rr.Code)
	var out handlers.TokenPairResponse
	s.NoError(json.Unmarshal(rr.Body.Bytes(), &out))
	s.NotEmpty(out.Token)
	s.NotEmpty(out.RefreshToken)
	_, err := auth.ValidateJWTToken(out.Token, "test-jwt")
	s.NoError(err)
}

func (s *AuthHandlerTestSuite) TestRefresh_Validation() {
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, httptest.NewRequest(http.MethodGet, "/refresh", nil))
	s.Equal(http.StatusMethodNotAllowed, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_EmptyBody() {
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString("not json")))
	s.Equal(http.StatusBadRequest, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_RequiresRefreshToken() {
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"refresh_token":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusBadRequest, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_LookupDBError_500() {
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db down"))
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"t"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusInternalServerError, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_UnknownOrExpired() {
	s.mockRefresh.EXPECT().
		GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(nil, refreshtokens.ErrNotFound)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"nope"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusUnauthorized, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_UserGone() {
	uid := uuid.New()
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(&refreshtokens.Token{ID: uid, UserID: 99}, nil)
	s.mockUsers.EXPECT().GetByID(gomock.Any(), int64(99)).Return(nil, fmt.Errorf("user not found"))
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusUnauthorized, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_UserNotActive() {
	uid := uuid.New()
	st := users.StatusInactive
	u := &users.User{ID: 1, Approved: true, Status: &st}
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(&refreshtokens.Token{ID: uid, UserID: 1}, nil)
	s.mockUsers.EXPECT().GetByID(gomock.Any(), int64(1)).Return(u, nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"t"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusUnauthorized, rr.Code)
}

func (s *AuthHandlerTestSuite) TestRefresh_RotateLostRace() {
	uid := uuid.New()
	st := users.StatusApproved
	u := &users.User{ID: 1, Role: users.RoleNonAdmin, Approved: true, Status: &st}
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(&refreshtokens.Token{ID: uid, UserID: 1}, nil)
	s.mockUsers.EXPECT().GetByID(gomock.Any(), int64(1)).Return(u, nil)
	s.mockRefresh.EXPECT().Rotate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(uuid.Nil, refreshtokens.ErrNotFound)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":"t"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Refresh(rr, req)
	s.Equal(http.StatusUnauthorized, rr.Code)
}

func (s *AuthHandlerTestSuite) TestLogout_Method() {
	rr := httptest.NewRecorder()
	s.handler.Logout(rr, httptest.NewRequest(http.MethodGet, "/logout", nil))
	s.Equal(http.StatusMethodNotAllowed, rr.Code)
}

func (s *AuthHandlerTestSuite) TestLogout_EmptyTokenStillOK() {
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(`{"refresh_token":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Logout(rr, req)
	s.Equal(http.StatusOK, rr.Code)
	var m map[string]string
	s.NoError(json.Unmarshal(rr.Body.Bytes(), &m))
	s.Equal("ok", m["message"])
}

func (s *AuthHandlerTestSuite) TestLogout_RevokesWhenPresent() {
	raw := "logout-raw"
	pepper := s.refreshPepper
	tid := uuid.New()
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), auth.HashRefreshToken(raw, pepper)).
		Return(&refreshtokens.Token{ID: tid, UserID: 1}, nil)
	s.mockRefresh.EXPECT().RevokeByID(gomock.Any(), tid).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(`{"refresh_token":"`+raw+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Logout(rr, req)
	s.Equal(http.StatusOK, rr.Code)
}

func (s *AuthHandlerTestSuite) TestLogout_UnknownTokenStill200() {
	s.mockRefresh.EXPECT().GetActiveByHash(gomock.Any(), gomock.Any()).
		Return(nil, refreshtokens.ErrNotFound)
	req := httptest.NewRequest(http.MethodPost, "/logout", bytes.NewBufferString(`{"refresh_token":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.handler.Logout(rr, req)
	s.Equal(http.StatusOK, rr.Code)
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
