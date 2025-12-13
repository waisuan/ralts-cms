package handlers

import (
	"encoding/json"
	"net/http"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"strconv"
	"time"
)

// AuditHandler handles HTTP requests for audit-related operations
type AuditHandler struct {
	deps *deps.Dependencies
}

// ListAuditEventsResponse represents the response structure for audit event listing
type ListAuditEventsResponse struct {
	Events     []*audit.Event `json:"events"`
	TotalCount int32          `json:"total_count"`
	Limit      int32          `json:"limit"`
	Offset     int32          `json:"offset"`
}

// NewAuditHandler creates a new audit handler instance with the given dependencies
func NewAuditHandler(deps *deps.Dependencies) *AuditHandler {
	return &AuditHandler{
		deps: deps,
	}
}

// ListAuditEvents handles GET /api/v1/admin/audit/events for listing audit events
// Query parameters:
//   - limit: number of events to return (default: 50, max: 100)
//   - offset: number of events to skip (default: 0)
//   - from_date: filter events from this date (ISO 8601 format)
//   - to_date: filter events until this date (ISO 8601 format)
//   - resource_type: filter by resource type (machine, maintenance, user, attachment)
//   - action: filter by action (created, updated, deleted, viewed, listed, login, logout, password_changed)
func (h *AuditHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query()

	// Parse limit with default and max
	limit := int32(50)
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsed, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(parsed)
			if limit > 100 {
				limit = 100
			}
			if limit < 1 {
				limit = 50
			}
		}
	}

	// Parse offset
	offset := int32(0)
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsed, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(parsed)
			if offset < 0 {
				offset = 0
			}
		}
	}

	// Build list options
	options := &audit.ListOptions{
		Limit:  limit,
		Offset: offset,
	}

	// Parse date filters
	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			options.FromDate = &fromDate
		}
	}

	if toDateStr := query.Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			options.ToDate = &toDate
		}
	}

	// Parse resource type filter
	if resourceType := query.Get("resource_type"); resourceType != "" {
		options.ResourceType = &resourceType
	}

	// Parse action filter
	if action := query.Get("action"); action != "" {
		options.Action = &action
	}

	// Fetch events from repository
	events, err := h.deps.AuditRepository.List(ctx, options)
	if err != nil {
		h.deps.Logger.Error("Failed to list audit events", "error", err)
		http.Error(w, "Failed to retrieve audit events", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	totalCount, err := h.deps.AuditRepository.Count(ctx, options)
	if err != nil {
		h.deps.Logger.Error("Failed to count audit events", "error", err)
		http.Error(w, "Failed to count audit events", http.StatusInternalServerError)
		return
	}

	// Build response
	response := ListAuditEventsResponse{
		Events:     events,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.deps.Logger.Error("Failed to encode audit events response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// StreamAuditEvents handles GET /api/v1/admin/audit/events/stream for streaming audit events via SSE
// This is a stub endpoint that will be implemented in a future iteration
func (h *AuditHandler) StreamAuditEvents(w http.ResponseWriter, r *http.Request) {
	// SSE streaming will be implemented in a future iteration
	// For now, return 501 Not Implemented
	http.Error(w, "Server-Sent Events streaming is not yet implemented", http.StatusNotImplemented)
}
