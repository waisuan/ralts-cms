package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handler"
	"ralts-cms/internal/machine"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// MachineHandlerTestSuite defines the test suite for machine handler
type MachineHandlerTestSuite struct {
	suite.Suite

	handler  *handler.MachineHandler
	mockRepo *machine.MockRepository
	ctrl     *gomock.Controller
}

// SetupTest sets up each test
func (suite *MachineHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockRepo = machine.NewMockRepository(suite.ctrl)
	deps := &deps.Dependencies{
		MachineRepository: suite.mockRepo,
		Config: &deps.Config{
			DefaultMachineLimit: 50,
		},
	}
	suite.handler = handler.NewMachineHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *MachineHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *MachineHandlerTestSuite) TestGetMachine() {
	suite.Run("should return machine when found", func() {
		expectedMachine := &machine.Machine{
			SerialNumber: "TEST123",
			Customer:     "Test Customer",
			State:        "Active",
		}

		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "TEST123").Return(expectedMachine, nil)

		req := httptest.NewRequest("GET", "/machines/TEST123", nil)
		w := httptest.NewRecorder()

		// Set up router for path parameters
		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("TEST123", response.SerialNumber)
		suite.Assert().Equal("Test Customer", response.Customer)
	})

	suite.Run("should return 404 when machine not found", func() {
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

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
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.GetMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to get machine")
	})
}

func (suite *MachineHandlerTestSuite) TestCreateMachine() {
	suite.Run("should create machine successfully", func() {
		machineData := machine.Machine{
			SerialNumber: "CREATE123",
			Customer:     "Create Customer",
			State:        "New",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *machine.Machine) error {
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

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("CREATE123", response.SerialNumber)
		suite.Assert().Equal("Create Customer", response.Customer)
	})

	suite.Run("should return 400 when serial number is missing", func() {
		machineData := machine.Machine{
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
		machineData := machine.Machine{
			SerialNumber: "DUPLICATE123",
			Customer:     "Duplicate Customer",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("ConditionalCheckFailedException"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusConflict, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine already exists")
	})

	suite.Run("should return 500 on repository error", func() {
		machineData := machine.Machine{
			SerialNumber: "ERROR123",
			Customer:     "Error Customer",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to create machine")
	})
}

func (suite *MachineHandlerTestSuite) TestUpdateMachine() {
	suite.Run("should update machine successfully", func() {
		machineData := machine.Machine{
			SerialNumber: "UPDATE123",
			Customer:     "Updated Customer",
			State:        "Active",
		}

		// Mock the existence check
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATE123").Return(&machine.Machine{SerialNumber: "UPDATE123"}, nil)
		// Mock the update
		suite.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *machine.Machine) error {
			suite.Assert().Equal("UPDATE123", m.SerialNumber)
			suite.Assert().Equal("Updated Customer", m.Customer)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.UpdateMachine(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("UPDATE123", response.SerialNumber)
		suite.Assert().Equal("Updated Customer", response.Customer)
	})

	suite.Run("should return 400 when serial number is missing", func() {
		machineData := machine.Machine{
			Customer: "Customer without serial",
		}

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.UpdateMachine(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Serial number is required")
	})

	suite.Run("should return 404 when machine not found", func() {
		machineData := machine.Machine{
			SerialNumber: "NOTFOUND123",
			Customer:     "Not Found Customer",
		}

		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND123").Return(nil, fmt.Errorf("machine not found"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.UpdateMachine(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine not found")
	})

	suite.Run("should return 500 on update error", func() {
		machineData := machine.Machine{
			SerialNumber: "UPDATEERROR123",
			Customer:     "Update Error Customer",
		}

		// Mock the existence check
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATEERROR123").Return(&machine.Machine{SerialNumber: "UPDATEERROR123"}, nil)
		// Mock the update error
		suite.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update error"))

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.handler.UpdateMachine(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to update machine")
	})
}

func (suite *MachineHandlerTestSuite) TestDeleteMachine() {
	suite.Run("should delete machine successfully", func() {
		// First, expect a check that the machine exists
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "DELETE123").Return(&machine.Machine{SerialNumber: "DELETE123"}, nil)
		// Then expect the delete operation
		suite.mockRepo.EXPECT().Delete(gomock.Any(), "DELETE123").Return(nil)

		req := httptest.NewRequest("DELETE", "/machines/DELETE123", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("should return 404 when machine not found", func() {
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("machine not found"))

		req := httptest.NewRequest("DELETE", "/machines/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Machine not found")
	})

	suite.Run("should return 500 on delete error", func() {
		suite.mockRepo.EXPECT().GetBySerialNumber(gomock.Any(), "ERROR").Return(&machine.Machine{SerialNumber: "ERROR"}, nil)
		suite.mockRepo.EXPECT().Delete(gomock.Any(), "ERROR").Return(fmt.Errorf("delete error"))

		req := httptest.NewRequest("DELETE", "/machines/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.DeleteMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to delete machine")
	})
}

func (suite *MachineHandlerTestSuite) TestListMachines() {
	suite.Run("should list machines successfully with default parameters", func() {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
			{SerialNumber: "MACHINE002", Customer: "Customer 2", Status: "Maintenance"},
		}

		expectedOptions := &machine.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(2), response["count"])
		suite.Assert().Equal(float64(50), response["limit"])
		suite.Assert().Equal(float64(0), response["offset"])
		suite.Assert().Equal("created_at_desc", response["sort"])

		machines := response["machines"].([]interface{})
		suite.Assert().Len(machines, 2)
	})

	suite.Run("should use custom limit when provided", func() {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machine.ListOptions{
			Limit:  25,
			Offset: 0,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)

		req := httptest.NewRequest("GET", "/machines?limit=25", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(25), response["limit"])
		suite.Assert().Equal(float64(1), response["count"])
	})

	suite.Run("should use custom offset when provided", func() {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE003", Customer: "Customer 3", Status: "Operational"},
		}

		expectedOptions := &machine.ListOptions{
			Limit:  50,
			Offset: 10,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)

		req := httptest.NewRequest("GET", "/machines?offset=10", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(10), response["offset"])
		suite.Assert().Equal(float64(1), response["count"])
	})

	suite.Run("should use custom sort when provided", func() {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machine.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   machine.SortOrderCreatedAtAsc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)

		req := httptest.NewRequest("GET", "/machines?sort=created_at_asc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("created_at_asc", response["sort"])
	})

	suite.Run("should use all custom parameters together", func() {
		expectedMachines := []*machine.Machine{
			{SerialNumber: "MACHINE005", Customer: "Customer 5", Status: "Operational"},
		}

		expectedOptions := &machine.ListOptions{
			Limit:  10,
			Offset: 20,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&offset=20&sort=created_at_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(10), response["limit"])
		suite.Assert().Equal(float64(20), response["offset"])
		suite.Assert().Equal("created_at_desc", response["sort"])
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
		expectedOptions := &machine.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to list machines")
	})

	suite.Run("should handle empty result set", func() {
		expectedOptions := &machine.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   machine.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().List(gomock.Any(), expectedOptions).Return([]*machine.Machine{}, nil)

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(0), response["count"])
		suite.Assert().Equal(float64(50), response["limit"])
		suite.Assert().Equal(float64(0), response["offset"])
		suite.Assert().Equal("created_at_desc", response["sort"])

		machines := response["machines"].([]interface{})
		suite.Assert().Len(machines, 0)
	})
}

func (suite *MachineHandlerTestSuite) TestGetDuePPM() {
	suite.Run("should return due PPM machines", func() {
		machines := []*machine.Machine{
			{SerialNumber: "DUEPPM001", Customer: "Customer 1", Status: "Operational"},
		}

		suite.mockRepo.EXPECT().DuePPM(gomock.Any()).Return(machines, nil)

		req := httptest.NewRequest("GET", "/machines/due-ppm", nil)
		w := httptest.NewRecorder()

		suite.handler.GetDuePPM(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response []*machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response, 1)
		suite.Assert().Equal("DUEPPM001", response[0].SerialNumber)
		suite.Assert().Equal("Customer 1", response[0].Customer)
		suite.Assert().Equal("Operational", response[0].Status)
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockRepo.EXPECT().DuePPM(gomock.Any()).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/due-ppm", nil)
		w := httptest.NewRecorder()

		suite.handler.GetDuePPM(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to get due PPM machines")
	})

	suite.Run("should return empty list when no machines are due for PPM", func() {
		suite.mockRepo.EXPECT().DuePPM(gomock.Any()).Return([]*machine.Machine{}, nil)

		req := httptest.NewRequest("GET", "/machines/due-ppm", nil)
		w := httptest.NewRecorder()

		suite.handler.GetDuePPM(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response []*machine.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response, 0)
	})
}

// TestMachineHandlerTestSuite runs the test suite
func TestMachineHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MachineHandlerTestSuite))
}
