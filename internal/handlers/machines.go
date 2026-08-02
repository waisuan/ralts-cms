package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/pkg/pgxutil"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ListMachinesResponse represents the response structure for machine listing endpoints
type ListMachinesResponse struct {
	Machines       []*machines.Machine `json:"machines"`
	OverdueCount   int32               `json:"overdue_count"`
	DueCount       int32               `json:"due_count"`
	AlmostDueCount int32               `json:"almost_due_count"`
	Count          int32               `json:"count"`
	Limit          int32               `json:"limit"`
	Offset         int32               `json:"offset"`
	Sort           string              `json:"sort"`
}

// MachinesHandler handles HTTP requests for machine-related operations
type MachinesHandler struct {
	deps *deps.Dependencies
}

// NewMachinesHandler creates a new machines handler instance with the given dependencies
func NewMachinesHandler(deps *deps.Dependencies) *MachinesHandler {
	return &MachinesHandler{
		deps: deps,
	}
}

// GetMachine handles GET /machines/{serial_number}
func (h *MachinesHandler) GetMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	machine, err := h.deps.MachinesRepository.GetBySerialNumber(r.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, machines.ErrNotFound) {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get machine: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log machine view
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionViewed, audit.ResourceMachine, serialNumber, nil))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machine)
}

// ListMachines handles GET /machines
func (h *MachinesHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination, sorting, search, and date filters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortStr := r.URL.Query().Get("sort")
	ppmStatusFilterStr := r.URL.Query().Get("ppm_status_filter")
	query := r.URL.Query().Get("q")
	// Date range filter parameters
	ppmDateFromStr := r.URL.Query().Get("ppm_date_from")
	ppmDateToStr := r.URL.Query().Get("ppm_date_to")
	tncDateFromStr := r.URL.Query().Get("tnc_date_from")
	tncDateToStr := r.URL.Query().Get("tnc_date_to")

	// Parse limit parameter
	limit := h.deps.Config.DefaultMachinesLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= h.deps.Config.MaxMachinesLimit {
			limit = int32(parsedLimit)
		} else {
			http.Error(w, fmt.Sprintf("Invalid limit parameter. Must be between 1 and %d", h.deps.Config.MaxMachinesLimit), http.StatusBadRequest)
			return
		}
	}

	// Parse offset parameter
	var offset int32
	if offsetStr != "" {
		if parsedOffset, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && parsedOffset >= 0 {
			offset = int32(parsedOffset)
		} else {
			http.Error(w, "Invalid offset parameter. Must be a non-negative integer", http.StatusBadRequest)
			return
		}
	}

	// Parse sort parameter
	sort := machines.SortOrderUpdatedAtDesc // Default to most recently updated first
	if sortStr != "" {
		switch sortStr {
		case "updated_at_desc":
			sort = machines.SortOrderUpdatedAtDesc
		case "updated_at_asc":
			sort = machines.SortOrderUpdatedAtAsc
		case "ppm_date_asc":
			sort = machines.SortOrderPpmDateAsc
		case "ppm_date_desc":
			sort = machines.SortOrderPpmDateDesc
		case "tnc_date_asc":
			sort = machines.SortOrderTncDateAsc
		case "tnc_date_desc":
			sort = machines.SortOrderTncDateDesc
		default:
			http.Error(w, "Invalid sort parameter. Valid options: 'updated_at_desc', 'updated_at_asc', 'ppm_date_asc', 'ppm_date_desc', 'tnc_date_asc', 'tnc_date_desc'", http.StatusBadRequest)
			return
		}
	}

	// Parse ppm_status_filter parameter
	var ppmStatusFilter machines.PPMStatus
	if ppmStatusFilterStr != "" {
		switch ppmStatusFilterStr {
		case "overdue":
			ppmStatusFilter = machines.PPMStatusOverdue
		case "due":
			ppmStatusFilter = machines.PPMStatusDue
		case "almost_due":
			ppmStatusFilter = machines.PPMStatusAlmostDue
		default:
			http.Error(w, "Invalid ppm_status_filter parameter. Must be 'overdue', 'due', or 'almost_due'", http.StatusBadRequest)
			return
		}
	}

	// Parse date range parameters (format: YYYY-MM-DD)
	var ppmDateFrom, ppmDateTo, tncDateFrom, tncDateTo *time.Time
	if ppmDateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", ppmDateFromStr); err == nil {
			ppmDateFrom = &parsed
		} else {
			http.Error(w, "Invalid ppm_date_from format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if ppmDateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", ppmDateToStr); err == nil {
			ppmDateTo = &parsed
		} else {
			http.Error(w, "Invalid ppm_date_to format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if tncDateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", tncDateFromStr); err == nil {
			tncDateFrom = &parsed
		} else {
			http.Error(w, "Invalid tnc_date_from format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if tncDateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", tncDateToStr); err == nil {
			tncDateTo = &parsed
		} else {
			http.Error(w, "Invalid tnc_date_to format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	// Create list options
	options := &machines.ListOptions{
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		PpmStatusFilter: ppmStatusFilter,
		PpmDateFrom:     ppmDateFrom,
		PpmDateTo:       ppmDateTo,
		TncDateFrom:     tncDateFrom,
		TncDateTo:       tncDateTo,
	}

	// Get machines from repository (search if query provided, otherwise list)
	var machines []*machines.Machine
	var count int
	var err error
	if query != "" {
		machines, err = h.deps.MachinesRepository.Search(r.Context(), query, options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to search machines: %v", err), http.StatusInternalServerError)
			return
		}

		// Count search results
		count, err = h.deps.MachinesRepository.CountSearch(r.Context(), query, options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count search results: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		machines, err = h.deps.MachinesRepository.List(r.Context(), options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
			return
		}

		count, err = h.deps.MachinesRepository.Count(r.Context(), options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count machines: %v", err), http.StatusInternalServerError)
			return
		}
	}

	overdueCount, dueCount, almostDueCount, err := h.deps.MachinesRepository.CountByStatus(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count machines by status: %v", err), http.StatusInternalServerError)
		return
	}

	if len(machines) > 0 {
		serials := make([]string, len(machines))
		for i, m := range machines {
			serials[i] = m.SerialNumber
		}
		counts, err := h.deps.MaintenanceRepository.CountByMachineSerials(r.Context(), serials)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count maintenance records for machines: %v", err), http.StatusInternalServerError)
			return
		}
		for _, machine := range machines {
			machine.MaintenanceCount = counts[machine.SerialNumber]
		}
	}

	// Build response
	response := ListMachinesResponse{
		Machines:       machines,
		OverdueCount:   overdueCount,
		DueCount:       dueCount,
		AlmostDueCount: almostDueCount,
		Count:          int32(count),
		Limit:          limit,
		Offset:         offset,
		Sort:           string(sort),
	}

	// Audit: Log machine list/search
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionListed, audit.ResourceMachine, "", map[string]any{
		"count":  count,
		"limit":  limit,
		"offset": offset,
		"query":  query,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateMachine handles POST /machines
func (h *MachinesHandler) CreateMachine(w http.ResponseWriter, r *http.Request) {
	var machine machines.Machine
	if err := json.NewDecoder(r.Body).Decode(&machine); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if machine.SerialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	// updated_by reflects the authenticated user, never client-supplied input.
	machine.UpdatedBy = resolveUpdatedByUsername(r.Context(), h.deps)

	err := h.deps.MachinesRepository.Create(r.Context(), &machine)
	if err != nil {
		if pgxutil.IsUniqueViolation(err) {
			http.Error(w, "Machine already exists", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create machine: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log machine creation
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionCreated, audit.ResourceMachine, machine.SerialNumber, map[string]any{
		"customer": machine.Customer,
		"model":    machine.Model,
		"brand":    machine.Brand,
	}))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(machine)
}

// UpdateMachine handles PUT /machines/{serial_number}
func (h *MachinesHandler) UpdateMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	var machine machines.Machine
	if err := json.NewDecoder(r.Body).Decode(&machine); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if machine.SerialNumber != serialNumber {
		http.Error(w, "Serial number in request body does not match URL path", http.StatusBadRequest)
		return
	}

	// Check if machine exists
	_, err := h.deps.MachinesRepository.GetBySerialNumber(r.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, machines.ErrNotFound) {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check machine existence: %v", err), http.StatusInternalServerError)
		return
	}

	// updated_by reflects the authenticated user, never client-supplied input.
	machine.UpdatedBy = resolveUpdatedByUsername(r.Context(), h.deps)

	err = h.deps.MachinesRepository.Update(r.Context(), &machine)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update machine: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log machine update
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionUpdated, audit.ResourceMachine, machine.SerialNumber, map[string]any{
		"customer": machine.Customer,
		"model":    machine.Model,
		"brand":    machine.Brand,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machine)
}

// DeleteMachine handles DELETE /machines/{serial_number}
func (h *MachinesHandler) DeleteMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	// Check if machine exists
	_, err := h.deps.MachinesRepository.GetBySerialNumber(r.Context(), serialNumber)
	if err != nil {
		if errors.Is(err, machines.ErrNotFound) {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check machine existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MachinesRepository.Delete(r.Context(), serialNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete machine: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log machine deletion
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionDeleted, audit.ResourceMachine, serialNumber, nil))

	w.WriteHeader(http.StatusNoContent)
}
