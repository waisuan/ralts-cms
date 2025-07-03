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
	mockmaintenance "ralts-cms/internal/maintenance"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

// MaintenanceHandlerTestSuite defines the test suite for maintenance handler
type MaintenanceHandlerTestSuite struct {
	suite.Suite

	handler  *handler.MaintenanceHandler
	mockRepo *mockmaintenance.MockRepository
	ctrl     *gomock.Controller
}

// SetupTest sets up each test
func (suite *MaintenanceHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.mockRepo = mockmaintenance.NewMockRepository(suite.ctrl)
	deps := &deps.Dependencies{
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
	suite.Run("should return list of maintenance records", func() {
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

		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "MACHINE123").Return(expectedMaintenance, nil)

		req := httptest.NewRequest("GET", "/machines/MACHINE123/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
		suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

		var response []*maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Len(response, 2)
		suite.Assert().Equal("WO001", response[0].WorkOrderNumber)
		suite.Assert().Equal("WO002", response[1].WorkOrderNumber)
	})

	suite.Run("should return empty list when no maintenance found", func() {
		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "NOMAINTAINANCE").Return([]*maintenance.Maintenance{}, nil)

		req := httptest.NewRequest("GET", "/machines/NOMAINTAINANCE/maintenance", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}/maintenance", suite.handler.ListMaintenance)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response []*maintenance.Maintenance
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Empty(response)
	})

	suite.Run("should return 500 on repository error", func() {
		suite.mockRepo.EXPECT().ListByMachine(gomock.Any(), "ERROR").Return(nil, fmt.Errorf("database error"))

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
