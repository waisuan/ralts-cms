package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type ListMachinesResponse struct {
	Machines []*machines.Machine `json:"machines"`
	Count    int32               `json:"count"`
	Limit    int32               `json:"limit"`
	Offset   int32               `json:"offset"`
	Sort     string              `json:"sort"`
}

type MachinesHandler struct {
	deps *deps.Dependencies
}

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
		if err.Error() == "machine not found" {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get machine: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machine)
}

// ListMachines handles GET /machines
func (h *MachinesHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination and sorting
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortStr := r.URL.Query().Get("sort")
	duePPM := r.URL.Query().Get("due_ppm")

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
	var offset int32 = 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && parsedOffset >= 0 {
			offset = int32(parsedOffset)
		} else {
			http.Error(w, "Invalid offset parameter. Must be a non-negative integer", http.StatusBadRequest)
			return
		}
	}

	// Parse sort parameter
	sort := machines.SortOrderCreatedAtDesc // Default to newest first
	if sortStr != "" {
		switch sortStr {
		case "created_at_desc":
			sort = machines.SortOrderCreatedAtDesc
		case "created_at_asc":
			sort = machines.SortOrderCreatedAtAsc
		default:
			http.Error(w, "Invalid sort parameter. Must be 'created_at_desc' or 'created_at_asc'", http.StatusBadRequest)
			return
		}
	}

	// Parse due_ppm parameter
	duePPMOnly := false
	if duePPM != "" {
		if duePPM == "true" {
			duePPMOnly = true
		} else if duePPM != "false" {
			http.Error(w, "Invalid due_ppm parameter. Must be 'true' or 'false'", http.StatusBadRequest)
			return
		}
	}

	// Create list options
	options := &machines.ListOptions{
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		DuePPMOnly: duePPMOnly,
	}

	// Get machines from repository
	machines, err := h.deps.MachinesRepository.List(r.Context(), options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
		return
	}

	count, err := h.deps.MachinesRepository.Count(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := ListMachinesResponse{
		Machines: machines,
		Count:    int32(count),
		Limit:    limit,
		Offset:   offset,
		Sort:     string(sort),
	}

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

	err := h.deps.MachinesRepository.Create(r.Context(), &machine)
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			http.Error(w, "Machine already exists", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create machine: %v", err), http.StatusInternalServerError)
		return
	}

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
		if err.Error() == "machine not found" {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check machine existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MachinesRepository.Update(r.Context(), &machine)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update machine: %v", err), http.StatusInternalServerError)
		return
	}

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
		if err.Error() == "machine not found" {
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

	w.WriteHeader(http.StatusNoContent)
}
