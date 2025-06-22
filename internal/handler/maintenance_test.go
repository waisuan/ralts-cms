package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/maintenance"
	mockmaintenance "ralts-cms/internal/maintenance"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMaintenanceTestHandler(t *testing.T) (*MaintenanceHandler, *mockmaintenance.MockRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := mockmaintenance.NewMockRepository(ctrl)
	deps := &deps.Dependencies{
		MaintenanceRepository: mockRepo,
	}
	handler := NewMaintenanceHandler(deps)
	return handler, mockRepo
}

func TestMaintenanceHandler_GetMaintenance(t *testing.T) {
	handler, mockRepo := setupMaintenanceTestHandler(t)

	t.Run("should return maintenance when found", func(t *testing.T) {
		expectedMaintenance := &maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Routine maintenance",
			ReportedBy:          "John Doe",
		}

		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.GetMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "MACHINE123", response.MachineSerialNumber)
		assert.Equal(t, "WO001", response.WorkOrderNumber)
		assert.Equal(t, "Routine maintenance", response.ActionTaken)
	})

	t.Run("should return 404 when maintenance not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.GetMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Maintenance not found")
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.GetMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to get maintenance")
	})
}

func TestMaintenanceHandler_ListMaintenance(t *testing.T) {
	handler, mockRepo := setupMaintenanceTestHandler(t)

	t.Run("should return list of maintenance records", func(t *testing.T) {
		expectedMaintenance := []*maintenance.Maintenance{
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO001",
				ActionTaken:         "First maintenance",
				ReportedBy:          "Tech1",
			},
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO002",
				ActionTaken:         "Second maintenance",
				ReportedBy:          "Tech2",
			},
		}

		mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123").Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.ListMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response []*maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "WO001", response[0].WorkOrderNumber)
		assert.Equal(t, "WO002", response[1].WorkOrderNumber)
	})

	t.Run("should return empty list when no maintenance found", func(t *testing.T) {
		mockRepo.EXPECT().ListByMachine(gomock.Any(), "NOMAINTAINANCE").Return([]*maintenance.Maintenance{}, nil)

		req := httptest.NewRequest("GET", "/machines/NOMAINTAINANCE/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.ListMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Empty(t, response)
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		mockRepo.EXPECT().ListByMachine(gomock.Any(), "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/ERROR/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.ListMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to list maintenance")
	})
}

func TestMaintenanceHandler_CreateMaintenance(t *testing.T) {
	handler, mockRepo := setupMaintenanceTestHandler(t)

	t.Run("should create maintenance successfully", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Routine maintenance",
			ReportedBy:          "John Doe",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *maintenance.Maintenance) error {
			assert.Equal(t, "MACHINE123", m.MachineSerialNumber)
			assert.Equal(t, "WO001", m.WorkOrderNumber)
			assert.Equal(t, "Routine maintenance", m.ActionTaken)
			return nil
		})

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "MACHINE123", response.MachineSerialNumber)
		assert.Equal(t, "WO001", response.WorkOrderNumber)
		assert.Equal(t, "Routine maintenance", response.ActionTaken)
	})

	t.Run("should return 400 when work order number is missing", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			ActionTaken:         "Maintenance without work order",
		}

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Work order number is required")
	})

	t.Run("should return 400 when request body is invalid", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("should return 409 when maintenance already exists", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Duplicate maintenance",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("ConditionalCheckFailedException"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Maintenance already exists")
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Error maintenance",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create maintenance")
	})
}

func TestMaintenanceHandler_UpdateMaintenance(t *testing.T) {
	handler, mockRepo := setupMaintenanceTestHandler(t)

	t.Run("should update maintenance successfully", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Updated maintenance",
			ReportedBy:          "Updated Tech",
		}

		// Mock the existence check
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the update
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *maintenance.Maintenance) error {
			assert.Equal(t, "MACHINE123", m.MachineSerialNumber)
			assert.Equal(t, "WO001", m.WorkOrderNumber)
			assert.Equal(t, "Updated maintenance", m.ActionTaken)
			return nil
		})

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "MACHINE123", response.MachineSerialNumber)
		assert.Equal(t, "WO001", response.WorkOrderNumber)
		assert.Equal(t, "Updated maintenance", response.ActionTaken)
	})

	t.Run("should return 400 when work order number is missing", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			ActionTaken:         "Maintenance without work order",
		}

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Work order number is required")
	})

	t.Run("should return 404 when maintenance not found", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "NOTFOUND",
			ActionTaken:         "Not found maintenance",
		}

		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Maintenance not found")
	})

	t.Run("should return 500 on update error", func(t *testing.T) {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Update error maintenance",
		}

		// Mock the existence check
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the update error
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update error"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update maintenance")
	})
}

func TestMaintenanceHandler_DeleteMaintenance(t *testing.T) {
	handler, mockRepo := setupMaintenanceTestHandler(t)

	t.Run("should delete maintenance successfully", func(t *testing.T) {
		// Mock the existence check
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the delete
		mockRepo.EXPECT().Delete(gomock.Any(), "MACHINE123", "WO001").Return(nil)

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return 404 when maintenance not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Maintenance not found")
	})

	t.Run("should return 500 on delete error", func(t *testing.T) {
		// Mock the existence check
		mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the delete error
		mockRepo.EXPECT().Delete(gomock.Any(), "MACHINE123", "WO001").Return(fmt.Errorf("delete error"))

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete maintenance")
	})
}
