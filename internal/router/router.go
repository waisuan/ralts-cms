// Package router provides HTTP routing functionality for the Ralts-CMS application.
// It sets up all API endpoints with appropriate middleware and authentication.
package router

import (
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handlers"
	"ralts-cms/internal/middlewares"

	"github.com/gorilla/mux"
)

// NewRouter creates a new HTTP router with all endpoints using gorilla/mux
func NewRouter(deps *deps.Dependencies) http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware to all endpoints
	r.Use(middlewares.CORSMiddleware)
	// Apply logging middleware to all endpoints
	r.Use(middlewares.LoggingMiddleware)

	// Health endpoint (no auth required)
	r.HandleFunc("/health", handlers.NewHealthHandler(deps).Health).Methods(http.MethodGet)

	// Global OPTIONS handler for CORS preflight
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	usersHandler := handlers.NewUsersHandler(deps)
	authHandler := handlers.NewAuthHandler(deps)

	// Public endpoints (no auth required)
	r.HandleFunc("/api/v1/users/login", usersHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/users", usersHandler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods(http.MethodPost)

	// Subrouter for protected endpoints
	api := r.PathPrefix("/api/v1/").Subrouter()

	// CSV export endpoints (protected) — registered before parameterized routes
	api.HandleFunc("/machines/export/csv", handlers.NewCSVExportHandler(deps).ExportMachinesCSV).Methods(http.MethodGet)

	// Machine endpoints (protected)
	api.HandleFunc("/machines", handlers.NewMachinesHandler(deps).ListMachines).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).GetMachine).Methods(http.MethodGet)
	api.HandleFunc("/machines", handlers.NewMachinesHandler(deps).CreateMachine).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).UpdateMachine).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).DeleteMachine).Methods(http.MethodDelete)

	// Maintenance CSV export (protected) — registered before parameterized routes
	api.HandleFunc("/machines/{serial_number}/maintenance/export/csv", handlers.NewCSVExportHandler(deps).ExportMaintenanceCSV).Methods(http.MethodGet)

	// Maintenance endpoints (protected)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).ListMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handlers.NewMaintenanceHandler(deps).GetMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).CreateMaintenance).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).UpdateMaintenance).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handlers.NewMaintenanceHandler(deps).DeleteMaintenance).Methods(http.MethodDelete)

	// Machine flag endpoints (protected). Create requires role=ADMIN and Resolve
	// requires admin-or-assignee; both are enforced inside the handlers so we can
	// keep the natural REST nesting under /machines/{serial_number}/flags.
	flagsHandler := handlers.NewFlagsHandler(deps)
	// Batch endpoint registered before parameterised /machines/{serial_number}
	// to avoid collision with the static "flags" segment.
	api.HandleFunc("/machines/flags/open-by-machine", flagsHandler.ListOpenByMachine).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/flags", flagsHandler.ListForMachine).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/flags", flagsHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/flags/{id}/resolve", flagsHandler.Resolve).Methods(http.MethodPost)
	// Cross-machine flagged-records listing. Admins see everything, everyone else
	// only flags on machines assigned to them; scoping happens in the handler.
	api.HandleFunc("/flags", flagsHandler.List).Methods(http.MethodGet)

	// Machine attachment endpoints (protected)
	api.HandleFunc("/machines/{serial_number}/attachments", handlers.NewAttachmentHandler(deps).CreateMachineAttachment).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).GetMachineAttachment).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).ReplaceMachineAttachment).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).DeleteMachineAttachment).Methods(http.MethodDelete)

	// Maintenance attachment endpoints (protected)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}/attachments", handlers.NewAttachmentHandler(deps).CreateMaintenanceAttachment).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).GetMaintenanceAttachment).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).ReplaceMaintenanceAttachment).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).DeleteMaintenanceAttachment).Methods(http.MethodDelete)

	// User self-service endpoints (protected)
	api.HandleFunc("/users/password", usersHandler.UpdatePassword).Methods(http.MethodPut)

	// Assignee dropdown source (any authenticated user).
	api.HandleFunc("/users/directory", usersHandler.ListUserDirectory).Methods(http.MethodGet)

	// Notification inbox endpoints (protected, per-user).
	notificationsHandler := handlers.NewNotificationsHandler(deps)
	api.HandleFunc("/notifications", notificationsHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/notifications/unread-count", notificationsHandler.UnreadCount).Methods(http.MethodGet)
	api.HandleFunc("/notifications/read-all", notificationsHandler.MarkAllRead).Methods(http.MethodPost)
	api.HandleFunc("/notifications/{id}/read", notificationsHandler.MarkRead).Methods(http.MethodPost)

	// Apply middleware to protected endpoints only
	api.Use(middlewares.AuthenticationMiddleware(deps.Config.JWTSecret))

	// Admin-only subrouter for admin endpoints
	adminAPI := r.PathPrefix("/api/v1/admin/").Subrouter()

	// Admin user management endpoints (protected by both auth and admin middleware)
	adminAPI.HandleFunc("/users", usersHandler.ListUsers).Methods(http.MethodGet)
	adminAPI.HandleFunc("/users/{id:[0-9]+}/status", usersHandler.UpdateUserStatus).Methods(http.MethodPut)
	adminAPI.HandleFunc("/users/bulk-status", usersHandler.BulkUpdateStatus).Methods(http.MethodPut)

	// Admin audit endpoints
	adminAPI.HandleFunc("/audit/events", handlers.NewAuditHandler(deps).ListAuditEvents).Methods(http.MethodGet)

	// Apply both authentication and admin middleware to admin endpoints
	adminAPI.Use(middlewares.AuthenticationMiddleware(deps.Config.JWTSecret))
	adminAPI.Use(middlewares.AdminOnlyMiddleware())

	return r
}
