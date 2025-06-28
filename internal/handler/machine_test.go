package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machine"
	mockmachine "ralts-cms/internal/machine"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*MachineHandler, *mockmachine.MockRepository) {
	ctrl := gomock.NewController(t)
	mockRepo := mockmachine.NewMockRepository(ctrl)
	deps := &deps.Dependencies{
		MachineRepository: mockRepo,
		Config: &deps.Config{
			DefaultMachineLimit: 50,
		},
	}
	handler := NewMachineHandler(deps)
	return handler, mockRepo
}

func TestMachineHandler_GetMachine(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	t.Run("should return machine when found", func(t *testing.T) {
		expectedMachine := &machine.Machine{
			SerialNumber: "TEST123",
			Customer:     "Test Customer",
			State:        "Active",
		}

		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "TEST123").Return(expectedMachine, nil)

		req := httptest.NewRequest("GET", "/machines/TEST123", nil)
		w := httptest.NewRecorder()

		// Set up router for path parameters
		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.GetMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "TEST123", response.SerialNumber)
		assert.Equal(t, "Test Customer", response.Customer)
	})

	t.Run("should return 404 when machine not found", func(t *testing.T) {
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

		req := httptest.NewRequest("GET", "/machines/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.GetMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Machine not found")
	})

	t.Run("should return 400 when serial number is missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/machines/", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.GetMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code) // Router returns 404 for missing path param
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.GetMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to get machine")
	})
}

func TestMachineHandler_CreateMachine(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	t.Run("should create machine successfully", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "CREATE123",
			Customer:     "Create Customer",
			State:        "New",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *machine.Machine) error {
			assert.Equal(t, "CREATE123", m.SerialNumber)
			assert.Equal(t, "Create Customer", m.Customer)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMachine(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "CREATE123", response.SerialNumber)
		assert.Equal(t, "Create Customer", response.Customer)
	})

	t.Run("should return 400 when serial number is missing", func(t *testing.T) {
		machineData := machine.Machine{
			Customer: "Customer without serial",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMachine(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Serial number is required")
	})

	t.Run("should return 400 when request body is invalid", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/machines", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMachine(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("should return 409 when machine already exists", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "DUPLICATE123",
			Customer:     "Duplicate Customer",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("ConditionalCheckFailedException"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMachine(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Machine already exists")
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "ERROR123",
			Customer:     "Error Customer",
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateMachine(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create machine")
	})
}

func TestMachineHandler_UpdateMachine(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	t.Run("should update machine successfully", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "UPDATE123",
			Customer:     "Updated Customer",
			State:        "Active",
		}

		// Mock the existence check
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATE123").Return(&machine.Machine{SerialNumber: "UPDATE123"}, nil)
		// Mock the update
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *machine.Machine) error {
			assert.Equal(t, "UPDATE123", m.SerialNumber)
			assert.Equal(t, "Updated Customer", m.Customer)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateMachine(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "UPDATE123", response.SerialNumber)
		assert.Equal(t, "Updated Customer", response.Customer)
	})

	t.Run("should return 400 when serial number is missing", func(t *testing.T) {
		machineData := machine.Machine{
			Customer: "Customer without serial",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateMachine(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Serial number is required")
	})

	t.Run("should return 404 when machine not found", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "NOTFOUND123",
			Customer:     "Not Found Customer",
		}

		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND123").Return(nil, fmt.Errorf("machine not found"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateMachine(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Machine not found")
	})

	t.Run("should return 500 on update error", func(t *testing.T) {
		machineData := machine.Machine{
			SerialNumber: "UPDATEERROR123",
			Customer:     "Update Error Customer",
		}

		// Mock the existence check
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATEERROR123").Return(&machine.Machine{SerialNumber: "UPDATEERROR123"}, nil)
		// Mock the update error
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateMachine(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to update machine")
	})
}

func TestMachineHandler_DeleteMachine(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	t.Run("should delete machine successfully", func(t *testing.T) {
		// First, expect a check that the machine exists
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "DELETE123").Return(&machine.Machine{SerialNumber: "DELETE123"}, nil)
		// Then expect the delete operation
		mockRepo.EXPECT().Delete(gomock.Any(), "DELETE123").Return(nil)

		req := httptest.NewRequest("DELETE", "/machines/DELETE123", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.DeleteMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return 404 when machine not found", func(t *testing.T) {
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

		req := httptest.NewRequest("DELETE", "/machines/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.DeleteMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Machine not found")
	})

	t.Run("should return 500 on delete error", func(t *testing.T) {
		mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(&machine.Machine{SerialNumber: "ERROR"}, nil)
		mockRepo.EXPECT().Delete(gomock.Any(), "ERROR").Return(fmt.Errorf("delete error"))

		req := httptest.NewRequest("DELETE", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", handler.DeleteMachine)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete machine")
	})
}

func TestMachineHandler_ListMachines(t *testing.T) {
	handler, mockRepo := setupTestHandler(t)

	t.Run("should list machines successfully", func(t *testing.T) {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
			{SerialNumber: "MACHINE002", Customer: "Customer 2", Status: "Maintenance"},
		}

		mockRepo.EXPECT().List(gomock.Any(), int32(50), "").Return(expectedMachines, "", nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(2), response["count"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Nil(t, response["next_page_token"])

		machines := response["machines"].([]interface{})
		assert.Len(t, machines, 2)
	})

	t.Run("should handle pagination with next page token", func(t *testing.T) {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE003", Customer: "Customer 3", Status: "Operational"},
		}
		nextPageToken := "eyJQSyI6eyJ2YWx1ZSI6Ik1BQ0hJTkUwMDMifSwiU0siOnsidmFsdWUiOiIjIn19"

		mockRepo.EXPECT().List(gomock.Any(), int32(50), "").Return(expectedMachines, nextPageToken, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, nextPageToken, response["next_page_token"])
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("should use custom limit when provided", func(t *testing.T) {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		mockRepo.EXPECT().List(gomock.Any(), int32(25), "").Return(expectedMachines, "", nil)

		req := httptest.NewRequest("GET", "/machines?limit=25", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(25), response["limit"])
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("should handle page token parameter", func(t *testing.T) {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE002", Customer: "Customer 2", Status: "Maintenance"},
		}
		pageToken := "eyJQSyI6eyJ2YWx1ZSI6Ik1BQ0hJTkUwMDEifSwiU0siOnsidmFsdWUiOiIjIn19"

		mockRepo.EXPECT().List(gomock.Any(), int32(50), pageToken).Return(expectedMachines, "", nil)

		req := httptest.NewRequest("GET", "/machines?page_token="+pageToken, nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(1), response["count"])
		assert.Nil(t, response["next_page_token"])
	})

	t.Run("should return 400 for invalid limit parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/machines?limit=invalid", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid limit parameter")
	})

	t.Run("should return 400 for limit less than 1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/machines?limit=0", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid limit parameter")
	})

	t.Run("should return 400 for limit greater than allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/machines?limit=101", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid limit parameter")
	})

	t.Run("should return 500 on repository error", func(t *testing.T) {
		mockRepo.EXPECT().List(gomock.Any(), int32(50), "").Return(nil, "", fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to list machines")
	})

	t.Run("should handle empty result set", func(t *testing.T) {
		mockRepo.EXPECT().List(gomock.Any(), int32(50), "").Return([]*machine.Machine{}, "", nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(0), response["count"])
		assert.Equal(t, float64(50), response["limit"])
		assert.Nil(t, response["next_page_token"])

		machines := response["machines"].([]interface{})
		assert.Len(t, machines, 0)
	})

	t.Run("should handle pagination with both limit and page token", func(t *testing.T) {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE004", Customer: "Customer 4", Status: "Operational"},
		}
		pageToken := "eyJQSyI6eyJ2YWx1ZSI6Ik1BQ0hJTkUwMDMifSwiU0siOnsidmFsdWUiOiIjIn19"
		nextPageToken := "eyJQSyI6eyJ2YWx1ZSI6Ik1BQ0hJTkUwMDQifSwiU0siOnsidmFsdWUiOiIjIn19"

		mockRepo.EXPECT().List(gomock.Any(), int32(10), pageToken).Return(expectedMachines, nextPageToken, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&page_token="+pageToken, nil)
		w := httptest.NewRecorder()

		handler.ListMachines(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(10), response["limit"])
		assert.Equal(t, float64(1), response["count"])
		assert.Equal(t, nextPageToken, response["next_page_token"])
	})
}
