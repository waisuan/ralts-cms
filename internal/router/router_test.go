package router_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ralts-cms/internal/audit"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/router"
	"ralts-cms/internal/users"
	"ralts-cms/pkg/auth"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "router-test-jwt-secret"

func testDeps(t *testing.T, ctrl *gomock.Controller) *deps.Dependencies {
	t.Helper()

	return &deps.Dependencies{
		Config: &deps.Config{
			JWTSecret:               testJWTSecret,
			DefaultMachinesLimit:    50,
			MaxMachinesLimit:        100,
			DefaultMaintenanceLimit: 50,
			MaxMaintenanceLimit:     100,
		},
		Logger:                slog.New(slog.NewTextHandler(io.Discard, nil)),
		MachinesRepository:    machines.NewMockRepository(ctrl),
		MaintenanceRepository: nil,
		UsersRepository:       nil,
		AuditRepository:       nil,
		AttachmentService:     nil,
		AuditService:          audit.NewMockAuditService(ctrl),
	}
}

func bearer(role string) (string, error) {
	tok, err := auth.GenerateJWTToken(1, role, testJWTSecret)
	if err != nil {
		return "", err
	}
	return "Bearer " + tok, nil
}

func TestNewRouter_Health(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := testDeps(t, ctrl)
	h := router.NewRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")
	assert.Contains(t, rr.Body.String(), "healthy")
}

func TestNewRouter_OPTIONS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := router.NewRouter(testDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/machines", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewRouter_Protected_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := router.NewRouter(testDeps(t, ctrl))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestNewRouter_MachinesExportCSV_UsesListRouteNotSerial(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := testDeps(t, ctrl)
	mockMachines := d.MachinesRepository.(*machines.MockRepository)
	mockAudit := d.AuditService.(*audit.MockAuditService)

	mockMachines.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*machines.Machine{}, nil)
	mockAudit.EXPECT().LogEvent(gomock.Any()).MinTimes(1)

	authz, err := bearer(users.RoleAdmin)
	require.NoError(t, err)

	h := router.NewRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/machines/export/csv", nil)
	req.Header.Set("Authorization", authz)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/csv")
}

func TestNewRouter_Admin_ForbiddenForNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	d := testDeps(t, ctrl)
	authz, err := bearer("user")
	require.NoError(t, err)

	h := router.NewRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", authz)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}
