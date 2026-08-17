package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/notifications"

	"github.com/gorilla/mux"
)

// NotificationsHandler serves the authenticated user's inbox.
// Every endpoint scopes reads and writes to the caller's user id from the
// JWT context so no user can access anyone else's notifications.
type NotificationsHandler struct {
	deps *deps.Dependencies
}

// NewNotificationsHandler constructs a NotificationsHandler.
func NewNotificationsHandler(d *deps.Dependencies) *NotificationsHandler {
	return &NotificationsHandler{deps: d}
}

// ListNotificationsResponse is the shape returned by GET /api/v1/notifications.
type ListNotificationsResponse struct {
	Notifications []*notifications.Notification `json:"notifications"`
	Count         int                           `json:"count"`
	UnreadCount   int                           `json:"unread_count"`
	Limit         int32                         `json:"limit"`
	Offset        int32                         `json:"offset"`
}

// List handles GET /api/v1/notifications.
func (h *NotificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := int32(50)
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil && parsed > 0 && parsed <= 200 {
			limit = int32(parsed)
		} else {
			http.Error(w, "Invalid limit parameter (1-200)", http.StatusBadRequest)
			return
		}
	}
	var offset int32
	if v := r.URL.Query().Get("offset"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil && parsed >= 0 {
			offset = int32(parsed)
		} else {
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
	}
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	opts := notifications.ListOptions{Limit: limit, Offset: offset, UnreadOnly: unreadOnly}

	items, err := h.deps.NotificationRepository.ListForUser(r.Context(), userCtx.UserID, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list notifications: %v", err), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []*notifications.Notification{}
	}

	count, err := h.deps.NotificationRepository.CountForUser(r.Context(), userCtx.UserID, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count notifications: %v", err), http.StatusInternalServerError)
		return
	}

	unreadCount, err := h.deps.NotificationRepository.CountUnread(r.Context(), userCtx.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count unread notifications: %v", err), http.StatusInternalServerError)
		return
	}

	response := ListNotificationsResponse{
		Notifications: items,
		Count:         count,
		UnreadCount:   unreadCount,
		Limit:         limit,
		Offset:        offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UnreadCount handles GET /api/v1/notifications/unread-count.
func (h *NotificationsHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.deps.NotificationRepository.CountUnread(r.Context(), userCtx.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count unread notifications: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"unread_count": count})
}

// MarkRead handles POST /api/v1/notifications/{id}/read.
func (h *NotificationsHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "Notification id is required", http.StatusBadRequest)
		return
	}

	if err := h.deps.NotificationRepository.MarkRead(r.Context(), userCtx.UserID, id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark notification read: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkAllRead handles POST /api/v1/notifications/read-all.
func (h *NotificationsHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updated, err := h.deps.NotificationRepository.MarkAllRead(r.Context(), userCtx.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark all notifications read: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"updated": updated})
}
