package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/maintenance"
	"strings"
)

type MaintenanceHandler struct {
	deps *deps.Dependencies
}

func NewMaintenanceHandler(deps *deps.Dependencies) *MaintenanceHandler {
	return &MaintenanceHandler{
		deps: deps,
	}
}

// GetMaintenance handles GET /machines/:serial_number/maintenance/:work_order_number
func (h *MaintenanceHandler) GetMaintenance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/machines/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
		http.Error(w, "Machine serial number and work order number are required", http.StatusBadRequest)
		return
	}

	machineSerialNumber := parts[0]
	workOrderNumber := parts[2]

	maintenance, err := h.deps.MaintenanceRepository.GetByWorkOrder(r.Context(), machineSerialNumber, workOrderNumber)
	if err != nil {
		if err.Error() == "maintenance not found" {
			http.Error(w, "Maintenance not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maintenance)
}

// ListMaintenance handles GET /machines/:serial_number/maintenance
func (h *MaintenanceHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/machines/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	machineSerialNumber := parts[0]

	maintenanceList, err := h.deps.MaintenanceRepository.ListByMachine(r.Context(), machineSerialNumber)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list maintenance: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maintenanceList)
}

// CreateMaintenance handles POST /machines/:serial_number/maintenance
func (h *MaintenanceHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/machines/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	machineSerialNumber := parts[0]

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(maintenance)
}

// UpdateMaintenance handles PUT /machines/:serial_number/maintenance
func (h *MaintenanceHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/machines/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Machine serial number is required", http.StatusBadRequest)
		return
	}

	machineSerialNumber := parts[0]

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(maintenance)
}

// DeleteMaintenance handles DELETE /machines/:serial_number/maintenance/:work_order_number
func (h *MaintenanceHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/machines/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
		http.Error(w, "Machine serial number and work order number are required", http.StatusBadRequest)
		return
	}

	machineSerialNumber := parts[0]
	workOrderNumber := parts[2]

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

	w.WriteHeader(http.StatusNoContent)
}
