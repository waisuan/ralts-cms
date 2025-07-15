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
	"ralts-cms/internal/maintenance"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// MaintenanceHandlerTestSuite defines the test suite for maintenance handler
type MaintenanceHandlerTestSuite struct {
	suite.Suite

	handler  *handler.MaintenanceHandler
	mockRepo *maintenance.MockRepository
	ctrl     *gomock.Controller
}

// SetupTest sets up each test
func (suite *MaintenanceHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.mockRepo = maintenance.NewMockRepository(suite.ctrl)
	deps := &deps.Dependencies{
		Config: &deps.Config{
			DefaultMaintenanceLimit: 50,
			MaxMaintenanceLimit:     100,
		},
		MaintenanceRepository: suite.mockRepo,
	}
	suite.handler = handler.NewMaintenanceHandler(deps)
}

// TearDownTest cleans up after each test
func (suite *MaintenanceHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *MaintenanceHandlerTestSuite) TestGetMaintenance() {
	suite.Run("should return maintenance when found", func() {
		expectedMaintenance := &maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Routine maintenance",
			ReportedBy:          "John Doe",
		}

		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.GetMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE123", response.MachineSerialNumber)
		suite.Assert().Equal("WO001", response.WorkOrderNumber)
		suite.Assert().Equal("Routine maintenance", response.ActionTaken)
	})

	suite.Run("should return 404 when maintenance not found", func() {
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.GetMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Maintenance not found")
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "ERROR").Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance/ERROR", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.GetMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to get maintenance")
	})
}

func (suite *MaintenanceHandlerTestSuite) TestListMaintenance() {
	suite.Run("should return list of maintenance records with default parameters", func() {
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

		expectedOptions := &maintenance.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123", expectedOptions).Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(2), response["count"])
		suite.Assert().Equal(float64(50), response["limit"])
		suite.Assert().Equal(float64(0), response["offset"])
		suite.Assert().Equal("work_order_date_desc", response["sort"])

		maintenanceList := response["maintenance"].([]interface{})
		suite.Assert().Len(maintenanceList, 2)
	})

	suite.Run("should use custom limit when provided", func() {
		expectedMaintenance := []*maintenance.Maintenance{
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO001",
				ActionTaken:         "First maintenance",
				ReportedBy:          "Tech1",
			},
		}

		expectedOptions := &maintenance.ListOptions{
			Limit:  25,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123", expectedOptions).Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?limit=25", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(25), response["limit"])
		suite.Assert().Equal(float64(1), response["count"])
	})

	suite.Run("should use custom offset when provided", func() {
		expectedMaintenance := []*maintenance.Maintenance{
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO003",
				ActionTaken:         "Third maintenance",
				ReportedBy:          "Tech3",
			},
		}

		expectedOptions := &maintenance.ListOptions{
			Limit:  50,
			Offset: 10,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123", expectedOptions).Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?offset=10", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(10), response["offset"])
		suite.Assert().Equal(float64(1), response["count"])
	})

	suite.Run("should use custom sort when provided", func() {
		expectedMaintenance := []*maintenance.Maintenance{
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO001",
				ActionTaken:         "First maintenance",
				ReportedBy:          "Tech1",
			},
		}

		expectedOptions := &maintenance.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123", expectedOptions).Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?sort=created_at_desc", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("created_at_desc", response["sort"])
	})

	suite.Run("should use all custom parameters together", func() {
		expectedMaintenance := []*maintenance.Maintenance{
			{
				MachineSerialNumber: "MACHINE123",
				WorkOrderNumber:     "WO005",
				ActionTaken:         "Fifth maintenance",
				ReportedBy:          "Tech5",
			},
		}

		expectedOptions := &maintenance.ListOptions{
			Limit:  10,
			Offset: 20,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123", expectedOptions).Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?limit=10&offset=20&sort=work_order_date_desc", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(10), response["limit"])
		suite.Assert().Equal(float64(20), response["offset"])
		suite.Assert().Equal("work_order_date_desc", response["sort"])
	})

	suite.Run("should return 400 for invalid limit parameter", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?limit=invalid", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for limit less than 1", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?limit=0", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for limit greater than allowed", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?limit=101", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid limit parameter")
	})

	suite.Run("should return 400 for invalid offset parameter", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?offset=invalid", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid offset parameter")
	})

	suite.Run("should return 400 for negative offset parameter", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?offset=-1", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid offset parameter")
	})

	suite.Run("should return 400 for invalid sort parameter", func() {
		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance?sort=invalid_sort", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid sort parameter")
	})

	suite.Run("should return empty list when no maintenance found", func() {
		expectedOptions := &maintenance.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "NOMAINTAINANCE", expectedOptions).Return([]*maintenance.Maintenance{}, nil)

		req := httptest.NewRequest("GET", "/machines/NOMAINTAINANCE/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(float64(0), response["count"])
		suite.Assert().Equal(float64(50), response["limit"])
		suite.Assert().Equal(float64(0), response["offset"])
		suite.Assert().Equal("work_order_date_desc", response["sort"])

		maintenanceList := response["maintenance"].([]interface{})
		suite.Assert().Len(maintenanceList, 0)
	})

	suite.Run("should return 500 on repository error", func() {
		expectedOptions := &maintenance.ListOptions{
			Limit:  50,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		}

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "ERROR", expectedOptions).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines/ERROR/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to list maintenance")
	})
}

func (suite *MaintenanceHandlerTestSuite) TestCreateMaintenance() {
	suite.Run("should create maintenance successfully", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Routine maintenance",
			ReportedBy:          "John Doe",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *maintenance.Maintenance) error {
			suite.Assert().Equal("MACHINE123", m.MachineSerialNumber)
			suite.Assert().Equal("WO001", m.WorkOrderNumber)
			suite.Assert().Equal("Routine maintenance", m.ActionTaken)
			return nil
		})

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusCreated, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE123", response.MachineSerialNumber)
		suite.Assert().Equal("WO001", response.WorkOrderNumber)
		suite.Assert().Equal("Routine maintenance", response.ActionTaken)
	})

	suite.Run("should return 400 when work order number is missing", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			ActionTaken:         "Maintenance without work order",
		}

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Work order number is required")
	})

	suite.Run("should return 400 when request body is invalid", func() {
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid request body")
	})

	suite.Run("should return 409 when maintenance already exists", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Duplicate maintenance",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("ConditionalCheckFailedException"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusConflict, w.Code)
		suite.Assert().Contains(w.Body.String(), "Maintenance already exists")
	})

	suite.Run("should return 500 on repository error", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Error maintenance",
		}

		suite.mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("database error"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("POST", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.CreateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to create maintenance")
	})
}

func (suite *MaintenanceHandlerTestSuite) TestUpdateMaintenance() {
	suite.Run("should update maintenance successfully", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Updated maintenance",
			ReportedBy:          "Updated Tech",
		}

		// Mock the existence check
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the update
		suite.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *maintenance.Maintenance) error {
			suite.Assert().Equal("MACHINE123", m.MachineSerialNumber)
			suite.Assert().Equal("WO001", m.WorkOrderNumber)
			suite.Assert().Equal("Updated maintenance", m.ActionTaken)
			return nil
		})

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE123", response.MachineSerialNumber)
		suite.Assert().Equal("WO001", response.WorkOrderNumber)
		suite.Assert().Equal("Updated maintenance", response.ActionTaken)
	})

	suite.Run("should return 400 when work order number is missing", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			ActionTaken:         "Maintenance without work order",
		}

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Work order number is required")
	})

	suite.Run("should return 404 when maintenance not found", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "NOTFOUND",
			ActionTaken:         "Not found maintenance",
		}

		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Maintenance not found")
	})

	suite.Run("should return 500 on update error", func() {
		maintenanceData := maintenance.Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Update error maintenance",
		}

		// Mock the existence check
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the update error
		suite.mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(fmt.Errorf("update error"))

		body, _ := json.Marshal(maintenanceData)
		req := httptest.NewRequest("PUT", "/machines/MACHINE123/maintenance", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.UpdateMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to update maintenance")
	})
}

func (suite *MaintenanceHandlerTestSuite) TestDeleteMaintenance() {
	suite.Run("should delete maintenance successfully", func() {
		// Mock the existence check
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the delete
		suite.mockRepo.EXPECT().Delete(gomock.Any(), "MACHINE123", "WO001").Return(nil)

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNoContent, w.Code)
	})

	suite.Run("should return 404 when maintenance not found", func() {
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "NOTFOUND").Return(nil, fmt.Errorf("maintenance not found"))

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/NOTFOUND", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusNotFound, w.Code)
		suite.Assert().Contains(w.Body.String(), "Maintenance not found")
	})

	suite.Run("should return 500 on delete error", func() {
		// Mock the existence check
		suite.mockRepo.EXPECT().GetByWorkOrder(gomock.Any(), "MACHINE123", "WO001").Return(&maintenance.Maintenance{WorkOrderNumber: "WO001"}, nil)
		// Mock the delete error
		suite.mockRepo.EXPECT().Delete(gomock.Any(), "MACHINE123", "WO001").Return(fmt.Errorf("delete error"))

		req := httptest.NewRequest("DELETE", "/machines/MACHINE123/maintenance/WO001", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", suite.handler.DeleteMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to delete maintenance")
	})
}

// TestMaintenanceHandlerTestSuite runs the test suite
func TestMaintenanceHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceHandlerTestSuite))
}
