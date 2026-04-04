package handlers_test

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
)

var errTest = errors.New("test error")

type CSVExportHandlerTestSuite struct {
	suite.Suite

	handler             *handlers.CSVExportHandler
	mockMachinesRepo    *machines.MockRepository
	mockMaintenanceRepo *maintenance.MockRepository
	mockAuditService    *audit.MockAuditService
	ctrl                *gomock.Controller
}

func (s *CSVExportHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockMachinesRepo = machines.NewMockRepository(s.ctrl)
	s.mockMaintenanceRepo = maintenance.NewMockRepository(s.ctrl)
	s.mockAuditService = audit.NewMockAuditService(s.ctrl)

	s.mockAuditService.EXPECT().LogEvent(gomock.Any()).AnyTimes()

	d := &deps.Dependencies{
		MachinesRepository:    s.mockMachinesRepo,
		MaintenanceRepository: s.mockMaintenanceRepo,
		AuditService:          s.mockAuditService,
		Config:                &deps.Config{},
	}
	s.handler = handlers.NewCSVExportHandler(d)
}

func (s *CSVExportHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestCSVExportHandlerSuite(t *testing.T) {
	suite.Run(t, new(CSVExportHandlerTestSuite))
}

func (s *CSVExportHandlerTestSuite) TestCSVInjectionSanitization() {
	s.Run("sanitizes formula-injection characters in machine fields", func() {
		injectionMachine := &machines.Machine{
			SerialNumber:    "=CMD()",
			Customer:        "+dangerous",
			State:           "-minus",
			District:        "@mention",
			Model:           "\ttab",
			AdditionalNotes: "\rcarriage",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		s.mockMachinesRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]*machines.Machine{injectionMachine}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/export/csv", nil)
		w := httptest.NewRecorder()
		s.handler.ExportMachinesCSV(w, req)

		s.Equal(http.StatusOK, w.Code)

		reader := csv.NewReader(strings.NewReader(w.Body.String()))
		records, err := reader.ReadAll()
		s.Require().NoError(err)
		s.Require().Len(records, 2)

		row := records[1]
		s.Equal("'=CMD()", row[0], "should prefix = with single quote")
		s.Equal("'+dangerous", row[1], "should prefix + with single quote")
		s.Equal("'-minus", row[2], "should prefix - with single quote")
		s.Equal("'@mention", row[3], "should prefix @ with single quote")
		s.True(strings.HasPrefix(row[4], "'"), "should prefix tab with single quote")
		s.True(strings.HasPrefix(row[13], "'"), "should prefix carriage return with single quote")
	})
}

func (s *CSVExportHandlerTestSuite) TestExportMachinesCSV() {
	s.Run("returns CSV with header and rows", func() {
		now := time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC)
		testMachines := []*machines.Machine{
			{
				SerialNumber:    "SN001",
				Customer:        "Acme Corp",
				State:           "Selangor",
				District:        "Petaling",
				Model:           "X100",
				Brand:           "BrandA",
				Status:          "Active",
				AccountType:     "Premium",
				PersonInCharge:  "Alice",
				ReportedBy:      "Bob",
				TncDate:         now,
				PpmDate:         now.AddDate(0, 1, 0),
				PpmStatus:       "",
				AdditionalNotes: "Some notes",
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			{
				SerialNumber: "SN002",
				Customer:     "Beta Inc",
				State:        "Johor",
				District:     "Muar",
				PpmStatus:    "overdue",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}

		s.mockMachinesRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(testMachines, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/export/csv", nil)
		w := httptest.NewRecorder()

		s.handler.ExportMachinesCSV(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Header().Get("Content-Type"), "text/csv")
		s.Contains(w.Header().Get("Content-Disposition"), "machines_")
		s.Contains(w.Header().Get("Content-Disposition"), ".csv")

		reader := csv.NewReader(strings.NewReader(w.Body.String()))
		records, err := reader.ReadAll()
		s.Require().NoError(err)

		s.Require().Len(records, 3) // header + 2 rows
		s.Equal("Serial Number", records[0][0])
		s.Equal("SN001", records[1][0])
		s.Equal("Acme Corp", records[1][1])
		s.Equal("SN002", records[2][0])
		s.Equal("overdue", records[2][12]) // PPM Status column
	})

	s.Run("uses search when query param provided", func() {
		s.mockMachinesRepo.EXPECT().
			Search(gomock.Any(), "test-query", gomock.Any()).
			Return([]*machines.Machine{}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/export/csv?q=test-query", nil)
		w := httptest.NewRecorder()

		s.handler.ExportMachinesCSV(w, req)

		s.Equal(http.StatusOK, w.Code)

		reader := csv.NewReader(strings.NewReader(w.Body.String()))
		records, err := reader.ReadAll()
		s.Require().NoError(err)
		s.Len(records, 1) // header only
	})

	s.Run("passes filters to list options", func() {
		s.mockMachinesRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ interface{}, opts *machines.ListOptions) ([]*machines.Machine, error) {
				s.Equal(machines.PPMStatusOverdue, opts.PpmStatusFilter)
				s.Equal(machines.SortOrderPpmDateAsc, opts.Sort)
				s.NotNil(opts.PpmDateFrom)
				return []*machines.Machine{}, nil
			})

		req := httptest.NewRequest("GET", "/api/v1/machines/export/csv?ppm_status_filter=overdue&sort=ppm_date_asc&ppm_date_from=2025-01-01", nil)
		w := httptest.NewRecorder()

		s.handler.ExportMachinesCSV(w, req)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run("returns 500 on repository error", func() {
		s.mockMachinesRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, errTest)

		req := httptest.NewRequest("GET", "/api/v1/machines/export/csv", nil)
		w := httptest.NewRecorder()

		s.handler.ExportMachinesCSV(w, req)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}

func (s *CSVExportHandlerTestSuite) TestExportMaintenanceCSV() {
	s.Run("returns CSV for a specific machine", func() {
		now := time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC)
		testRecords := []*maintenance.Maintenance{
			{
				WorkOrderNumber:     "WO-001",
				MachineSerialNumber: "SN001",
				WorkOrderDate:       now,
				WorkOrderType:       "Preventative",
				ActionTaken:         "Replaced filter",
				ReportedBy:          "Alice",
				CreatedAt:           now,
				UpdatedAt:           now,
			},
		}

		s.mockMaintenanceRepo.EXPECT().
			ListByMachine(gomock.Any(), "SN001", gomock.Any()).
			Return(testRecords, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/SN001/maintenance/export/csv", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/machines/{serial_number}/maintenance/export/csv", s.handler.ExportMaintenanceCSV)
		router.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Header().Get("Content-Type"), "text/csv")
		s.Contains(w.Header().Get("Content-Disposition"), "maintenance_SN001")

		reader := csv.NewReader(strings.NewReader(w.Body.String()))
		records, err := reader.ReadAll()
		s.Require().NoError(err)

		s.Require().Len(records, 2) // header + 1 row
		s.Equal("Work Order Number", records[0][0])
		s.Equal("WO-001", records[1][0])
		s.Equal("Preventative", records[1][3])
	})

	s.Run("uses search when query param provided", func() {
		s.mockMaintenanceRepo.EXPECT().
			SearchByMachine(gomock.Any(), "SN001", "filter-test", gomock.Any()).
			Return([]*maintenance.Maintenance{}, nil)

		req := httptest.NewRequest("GET", "/api/v1/machines/SN001/maintenance/export/csv?q=filter-test", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/machines/{serial_number}/maintenance/export/csv", s.handler.ExportMaintenanceCSV)
		router.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)
	})

	s.Run("returns 500 on repository error", func() {
		s.mockMaintenanceRepo.EXPECT().
			ListByMachine(gomock.Any(), "SN001", gomock.Any()).
			Return(nil, errTest)

		req := httptest.NewRequest("GET", "/api/v1/machines/SN001/maintenance/export/csv", nil)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/machines/{serial_number}/maintenance/export/csv", s.handler.ExportMaintenanceCSV)
		router.ServeHTTP(w, req)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}
