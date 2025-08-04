package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// MachinesHandlerTestSuite defines the test suite for machine handler
type MachinesHandlerTestSuite struct {
	suite.Suite

	handler             *handlers.MachinesHandler
	mockMachinesRepo    *machines.MockRepository
	mockMaintenanceRepo *maintenance.MockRepository
	ctrl                *gomock.Controller
}

// SetupTest sets up each test
func (suite *MachinesHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockMachinesRepo = machines.NewMockRepository(suite.ctrl)
	suite.mockMaintenanceRepo = maintenance.NewMockRepository(suite.ctrl)

	deps := &deps.Dependencies{
		MachinesRepository:    suite.mockMachinesRepo,
		MaintenanceRepository: suite.mockMaintenanceRepo,
		Config: &deps.Config{
			DefaultMachinesLimit: 50,
			MaxMachinesLimit:     100,
		},
	}
	suite.handler = handlers.NewMachinesHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *MachinesHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *MachinesHandlerTestSuite) TestGetMachine() {
	suite.Run("should return machine when found", func() {
		expectedMachine := &machines.Machine{
			SerialNumber: "TEST123",
			Customer:     "Test Customer",
			State:        "Active",
		}

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "TEST123").Return(expectedMachine, nil)

		req := httptest.NewRequest("GET", "/machines/TEST123", nil)
		w := httptest.NewRecorder()

		// Set up router for path parameters
		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response machines.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("TEST123", response.SerialNumber)
		suite.Assert().Equal("Test Customer", response.Customer)
	})

	suite.Run("should return 404 when machine not found", func() {
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

		req := httptest.NewRequest("GET", "/machines/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine not found")
	})

	suite.Run("should return 400 when serial number is missing", func() {
		req := httptest.NewRequest("GET", "/machines/", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code) // Router returns 404 for missing path param
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to get machine")
	})
}

func (suite *MachinesHandlerTestSuite) TestCreateMachine() {
	suite.Run("should create machine successfully", func() {
		machineData := machines.Machine{
			SerialNumber: "CREATE123",
			Customer:     "Create Customer",
			State:        "New",
		}

		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *machines.Machine) error {
			suite.Assert().Equal("CREATE123", m.SerialNumber)
			suite.Assert().Equal("Create Customer", m.Customer)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusCreated, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response machines.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("CREATE123", response.SerialNumber)
		suite.Assert().Equal("Create Customer", response.Customer)
	})

	suite.Run("should return 400 when serial number is missing", func() {
		machineData := machines.Machine{
			Customer: "Customer without serial",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Serial number is required")
	})

	suite.Run("should return 400 when request body is invalid", func() {
		req := httptest.NewRequest("POST", "/machines", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})

	suite.Run("should return 409 when machine already exists", func() {
		machineData := machines.Machine{
			SerialNumber: "DUPLICATE123",
			Customer:     "Duplicate Customer",
		}

		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("ConditionalCheckFailedException"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusConflict, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine already exists")
	})

	suite.Run("should return 500 on repository error", func() {
		machineData := machines.Machine{
			SerialNumber: "ERROR123",
			Customer:     "Error Customer",
		}

		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to create machine")
	})
}

func (suite *MachinesHandlerTestSuite) TestUpdateMachine() {
	suite.Run("should update machine successfully", func() {
		machineData := machines.Machine{
			SerialNumber: "UPDATE123",
			Customer:     "Updated Customer",
			State:        "Active",
		}

		// Mock the existence check
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATE123").Return(&machines.Machine{SerialNumber: "UPDATE123"}, nil)
		// Mock the update
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *machines.Machine) error {
			suite.Assert().Equal("UPDATE123", m.SerialNumber)
			suite.Assert().Equal("Updated Customer", m.Customer)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/UPDATE123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Set up router for path parameters
		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response machines.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("UPDATE123", response.SerialNumber)
		suite.Assert().Equal("Updated Customer", response.Customer)
	})

	suite.Run("should return 400 when serial number in URL doesn't match request body", func() {
		machineData := machines.Machine{
			SerialNumber: "BODY123",
			Customer:     "Mismatch Customer",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/URL123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Serial number in request body does not match URL path")
	})

	suite.Run("should return 400 when serial number is missing from URL", func() {
		machineData := machines.Machine{
			SerialNumber: "TEST123",
			Customer:     "Test Customer",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code) // Router returns 404 for missing path param
	})

	suite.Run("should return 404 when machine not found", func() {
		machineData := machines.Machine{
			SerialNumber: "NOTFOUND123",
			Customer:     "Not Found Customer",
		}

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND123").Return(nil, fmt.Errorf("machine not found"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/NOTFOUND123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine not found")
	})

	suite.Run("should return 400 when request body is invalid", func() {
		req := httptest.NewRequest("PUT", "/machines/INVALID123", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})

	suite.Run("should return 500 on update error", func() {
		machineData := machines.Machine{
			SerialNumber: "UPDATEERROR123",
			Customer:     "Update Error Customer",
		}

		// Mock the existence check
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATEERROR123").Return(&machines.Machine{SerialNumber: "UPDATEERROR123"}, nil)
		// Mock the update error
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/UPDATEERROR123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to update machine")
	})

	suite.Run("should return 500 on repository error during existence check", func() {
		machineData := machines.Machine{
			SerialNumber: "ERROR123",
			Customer:     "Error Customer",
		}

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR123").Return(nil, fmt.Errorf("database error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/ERROR123", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to check machine existence")
	})
}

func (suite *MachinesHandlerTestSuite) TestDeleteMachine() {
	suite.Run("should delete machine successfully", func() {
		// First, expect a check that the machine exists
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "DELETE123").Return(&machines.Machine{SerialNumber: "DELETE123"}, nil)
		// Then expect the delete operation
		suite.mockMachinesRepo.EXPECT().Delete(gomock.Any(), "DELETE123").Return(nil)

		req := httptest.NewRequest("DELETE", "/machines/DELETE123", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("should return 404 when machine not found", func() {
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

		req := httptest.NewRequest("DELETE", "/machines/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine not found")
	})

	suite.Run("should return 500 on delete error", func() {
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(&machines.Machine{SerialNumber: "ERROR"}, nil)
		suite.mockMachinesRepo.EXPECT().Delete(gomock.Any(), "ERROR").Return(fmt.Errorf("delete error"))

		req := httptest.NewRequest("DELETE", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to delete machine")
	})
}

func (suite *MachinesHandlerTestSuite) TestListMachines() {
	suite.Run("should list machines successfully with default parameters", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
			{SerialNumber: "MACHINE002", Customer: "Customer 2", Status: "Maintenance"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(1), int32(2), int32(3), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE001").Return(1, nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE002").Return(2, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(2), response.Count)
		suite.Assert().Equal(int32(50), response.Limit)
		suite.Assert().Equal(int32(0), response.Offset)
		suite.Assert().Equal("created_at_desc", response.Sort)
		suite.Assert().Equal(int32(1), response.OverdueCount)
		suite.Assert().Equal(int32(2), response.DueCount)
		suite.Assert().Equal(int32(3), response.AlmostDueCount)

		suite.Assert().Len(response.Machines, 2)
		suite.Assert().Equal(int(1), response.Machines[0].MaintenanceCount)
		suite.Assert().Equal(int(2), response.Machines[1].MaintenanceCount)
	})

	suite.Run("should use custom limit when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           25,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?limit=25", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(25), response.Limit)
		suite.Assert().Equal(int32(1), response.Count)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use custom offset when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE003", Customer: "Customer 3", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          10,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE003").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?offset=10", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(10), response.Offset)
		suite.Assert().Equal(int32(1), response.Count)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use custom sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtAsc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?sort=created_at_asc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("created_at_asc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use all custom parameters together", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE005", Customer: "Customer 5", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           10,
			Offset:          20,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE005").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&offset=20&sort=created_at_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(20), response.Offset)
		suite.Assert().Equal("created_at_desc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should include PPM counts in response", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(5), int32(3), int32(2), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(5), response.OverdueCount)
		suite.Assert().Equal(int32(3), response.DueCount)
		suite.Assert().Equal(int32(2), response.AlmostDueCount)
	})

	suite.Run("should return 400 for invalid limit parameter", func() {
		req := httptest.NewRequest("GET", "/machines?limit=invalid", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for limit less than 1", func() {
		req := httptest.NewRequest("GET", "/machines?limit=0", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for limit greater than allowed", func() {
		req := httptest.NewRequest("GET", "/machines?limit=101", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for invalid offset parameter", func() {
		req := httptest.NewRequest("GET", "/machines?offset=invalid", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid offset parameter")
	})

	suite.Run("should return 400 for negative offset parameter", func() {
		req := httptest.NewRequest("GET", "/machines?offset=-1", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid offset parameter")
	})

	suite.Run("should return 400 for invalid sort parameter", func() {
		req := httptest.NewRequest("GET", "/machines?sort=invalid_sort", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid sort parameter")
	})

	suite.Run("should return 500 on repository error", func() {
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to list machines")
	})

	suite.Run("should handle empty result set", func() {
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(0, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "MACHINE001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(0), response.Count)
		suite.Assert().Equal(int32(50), response.Limit)
		suite.Assert().Equal(int32(0), response.Offset)
		suite.Assert().Equal("created_at_desc", response.Sort)

		suite.Assert().Len(response.Machines, 0)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should return 500 on repository error when counting machines", func() {
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(0, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to count machines")
	})

	suite.Run("should return 500 on repository error when counting by status", func() {
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(0, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to count machines by status")
	})

	suite.Run("should use PpmStatusFilter=overdue when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "OVERDUE001", Customer: "Customer 1", Status: "Overdue"},
			{SerialNumber: "OVERDUE002", Customer: "Customer 2", Status: "Overdue"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(2), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "OVERDUE001").Return(1, nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "OVERDUE002").Return(2, nil)

		req := httptest.NewRequest("GET", "/machines?ppm_status_filter=overdue", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 2)
		suite.Assert().Equal(int32(2), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use PpmStatusFilter=due when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "DUE001", Customer: "Customer 1", Status: "Due"},
			{SerialNumber: "DUE002", Customer: "Customer 2", Status: "Due"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(2), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "DUE001").Return(1, nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "DUE002").Return(2, nil)

		req := httptest.NewRequest("GET", "/machines?ppm_status_filter=due", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 2)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(2), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use PpmStatusFilter=almost_due when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "ALMOSTDUE001", Customer: "Customer 1", Status: "Almost Due"},
			{SerialNumber: "ALMOSTDUE002", Customer: "Customer 2", Status: "Almost Due"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusAlmostDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(2), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "ALMOSTDUE001").Return(1, nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "ALMOSTDUE002").Return(2, nil)

		req := httptest.NewRequest("GET", "/machines?ppm_status_filter=almost_due", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 2)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(2), response.AlmostDueCount)
	})

	suite.Run("should combine PpmStatusFilter with other parameters", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "FILTERED001", Customer: "Customer 1", Status: "Due"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           10,
			Offset:          5,
			Sort:            machines.SortOrderCreatedAtAsc,
			PpmStatusFilter: machines.PPMStatusDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(1), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "FILTERED001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&offset=5&sort=created_at_asc&ppm_status_filter=due", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(5), response.Offset)
		suite.Assert().Equal("created_at_asc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(1), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should return 400 for invalid ppm_status_filter parameter", func() {
		req := httptest.NewRequest("GET", "/machines?ppm_status_filter=invalid", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid ppm_status_filter parameter")
	})

	suite.Run("should default to empty string when ppm_status_filter is not provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should call Search when query parameter is provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "SEARCH001", Customer: "HP Customer", Brand: "HP"},
			{SerialNumber: "SEARCH002", Customer: "HP Customer", Brand: "HP"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect Search to be called instead of List
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "HP", expectedOptions).Return(expectedMachines, nil)
		// Expect CountSearch to be called for search results
		suite.mockMachinesRepo.EXPECT().CountSearch(gomock.Any(), "HP", expectedOptions).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "SEARCH001").Return(1, nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "SEARCH002").Return(2, nil)

		req := httptest.NewRequest("GET", "/machines?q=HP", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 2)
		suite.Assert().Equal("SEARCH001", response.Machines[0].SerialNumber)
		suite.Assert().Equal("SEARCH002", response.Machines[1].SerialNumber)
		suite.Assert().Equal(int32(2), response.Count) // Verify search count is used
	})

	suite.Run("should call List when query parameter is empty", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "LIST001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect List to be called when q parameter is empty
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "LIST001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?q=", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal("LIST001", response.Machines[0].SerialNumber)
	})

	suite.Run("should combine search with other parameters", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "COMBINED001", Customer: "HP Customer", Brand: "HP"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           10,
			Offset:          5,
			Sort:            machines.SortOrderCreatedAtAsc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		// Expect Search to be called with combined parameters
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "printer", expectedOptions).Return(expectedMachines, nil)
		// Expect CountSearch to be called for search results
		suite.mockMachinesRepo.EXPECT().CountSearch(gomock.Any(), "printer", expectedOptions).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(1), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachine(gomock.Any(), "COMBINED001").Return(1, nil)

		req := httptest.NewRequest("GET", "/machines?q=printer&limit=10&offset=5&sort=created_at_asc&ppm_status_filter=overdue", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(5), response.Offset)
		suite.Assert().Equal("created_at_asc", response.Sort)
		suite.Assert().Equal(int32(1), response.OverdueCount)
		suite.Assert().Equal(int32(1), response.Count) // Verify search count is used
	})

	suite.Run("should return error when Search fails", func() {
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect Search to be called and return error
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "invalid", expectedOptions).Return(nil, fmt.Errorf("search failed"))

		req := httptest.NewRequest("GET", "/machines?q=invalid", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to search machines")
	})

	suite.Run("should return error when CountSearch fails", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "SEARCH001", Customer: "HP Customer", Brand: "HP"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect Search to succeed but CountSearch to fail
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "HP", expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().CountSearch(gomock.Any(), "HP", expectedOptions).Return(0, fmt.Errorf("count search failed"))

		req := httptest.NewRequest("GET", "/machines?q=HP", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to count search results")
	})
}

// TestMachinesHandlerTestSuite runs the test suite
func TestMachinesHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MachinesHandlerTestSuite))
}
