package handlers

import (
	"encoding/json"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/refreshtokens"
	"ralts-cms/pkg/auth"
	"time"
)

// AuthHandler handles token refresh and logout.
type AuthHandler struct {
	deps *deps.Dependencies
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(d *deps.Dependencies) *AuthHandler {
	return &AuthHandler{deps: d}
}

// RefreshRequest is the body for POST /api/v1/auth/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenPairResponse is returned from login and refresh.
type TokenPairResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh issues new access and refresh tokens.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}
	pepper := h.deps.Config.RefreshPepper()
	if pepper == "" {
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}
	rawHash := auth.HashRefreshToken(body.RefreshToken, pepper)
	tok, err := h.deps.RefreshTokenRepository.GetActiveByHash(r.Context(), rawHash)
	if err != nil {
		if err == refreshtokens.ErrNotFound {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		h.deps.Logger.Error("refresh token lookup failed", "error", err)
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}
	u, err := h.deps.UsersRepository.GetByID(r.Context(), tok.UserID)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	if !u.Approved || !u.IsStatusActive() {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	newRaw, err := auth.GenerateRawRefreshToken()
	if err != nil {
		h.deps.Logger.Error("generate refresh token", "error", err)
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}
	newHash := auth.HashRefreshToken(newRaw, pepper)
	newExp := time.Now().UTC().Add(h.deps.Config.RefreshTokenLifetime)
	_, err = h.deps.RefreshTokenRepository.Rotate(r.Context(), tok.ID, u.ID, newHash, newExp)
	if err != nil {
		if err == refreshtokens.ErrNotFound {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		h.deps.Logger.Error("rotate refresh token", "error", err)
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}
	access, err := auth.GenerateAccessToken(int(u.ID), u.Role, h.deps.Config.JWTSecret, h.deps.Config.AccessTokenLifetime)
	if err != nil {
		h.deps.Logger.Error("generate access token", "error", err)
		http.Error(w, "Failed to refresh session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TokenPairResponse{Token: access, RefreshToken: newRaw})
}

// LogoutRequest is the body for POST /api/v1/auth/logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func writeLogoutOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
}

// Logout revokes a refresh token if it exists and is still active.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pepper := h.deps.Config.RefreshPepper()
	if body.RefreshToken == "" || pepper == "" {
		writeLogoutOK(w)
		return
	}
	tok, err := h.deps.RefreshTokenRepository.GetActiveByHash(r.Context(), auth.HashRefreshToken(body.RefreshToken, pepper))
	if err == nil {
		_ = h.deps.RefreshTokenRepository.RevokeByID(r.Context(), tok.ID)
	} else if err != refreshtokens.ErrNotFound {
		h.deps.Logger.Error("logout refresh lookup", "error", err)
	}
	writeLogoutOK(w)
}
