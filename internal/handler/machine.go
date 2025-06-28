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
	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	pageToken := r.URL.Query().Get("page_token")

	// Use configurable default limit
	limit := h.deps.Config.DefaultMachineLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = int32(parsedLimit)
		} else {
			http.Error(w, "Invalid limit parameter. Must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}

	// Get machines from repository
	machines, nextPageToken, err := h.deps.MachineRepository.List(r.Context(), limit, pageToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"machines": machines,
		"count":    len(machines),
		"limit":    limit,
	}

	// Add next page token if there are more results
	if nextPageToken != "" {
		response["next_page_token"] = nextPageToken
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
