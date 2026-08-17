package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/audit"
	appctx "ralts-cms/internal/context"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/users"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/suite"
)

// MachinesHandlerTestSuite defines the test suite for machine handler
type MachinesHandlerTestSuite struct {
	suite.Suite

	handler             *handlers.MachinesHandler
	mockMachinesRepo    *machines.MockRepository
	mockMaintenanceRepo *maintenance.MockRepository
	mockUsersRepo       *users.MockRepository
	mockAuditService    *audit.MockAuditService
	ctrl                *gomock.Controller
}

// SetupTest sets up each test
func (suite *MachinesHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockMachinesRepo = machines.NewMockRepository(suite.ctrl)
	suite.mockMaintenanceRepo = maintenance.NewMockRepository(suite.ctrl)
	suite.mockUsersRepo = users.NewMockRepository(suite.ctrl)
	suite.mockAuditService = audit.NewMockAuditService(suite.ctrl)

	// Allow any audit events to be logged
	suite.mockAuditService.EXPECT().LogEvent(gomock.Any()).AnyTimes()

	deps := &deps.Dependencies{
		MachinesRepository:    suite.mockMachinesRepo,
		MaintenanceRepository: suite.mockMaintenanceRepo,
		UsersRepository:       suite.mockUsersRepo,
		AuditService:          suite.mockAuditService,
		Config: &deps.Config{
			DefaultMachinesLimit: 50,
			MaxMachinesLimit:     100,
		},
	}
	suite.handler = handlers.NewMachinesHandler(deps)
}

// withAuthenticatedUser attaches an authenticated UserContext to the request.
func withAuthenticatedUser(req *http.Request, userID int64) *http.Request {
	userCtx := &appctx.UserContext{UserID: userID, EntityID: fmt.Sprintf("%d", userID), Role: users.RoleNonAdmin}
	return req.WithContext(appctx.WithUser(req.Context(), userCtx))
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
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("%w", machines.ErrNotFound))

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
			// No authenticated user in context, so updated_by should be left blank.
			suite.Assert().Empty(m.UpdatedBy)
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

	suite.Run("should stamp updated_by from the authenticated user, ignoring client-supplied value", func() {
		machineData := machines.Machine{
			SerialNumber: "CREATE124",
			Customer:     "Create Customer",
			UpdatedBy:    "spoofed.user",
		}

		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), int64(42)).Return(&users.User{ID: 42, Username: "real.user"}, nil)
		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *machines.Machine) error {
			suite.Assert().Equal("real.user", m.UpdatedBy)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAuthenticatedUser(req, 42)
		w := httptest.NewRecorder()

		suite.handler.CreateMachine(w, req)

		suite.Assert().Equal(http.StatusCreated, w.Code)

		var response machines.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("real.user", response.UpdatedBy)
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

		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&pgconn.PgError{Code: "23505"})

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

// assignableUser builds a user the assignee dropdown would offer: approved and
// active. Only such an account may be named as a new assignee.
func assignableUser(id int64, username string) *users.User {
	status := users.StatusApproved
	return &users.User{
		ID:       id,
		Username: username,
		Email:    username + "@example.com",
		Approved: true,
		Status:   &status,
	}
}

// TestCreateMachineAssignee covers the two ways a machine can be assigned: to a
// registered user via assigned_user_id, or to a free-text name for someone who
// has no account yet.
func (suite *MachinesHandlerTestSuite) TestCreateMachineAssignee() {
	createMachine := func(m machines.Machine) *httptest.ResponseRecorder {
		body, _ := json.Marshal(m)
		req := httptest.NewRequest("POST", "/machines", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.handler.CreateMachine(w, req)
		return w
	}

	suite.Run("derives personInCharge from the assigned user when assigned_user_id is set", func() {
		assigneeID := int64(11)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), assigneeID).
			Return(assignableUser(assigneeID, "tech.one"), nil)
		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, m *machines.Machine) error {
				suite.Assert().Equal("tech.one", m.PersonInCharge, "client-supplied text must be overwritten")
				suite.Require().NotNil(m.AssignedUserID)
				suite.Assert().Equal(assigneeID, *m.AssignedUserID)
				return nil
			})

		w := createMachine(machines.Machine{
			SerialNumber:   "ASSIGN-FK",
			AssignedUserID: &assigneeID,
			PersonInCharge: "ignored text",
		})
		suite.Require().Equal(http.StatusCreated, w.Code)

		var response machines.Machine
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &response))
		suite.Require().NotNil(response.AssignedUser)
		suite.Assert().Equal("tech.one", response.AssignedUser.Username)
	})

	suite.Run("keeps free-text personInCharge when no assigned_user_id is given", func() {
		suite.mockMachinesRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, m *machines.Machine) error {
				suite.Assert().Equal("New Starter", m.PersonInCharge)
				suite.Assert().Nil(m.AssignedUserID)
				return nil
			})

		w := createMachine(machines.Machine{
			SerialNumber:   "ASSIGN-GHOST",
			PersonInCharge: "  New Starter  ",
		})
		suite.Require().Equal(http.StatusCreated, w.Code)

		var response machines.Machine
		suite.Require().NoError(json.Unmarshal(w.Body.Bytes(), &response))
		suite.Assert().Equal("New Starter", response.PersonInCharge)
		suite.Assert().Nil(response.AssignedUser)
	})

	suite.Run("returns 400 when assigned_user_id does not resolve to a user", func() {
		missingID := int64(999)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), missingID).
			Return(nil, fmt.Errorf("not found"))

		w := createMachine(machines.Machine{
			SerialNumber:   "ASSIGN-BAD",
			AssignedUserID: &missingID,
		})
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "does not resolve to a user")
	})

	suite.Run("returns 400 when free-text assignee exceeds the column width", func() {
		w := createMachine(machines.Machine{
			SerialNumber:   "ASSIGN-LONG",
			PersonInCharge: strings.Repeat("a", 201),
		})
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "at most 200 characters")
	})

	// Accounts the dropdown does not offer cannot be assigned through the API
	// either: they would be sent notifications they cannot come back and act on.
	for _, tc := range []struct {
		name   string
		id     int64
		mutate func(*users.User)
		serial string
	}{
		{
			name:   "awaiting approval",
			id:     12,
			serial: "ASSIGN-PENDING",
			mutate: func(u *users.User) {
				u.Approved = false
				status := users.StatusPendingApproval
				u.Status = &status
			},
		},
		{
			name:   "suspended",
			id:     13,
			serial: "ASSIGN-SUSPENDED",
			mutate: func(u *users.User) {
				status := users.StatusSuspended
				u.Status = &status
			},
		},
	} {
		suite.Run("returns 400 when the assignee is "+tc.name, func() {
			user := assignableUser(tc.id, "ineligible.user")
			tc.mutate(user)
			suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), tc.id).Return(user, nil)

			id := tc.id
			w := createMachine(machines.Machine{SerialNumber: tc.serial, AssignedUserID: &id})
			suite.Assert().Equal(http.StatusBadRequest, w.Code)
			suite.Assert().Contains(w.Body.String(), "not an approved, active account")
		})
	}
}

// TestUpdateMachineAssignee covers what an update does to an existing
// assignment: a body that names no assignee leaves it alone, an explicit null
// clears it, and an account that has since stopped being assignable can still be
// saved as long as the request is not moving the machine to them.
func (suite *MachinesHandlerTestSuite) TestUpdateMachineAssignee() {
	updateMachine := func(serial string, body map[string]any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/machines/"+serial, bytes.NewBuffer(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)
		return w
	}

	suite.Run("a body that names no assignee keeps the current one", func() {
		assigneeID := int64(21)
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "KEEP-1").
			Return(&machines.Machine{SerialNumber: "KEEP-1", AssignedUserID: &assigneeID, PersonInCharge: "tech.one"}, nil)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), assigneeID).
			Return(assignableUser(assigneeID, "tech.one"), nil)
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, m *machines.Machine) error {
				suite.Require().NotNil(m.AssignedUserID, "an unmentioned assignee must survive the update")
				suite.Assert().Equal(assigneeID, *m.AssignedUserID)
				suite.Assert().Equal("tech.one", m.PersonInCharge)
				return nil
			})

		w := updateMachine("KEEP-1", map[string]any{
			"serial_number": "KEEP-1",
			"customer":      "Only the customer changed",
		})
		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("a null assigned_user_id hands the machine to a free-text name", func() {
		assigneeID := int64(22)
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "CLEAR-1").
			Return(&machines.Machine{SerialNumber: "CLEAR-1", AssignedUserID: &assigneeID, PersonInCharge: "tech.one"}, nil)
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, m *machines.Machine) error {
				suite.Assert().Nil(m.AssignedUserID)
				suite.Assert().Equal("Ghost Contractor", m.PersonInCharge)
				return nil
			})

		w := updateMachine("CLEAR-1", map[string]any{
			"serial_number":    "CLEAR-1",
			"assigned_user_id": nil,
			"person_in_charge": "Ghost Contractor",
		})
		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("an assignee suspended since being assigned is still saveable", func() {
		assigneeID := int64(23)
		suspended := assignableUser(assigneeID, "suspended.user")
		status := users.StatusSuspended
		suspended.Status = &status

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "SUSPENDED-1").
			Return(&machines.Machine{SerialNumber: "SUSPENDED-1", AssignedUserID: &assigneeID}, nil)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), assigneeID).Return(suspended, nil)
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		w := updateMachine("SUSPENDED-1", map[string]any{
			"serial_number":    "SUSPENDED-1",
			"assigned_user_id": assigneeID,
		})
		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("moving a machine to a suspended account is rejected", func() {
		previousID := int64(24)
		suspendedID := int64(25)
		suspended := assignableUser(suspendedID, "suspended.user")
		status := users.StatusSuspended
		suspended.Status = &status

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "SUSPENDED-2").
			Return(&machines.Machine{SerialNumber: "SUSPENDED-2", AssignedUserID: &previousID}, nil)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), suspendedID).Return(suspended, nil)

		w := updateMachine("SUSPENDED-2", map[string]any{
			"serial_number":    "SUSPENDED-2",
			"assigned_user_id": suspendedID,
		})
		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "not an approved, active account")
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

	suite.Run("should stamp updated_by from the authenticated user, ignoring client-supplied value", func() {
		machineData := machines.Machine{
			SerialNumber: "UPDATE124",
			Customer:     "Updated Customer",
			UpdatedBy:    "spoofed.user",
		}

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "UPDATE124").Return(&machines.Machine{SerialNumber: "UPDATE124"}, nil)
		suite.mockUsersRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(&users.User{ID: 7, Username: "real.editor"}, nil)
		suite.mockMachinesRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, m *machines.Machine) error {
			suite.Assert().Equal("real.editor", m.UpdatedBy)
			return nil
		})

		body, _ := json.Marshal(machineData)
		req := httptest.NewRequest("PUT", "/machines/UPDATE124", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAuthenticatedUser(req, 7)
		w := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/machines/{serial_number}", suite.handler.UpdateMachine)
		router.ServeHTTP(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response machines.Machine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)
		suite.Assert().Equal("real.editor", response.UpdatedBy)
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

		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND123").Return(nil, fmt.Errorf("%w", machines.ErrNotFound))

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
		suite.mockMachinesRepo.EXPECT().GetBySerialNumber(gomock.Any(), "NOTFOUND").Return(nil, fmt.Errorf("%w", machines.ErrNotFound))

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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(1), int32(2), int32(3), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001", "MACHINE002"}).Return(map[string]int{"MACHINE001": 1, "MACHINE002": 2}, nil)

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
		suite.Assert().Equal("updated_at_desc", response.Sort)
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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE003"}).Return(map[string]int{"MACHINE003": 1}, nil)

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

	suite.Run("should use all custom parameters together", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE005", Customer: "Customer 5", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           10,
			Offset:          20,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE005"}).Return(map[string]int{"MACHINE005": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&offset=20&sort=updated_at_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(20), response.Offset)
		suite.Assert().Equal("updated_at_desc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use updated_at_desc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=updated_at_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("updated_at_desc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use updated_at_asc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtAsc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=updated_at_asc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("updated_at_asc", response.Sort)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should use ppm_date_asc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderPpmDateAsc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=ppm_date_asc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("ppm_date_asc", response.Sort)
	})

	suite.Run("should use ppm_date_desc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderPpmDateDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=ppm_date_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("ppm_date_desc", response.Sort)
	})

	suite.Run("should use tnc_date_asc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderTncDateAsc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=tnc_date_asc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("tnc_date_asc", response.Sort)
	})

	suite.Run("should use tnc_date_desc sort when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderTncDateDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?sort=tnc_date_desc", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Equal("tnc_date_desc", response.Sort)
	})

	suite.Run("should filter by ppm_date_from when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		ppmDateFrom := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
			PpmDateFrom:     &ppmDateFrom,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?ppm_date_from=2025-01-15", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should filter by ppm_date range when both provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		ppmDateFrom := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		ppmDateTo := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
			PpmDateFrom:     &ppmDateFrom,
			PpmDateTo:       &ppmDateTo,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?ppm_date_from=2025-01-01&ppm_date_to=2025-01-31", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should filter by tnc_date range when provided", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		tncDateFrom := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		tncDateTo := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
			TncDateFrom:     &tncDateFrom,
			TncDateTo:       &tncDateTo,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?tnc_date_from=2025-02-01&tnc_date_to=2025-02-28", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)
	})

	suite.Run("should return 400 for invalid ppm_date_from format", func() {
		req := httptest.NewRequest("GET", "/machines?ppm_date_from=invalid-date", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid ppm_date_from format")
	})

	suite.Run("should return 400 for invalid ppm_date_to format", func() {
		req := httptest.NewRequest("GET", "/machines?ppm_date_to=2025/01/15", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid ppm_date_to format")
	})

	suite.Run("should return 400 for invalid tnc_date_from format", func() {
		req := httptest.NewRequest("GET", "/machines?tnc_date_from=01-15-2025", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid tnc_date_from format")
	})

	suite.Run("should return 400 for invalid tnc_date_to format", func() {
		req := httptest.NewRequest("GET", "/machines?tnc_date_to=not-a-date", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusBadRequest, w.Code)
		suite.Assert().Contains(w.Body.String(), "Invalid tnc_date_to format")
	})

	suite.Run("should include PPM counts in response", func() {
		expectedMachines := []*machines.Machine{
			{SerialNumber: "MACHINE001", Customer: "Customer 1", Status: "Operational"},
		}

		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(5), int32(3), int32(2), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

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
			Sort:            machines.SortOrderUpdatedAtDesc,
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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(0, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"MACHINE001"}).Return(map[string]int{"MACHINE001": 1}, nil)

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
		suite.Assert().Equal("updated_at_desc", response.Sort)

		suite.Assert().Len(response.Machines, 0)
		suite.Assert().Equal(int32(0), response.OverdueCount)
		suite.Assert().Equal(int32(0), response.DueCount)
		suite.Assert().Equal(int32(0), response.AlmostDueCount)
	})

	suite.Run("should return 500 on repository error when counting machines", func() {
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(0, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/machines", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusInternalServerError, w.Code)
		suite.Assert().Contains(w.Body.String(), "Failed to count machines")
	})

	suite.Run("should return 500 on repository error when counting by status", func() {
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*machines.Machine{}, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(0, nil)
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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(2), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"OVERDUE001", "OVERDUE002"}).Return(map[string]int{"OVERDUE001": 1, "OVERDUE002": 2}, nil)

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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: machines.PPMStatusDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(2), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"DUE001", "DUE002"}).Return(map[string]int{"DUE001": 1, "DUE002": 2}, nil)

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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: machines.PPMStatusAlmostDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(2), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"ALMOSTDUE001", "ALMOSTDUE002"}).Return(map[string]int{"ALMOSTDUE001": 1, "ALMOSTDUE002": 2}, nil)

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
			Sort:            machines.SortOrderUpdatedAtAsc,
			PpmStatusFilter: machines.PPMStatusDue,
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(1), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"FILTERED001"}).Return(map[string]int{"FILTERED001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?limit=10&offset=5&sort=updated_at_asc&ppm_status_filter=due", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(5), response.Offset)
		suite.Assert().Equal("updated_at_asc", response.Sort)
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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect Search to be called instead of List
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "HP", expectedOptions).Return(expectedMachines, nil)
		// Expect CountSearch to be called for search results
		suite.mockMachinesRepo.EXPECT().CountSearch(gomock.Any(), "HP", expectedOptions).Return(2, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"SEARCH001", "SEARCH002"}).Return(map[string]int{"SEARCH001": 1, "SEARCH002": 2}, nil)

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
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: "",
		}

		// Expect List to be called when q parameter is empty
		suite.mockMachinesRepo.EXPECT().List(gomock.Any(), expectedOptions).Return(expectedMachines, nil)
		suite.mockMachinesRepo.EXPECT().Count(gomock.Any(), gomock.Any()).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(0), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"LIST001"}).Return(map[string]int{"LIST001": 1}, nil)

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
			Sort:            machines.SortOrderUpdatedAtAsc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		// Expect Search to be called with combined parameters
		suite.mockMachinesRepo.EXPECT().Search(gomock.Any(), "printer", expectedOptions).Return(expectedMachines, nil)
		// Expect CountSearch to be called for search results
		suite.mockMachinesRepo.EXPECT().CountSearch(gomock.Any(), "printer", expectedOptions).Return(1, nil)
		suite.mockMachinesRepo.EXPECT().CountByStatus(gomock.Any()).Return(int32(1), int32(0), int32(0), nil)
		suite.mockMaintenanceRepo.EXPECT().CountByMachineSerials(gomock.Any(), []string{"COMBINED001"}).Return(map[string]int{"COMBINED001": 1}, nil)

		req := httptest.NewRequest("GET", "/machines?q=printer&limit=10&offset=5&sort=updated_at_asc&ppm_status_filter=overdue", nil)
		w := httptest.NewRecorder()

		suite.handler.ListMachines(w, req)

		suite.Assert().Equal(http.StatusOK, w.Code)

		var response handlers.ListMachinesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		suite.Require().NoError(err)

		suite.Assert().Len(response.Machines, 1)
		suite.Assert().Equal(int32(10), response.Limit)
		suite.Assert().Equal(int32(5), response.Offset)
		suite.Assert().Equal("updated_at_asc", response.Sort)
		suite.Assert().Equal(int32(1), response.OverdueCount)
		suite.Assert().Equal(int32(1), response.Count) // Verify search count is used
	})

	suite.Run("should return error when Search fails", func() {
		expectedOptions := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
			Sort:            machines.SortOrderUpdatedAtDesc,
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
