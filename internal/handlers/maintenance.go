package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/maintenance"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// ListMaintenanceResponse represents the response structure for maintenance listing endpoints
type ListMaintenanceResponse struct {
	Maintenance       []*maintenance.Maintenance `json:"maintenance"`
	PreventativeCount int32                      `json:"preventative_count"`
	CorrectiveCount   int32                      `json:"corrective_count"`
	EmergencyCount    int32                      `json:"emergency_count"`
	InspectionCount   int32                      `json:"inspection_count"`
	OtherCount        int32                      `json:"other_count"`
	Count             int32                      `json:"count"`
	Limit             int32                      `json:"limit"`
	Offset            int32                      `json:"offset"`
	Sort              string                     `json:"sort"`
}

// MaintenanceHandler handles HTTP requests for maintenance-related operations
type MaintenanceHandler struct {
	deps *deps.Dependencies
}

// NewMaintenanceHandler creates a new maintenance handler instance with the given dependencies
func NewMaintenanceHandler(deps *deps.Dependencies) *MaintenanceHandler {
	return &MaintenanceHandler{
		deps: deps,
	}
}

// GetMaintenance handles GET /machines/{serial_number}/maintenance/{work_order_number}
func (h *MaintenanceHandler) GetMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]
	if machineSerialNumber == "" || workOrderNumber == "" {
		http.Error(w, "Machine serial number and work order number are required", http.StatusBadRequest)
		return
	}

	maintenance, err := h.deps.MaintenanceRepository.GetByWorkOrder(r.Context(), machineSerialNumber, workOrderNumber)
	if err != nil {
		if err.Error() == "maintenance not found" {
			http.Error(w, "Maintenance not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance view
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionViewed, audit.ResourceMaintenance, workOrderNumber, map[string]any{
		"machine_serial_number": machineSerialNumber,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maintenance)
}

// ListMaintenance handles GET /machines/{serial_number}/maintenance
func (h *MaintenanceHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	if machineSerialNumber == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters for pagination, sorting, and search
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortStr := r.URL.Query().Get("sort")
	query := r.URL.Query().Get("q")

	// Parse limit parameter
	limit := h.deps.Config.DefaultMaintenanceLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= h.deps.Config.MaxMaintenanceLimit {
			limit = int32(parsedLimit)
		} else {
			http.Error(w, fmt.Sprintf("Invalid limit parameter. Must be between 1 and %d", h.deps.Config.MaxMaintenanceLimit), http.StatusBadRequest)
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
	sort := maintenance.SortOrderUpdatedAtDesc // Default to most recently updated first
	if sortStr != "" {
		switch sortStr {
		case "work_order_date_desc":
			sort = maintenance.SortOrderWorkOrderDateDesc
		case "work_order_date_asc":
			sort = maintenance.SortOrderWorkOrderDateAsc
		case "updated_at_desc":
			sort = maintenance.SortOrderUpdatedAtDesc
		case "updated_at_asc":
			sort = maintenance.SortOrderUpdatedAtAsc
		default:
			http.Error(w, "Invalid sort parameter. Must be 'work_order_date_desc', 'work_order_date_asc', 'updated_at_desc', or 'updated_at_asc'", http.StatusBadRequest)
			return
		}
	}

	// Create list options
	options := &maintenance.ListOptions{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}

	// Get maintenance records from repository
	var maintenanceList []*maintenance.Maintenance
	var count int
	var err error

	if query != "" {
		// Use general full-text search filtered by machine
		maintenanceList, err = h.deps.MaintenanceRepository.SearchByMachine(r.Context(), machineSerialNumber, query, options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to search maintenance: %v", err), http.StatusInternalServerError)
			return
		}

		count, err = h.deps.MaintenanceRepository.CountSearchByMachine(r.Context(), machineSerialNumber, query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count search results: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// No search, just list by machine
		maintenanceList, err = h.deps.MaintenanceRepository.ListByMachine(r.Context(), machineSerialNumber, options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list maintenance: %v", err), http.StatusInternalServerError)
			return
		}

		count, err = h.deps.MaintenanceRepository.CountByMachine(r.Context(), machineSerialNumber)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to count maintenance: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get counts by work order type
	preventativeCount, correctiveCount, emergencyCount, inspectionCount, otherCount, err := h.deps.MaintenanceRepository.CountByWorkOrderType(r.Context(), machineSerialNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count maintenance by work order type: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := ListMaintenanceResponse{
		Maintenance:       maintenanceList,
		PreventativeCount: int32(preventativeCount),
		CorrectiveCount:   int32(correctiveCount),
		EmergencyCount:    int32(emergencyCount),
		InspectionCount:   int32(inspectionCount),
		OtherCount:        int32(otherCount),
		Count:             int32(count),
		Limit:             limit,
		Offset:            offset,
		Sort:              string(sort),
	}

	// Audit: Log maintenance list/search
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionListed, audit.ResourceMaintenance, "", map[string]any{
		"machine_serial_number": machineSerialNumber,
		"count":                 count,
		"limit":                 limit,
		"offset":                offset,
		"query":                 query,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateMaintenance handles POST /machines/{serial_number}/maintenance
func (h *MaintenanceHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	if machineSerialNumber == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	var maintenance maintenance.Maintenance
	if err := json.NewDecoder(r.Body).Decode(&maintenance); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set the machine serial number from the URL
	maintenance.MachineSerialNumber = machineSerialNumber

	if maintenance.WorkOrderNumber == "" {
		http.Error(w, "Work order number is required", http.StatusBadRequest)
		return
	}

	err := h.deps.MaintenanceRepository.Create(r.Context(), &maintenance)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			http.Error(w, "Maintenance already exists", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance creation
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionCreated, audit.ResourceMaintenance, maintenance.WorkOrderNumber, map[string]any{
		"machine_serial_number": maintenance.MachineSerialNumber,
		"work_order_type":       maintenance.WorkOrderType,
	}))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(maintenance)
}

// UpdateMaintenance handles PUT /machines/{serial_number}/maintenance
func (h *MaintenanceHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	if machineSerialNumber == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	var maintenance maintenance.Maintenance
	if err := json.NewDecoder(r.Body).Decode(&maintenance); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set the machine serial number from the URL
	maintenance.MachineSerialNumber = machineSerialNumber

	if maintenance.WorkOrderNumber == "" {
		http.Error(w, "Work order number is required", http.StatusBadRequest)
		return
	}

	// Check if maintenance exists
	_, err := h.deps.MaintenanceRepository.GetByWorkOrder(r.Context(), machineSerialNumber, maintenance.WorkOrderNumber)
	if err != nil {
		if err.Error() == "maintenance not found" {
			http.Error(w, "Maintenance not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check maintenance existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MaintenanceRepository.Update(r.Context(), &maintenance)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance update
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionUpdated, audit.ResourceMaintenance, maintenance.WorkOrderNumber, map[string]any{
		"machine_serial_number": maintenance.MachineSerialNumber,
		"work_order_type":       maintenance.WorkOrderType,
	}))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maintenance)
}

// DeleteMaintenance handles DELETE /machines/{serial_number}/maintenance/{work_order_number}
func (h *MaintenanceHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineSerialNumber := vars["serial_number"]
	workOrderNumber := vars["work_order_number"]
	if machineSerialNumber == "" || workOrderNumber == "" {
		http.Error(w, "Machine serial number and work order number are required", http.StatusBadRequest)
		return
	}

	// Check if maintenance exists
	_, err := h.deps.MaintenanceRepository.GetByWorkOrder(r.Context(), machineSerialNumber, workOrderNumber)
	if err != nil {
		if err.Error() == "maintenance not found" {
			http.Error(w, "Maintenance not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check maintenance existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MaintenanceRepository.Delete(r.Context(), machineSerialNumber, workOrderNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	// Audit: Log maintenance deletion
	h.deps.AuditService.LogEvent(audit.NewEvent(r, audit.ActionDeleted, audit.ResourceMaintenance, workOrderNumber, map[string]any{
		"machine_serial_number": machineSerialNumber,
	}))

	w.WriteHeader(http.StatusNoContent)
}
