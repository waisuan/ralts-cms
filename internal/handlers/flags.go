package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/flags"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/users"

	"github.com/gorilla/mux"
)

// FlagsHandler serves the machine-flags API.
//
// Per-machine listing is available to any authenticated user (so the frontend
// can render badges). Creation requires role=ADMIN. Resolution is allowed for
// admins and for the machine's assignee, so a technician can clear a flag they
// were asked to deal with. The cross-machine List is open to any authenticated
// user but returns only the caller's assigned machines unless they are an
// admin. All of this is enforced inside the handlers using the JWT user
// context, since role-scoped reads are beyond what middleware can express.
type FlagsHandler struct {
	deps *deps.Dependencies
}

// NewFlagsHandler constructs a FlagsHandler.
func NewFlagsHandler(d *deps.Dependencies) *FlagsHandler {
	return &FlagsHandler{deps: d}
}

// BatchOpenByMachineResponse is the shape returned by
// GET /api/v1/machines/flags/open-by-machine?serials=...
// Keys are serial numbers; the value is the list of open flags for that
// machine so the frontend can render badges without a request per row.
type BatchOpenByMachineResponse struct {
	Flags map[string][]*flags.Flag `json:"flags"`
}

// ListOpenByMachine handles GET /api/v1/machines/flags/open-by-machine.
// The `serials` query parameter is a comma-separated list of machine serial
// numbers to look up in a single batch.
func (h *FlagsHandler) ListOpenByMachine(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("serials")
	if raw == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BatchOpenByMachineResponse{Flags: map[string][]*flags.Flag{}})
		return
	}

	serials := splitAndTrim(raw)
	if len(serials) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BatchOpenByMachineResponse{Flags: map[string][]*flags.Flag{}})
		return
	}

	// Bound to protect the backend if a client tries something silly.
	const maxSerials = 200
	if len(serials) > maxSerials {
		http.Error(w, fmt.Sprintf("Too many serials (max %d)", maxSerials), http.StatusBadRequest)
		return
	}

	m, err := h.deps.FlagsRepository.ListOpenBySerials(r.Context(), serials)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load flags: %v", err), http.StatusInternalServerError)
		return
	}
	if m == nil {
		m = map[string][]*flags.Flag{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BatchOpenByMachineResponse{Flags: m})
}

// splitAndTrim splits a comma-separated string and trims whitespace,
// dropping empty tokens.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// CreateFlagRequest is the JSON body for POST /api/v1/machines/{serial}/flags.
type CreateFlagRequest struct {
	Reason flags.Reason `json:"reason"`
	Note   string       `json:"note"`
}

// ListFlagsResponse is the shape returned by GET /api/v1/flags.
type ListFlagsResponse struct {
	Flags  []*flags.Flag `json:"flags"`
	Count  int           `json:"count"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
	// Scope is "all" for admins and "assigned" for everyone else, so the page can
	// say which set of flags it is showing.
	Scope string `json:"scope"`
}

// Flag listing scopes reported in ListFlagsResponse.
const (
	FlagScopeAll      = "all"
	FlagScopeAssigned = "assigned"
)

// List handles GET /api/v1/flags, the cross-machine flagged-records view.
// Admins see every flag; everyone else sees only flags on machines assigned to
// them, which is the same set they are allowed to resolve, and of the resolved
// ones only those they closed themselves.
//
// Query parameters:
//   - status: open (default), resolved, or all
//   - limit: 1-200, default 50
//   - offset: default 0
func (h *FlagsHandler) List(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	opts := flags.ListOptions{Status: flags.StatusOpen, Limit: 50}
	scope := FlagScopeAll
	if userCtx.Role != users.RoleAdmin {
		scope = FlagScopeAssigned
		opts.AssignedUserID = &userCtx.UserID
		// Being assigned a machine should not hand over the record of what other
		// people closed on it beforehand, so resolved rows are limited to the
		// caller's own. Open flags, which are the work itself, are untouched.
		opts.OwnResolutionsUserID = &userCtx.UserID
	}

	switch status := r.URL.Query().Get("status"); status {
	case "", string(flags.StatusOpen):
		// Keep the default: open flags are the ones needing attention.
	case string(flags.StatusResolved):
		opts.Status = flags.StatusResolved
	case "all":
		opts.Status = ""
	default:
		http.Error(w, "Invalid status parameter (open, resolved, all)", http.StatusBadRequest)
		return
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed < 1 || parsed > 200 {
			http.Error(w, "Invalid limit parameter (1-200)", http.StatusBadRequest)
			return
		}
		opts.Limit = int32(parsed)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed < 0 {
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		opts.Offset = int32(parsed)
	}

	list, err := h.deps.FlagsRepository.List(r.Context(), opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list flags: %v", err), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*flags.Flag{}
	}

	count, err := h.deps.FlagsRepository.Count(r.Context(), opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count flags: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListFlagsResponse{
		Flags:  list,
		Count:  count,
		Limit:  opts.Limit,
		Offset: opts.Offset,
		Scope:  scope,
	})
}

// ListForMachine handles GET /api/v1/machines/{serial}/flags.
func (h *FlagsHandler) ListForMachine(w http.ResponseWriter, r *http.Request) {
	serial := mux.Vars(r)["serial_number"]
	if serial == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}
	includeResolved := r.URL.Query().Get("include_resolved") == "true"

	list, err := h.deps.FlagsRepository.ListForMachine(r.Context(), serial, includeResolved)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list flags: %v", err), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*flags.Flag{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"flags": list})
}

// Create handles POST /api/v1/machines/{serial}/flags (admin only).
//
// A machine holds one open flag at a time: flagging a machine that already has
// one rewrites it and answers 200 instead of 201.
func (h *FlagsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if userCtx.Role != users.RoleAdmin {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	serial := mux.Vars(r)["serial_number"]
	if serial == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	var req CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if h.deps.FlagsService == nil {
		http.Error(w, "Flags service not available", http.StatusServiceUnavailable)
		return
	}

	flag, isNew, err := h.deps.FlagsService.CreateManualFlag(r.Context(), userCtx.UserID, serial, req.Reason, req.Note)
	if err != nil {
		if errors.Is(err, machines.ErrNotFound) {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// 200 when the machine's existing open flag was rewritten rather than a new
	// one created, so a client can tell the two apart.
	if isNew {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(flag)
}

// ResolveFlagRequest is the optional JSON body for the resolve endpoint.
type ResolveFlagRequest struct {
	// Note records what was done about the flag. Optional.
	Note string `json:"note"`
}

// decodeResolveNote reads the optional resolution note. Resolving with no body
// at all is valid and means "no note", so an empty body is not an error.
func decodeResolveNote(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	var req ResolveFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return "", nil
		}
		return "", fmt.Errorf("invalid request body: %w", err)
	}
	return req.Note, nil
}

// Resolve handles POST /api/v1/machines/{serial}/flags/{id}/resolve.
//
// Admins may resolve any flag. Other users may resolve flags on machines
// currently assigned to them, which is exactly the set List shows them, so
// everything on their flagged-records page is actionable. An optional JSON body
// of {"note": "..."} records what was done. When the resolver is not an admin,
// the admin who raised the flag gets an inbox notification.
func (h *FlagsHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	userCtx, err := appctx.GetUserFromContext(r.Context())
	if err != nil || userCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Flag id is required", http.StatusBadRequest)
		return
	}

	if h.deps.FlagsService == nil {
		http.Error(w, "Flags service not available", http.StatusServiceUnavailable)
		return
	}

	isAdmin := userCtx.Role == users.RoleAdmin
	allowed, err := h.deps.FlagsService.CanResolve(r.Context(), userCtx.UserID, isAdmin, id)
	if err != nil {
		if errors.Is(err, flags.ErrNotFound) {
			http.Error(w, "Flag not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, machines.ErrNotFound) {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to resolve flag: %v", err), http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "You can only resolve flags on machines assigned to you", http.StatusForbidden)
		return
	}

	note, err := decodeResolveNote(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.deps.FlagsService.Resolve(r.Context(), userCtx.UserID, isAdmin, id, note); err != nil {
		if errors.Is(err, flags.ErrNotFound) {
			http.Error(w, "Flag not found or already resolved", http.StatusNotFound)
			return
		}
		if errors.Is(err, flags.ErrNoteTooLong) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to resolve flag: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
