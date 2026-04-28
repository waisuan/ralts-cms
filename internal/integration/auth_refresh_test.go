//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"ralts-cms/internal/users"
)

func (s *IntegrationSuite) TestAuthRefresh() {
	s.createApprovedUser("refreshuser", "secret123", "refresh@example.com", users.RoleAdmin)
	_ = s.login("refreshuser", "secret123")

	resp := s.postJSON("/api/v1/auth/refresh", map[string]string{
		"refresh_token": "invalid",
	}, "")
	defer resp.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	var loginOut struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	body, err := json.Marshal(map[string]string{"username": "refreshuser", "password": "secret123"})
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.apiURL("/api/v1/users/login"), bytes.NewReader(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	loginResp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer loginResp.Body.Close()
	raw, err := io.ReadAll(loginResp.Body)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, loginResp.StatusCode, string(raw))
	s.Require().NoError(json.Unmarshal(raw, &loginOut))
	s.Require().NotEmpty(loginOut.RefreshToken)

	refreshResp := s.postJSON("/api/v1/auth/refresh", map[string]string{
		"refresh_token": loginOut.RefreshToken,
	}, "")
	defer refreshResp.Body.Close()
	refreshRaw, err := io.ReadAll(refreshResp.Body)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, refreshResp.StatusCode, string(refreshRaw))
	var pair struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	s.Require().NoError(json.Unmarshal(refreshRaw, &pair))
	s.Require().NotEmpty(pair.Token)
	s.Require().NotEmpty(pair.RefreshToken)
	// New opaque refresh is always different after rotation; access JWT may
	// repeat if issued in the same second.
	s.Require().NotEqual(loginOut.RefreshToken, pair.RefreshToken)

	// Old refresh is one-time use
	replay := s.postJSON("/api/v1/auth/refresh", map[string]string{
		"refresh_token": loginOut.RefreshToken,
	}, "")
	defer replay.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, replay.StatusCode)

	// New access works on protected API
	m := s.get("/api/v1/machines", pair.Token)
	defer m.Body.Close()
	s.Require().Equal(http.StatusOK, m.StatusCode)
}

func (s *IntegrationSuite) TestAuthLogout_RevokesRefresh() {
	s.createApprovedUser("logoutu", "pw1234!!!", "logout@example.com", users.RoleNonAdmin)
	login := s.postJSON("/api/v1/users/login", map[string]string{
		"username": "logoutu",
		"password": "pw1234!!!",
	}, "")
	defer login.Body.Close()
	body, _ := io.ReadAll(login.Body)
	s.Require().Equal(http.StatusOK, login.StatusCode, string(body))
	var lo struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	s.Require().NoError(json.Unmarshal(body, &lo))

	logout := s.postJSON("/api/v1/auth/logout", map[string]string{
		"refresh_token": lo.RefreshToken,
	}, "")
	defer logout.Body.Close()
	s.Require().Equal(http.StatusOK, logout.StatusCode)

	again := s.postJSON("/api/v1/auth/refresh", map[string]string{
		"refresh_token": lo.RefreshToken,
	}, "")
	defer again.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, again.StatusCode)
}

func (s *IntegrationSuite) TestAuthPasswordChange_RevokesAllRefreshSessions() {
	s.createApprovedUser("pwchuser", "oldpass99", "pwch@example.com", users.RoleNonAdmin)
	login1 := s.postJSON("/api/v1/users/login", map[string]string{
		"username": "pwchuser",
		"password": "oldpass99",
	}, "")
	defer login1.Body.Close()
	b0, _ := io.ReadAll(login1.Body)
	s.Require().Equal(http.StatusOK, login1.StatusCode, string(b0))
	var creds struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	s.Require().NoError(json.Unmarshal(b0, &creds))
	s.Require().NotEmpty(creds.RefreshToken)

	put := s.putJSON("/api/v1/users/password", map[string]string{
		"current_password": "oldpass99",
		"new_password":     "newpass99!",
	}, creds.Token)
	defer put.Body.Close()
	putBody, _ := io.ReadAll(put.Body)
	s.Require().Equal(http.StatusOK, put.StatusCode, string(putBody))

	refreshAfter := s.postJSON("/api/v1/auth/refresh", map[string]string{
		"refresh_token": creds.RefreshToken,
	}, "")
	defer refreshAfter.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, refreshAfter.StatusCode)

	tok1 := s.login("pwchuser", "newpass99!")
	gr := s.get("/api/v1/machines", tok1)
	defer gr.Body.Close()
	s.Require().Equal(http.StatusOK, gr.StatusCode)
}
