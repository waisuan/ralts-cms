package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machine"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type MachineHandler struct {
	deps *deps.Dependencies
}

func NewMachineHandler(deps *deps.Dependencies) *MachineHandler {
	return &MachineHandler{
		deps: deps,
	}
}

// GetMachine handles GET /machines/{serial_number}
func (h *MachineHandler) GetMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	machine, err := h.deps.MachineRepository.GetBySerialNumber(r.Context(), serialNumber)
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
func (h *MachineHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination and sorting
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortStr := r.URL.Query().Get("sort")

	// Parse limit parameter
	limit := h.deps.Config.DefaultMachineLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = int32(parsedLimit)
		} else {
			http.Error(w, "Invalid limit parameter. Must be between 1 and 100", http.StatusBadRequest)
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
	sort := machine.SortOrderCreatedAtDesc // Default to newest first
	if sortStr != "" {
		switch sortStr {
		case "created_at_desc":
			sort = machine.SortOrderCreatedAtDesc
		case "created_at_asc":
			sort = machine.SortOrderCreatedAtAsc
		default:
			http.Error(w, "Invalid sort parameter. Must be 'created_at_desc' or 'created_at_asc'", http.StatusBadRequest)
			return
		}
	}

	// Create list options
	options := &machine.ListOptions{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}

	// Get machines from repository
	machines, err := h.deps.MachineRepository.List(r.Context(), options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"machines": machines,
		"count":    len(machines),
		"limit":    limit,
		"offset":   offset,
		"sort":     string(sort),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateMachine handles POST /machines
func (h *MachineHandler) CreateMachine(w http.ResponseWriter, r *http.Request) {
	var machine machine.Machine
	if err := json.NewDecoder(r.Body).Decode(&machine); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if machine.SerialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	err := h.deps.MachineRepository.Create(r.Context(), &machine)
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

// UpdateMachine handles PUT /machines
func (h *MachineHandler) UpdateMachine(w http.ResponseWriter, r *http.Request) {
	var machine machine.Machine
	if err := json.NewDecoder(r.Body).Decode(&machine); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if machine.SerialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	// Check if machine exists
	_, err := h.deps.MachineRepository.GetBySerialNumber(r.Context(), machine.SerialNumber)
	if err != nil {
		if err.Error() == "machine not found" {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check machine existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MachineRepository.Update(r.Context(), &machine)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update machine: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machine)
}

// DeleteMachine handles DELETE /machines/{serial_number}
func (h *MachineHandler) DeleteMachine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serialNumber := vars["serial_number"]
	if serialNumber == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	// Check if machine exists
	_, err := h.deps.MachineRepository.GetBySerialNumber(r.Context(), serialNumber)
	if err != nil {
		if err.Error() == "machine not found" {
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check machine existence: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.deps.MachineRepository.Delete(r.Context(), serialNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete machine: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDuePPM handles GET /machines/due-ppm
func (h *MachineHandler) GetDuePPM(w http.ResponseWriter, r *http.Request) {
	machines, err := h.deps.MachineRepository.DuePPM(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get due PPM machines: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machines)
}
