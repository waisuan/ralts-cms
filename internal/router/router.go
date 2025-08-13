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

	// Public endpoints (no auth required)
	r.HandleFunc("/api/v1/users/login", handlers.NewUsersHandler(deps).Login).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/users", handlers.NewUsersHandler(deps).CreateUser).Methods(http.MethodPost)

	// Subrouter for protected endpoints
	api := r.PathPrefix("/api/v1/").Subrouter()

	// Machine endpoints (protected)
	api.HandleFunc("/machines", handlers.NewMachinesHandler(deps).ListMachines).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).GetMachine).Methods(http.MethodGet)
	api.HandleFunc("/machines", handlers.NewMachinesHandler(deps).CreateMachine).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).UpdateMachine).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}", handlers.NewMachinesHandler(deps).DeleteMachine).Methods(http.MethodDelete)

	// Maintenance endpoints (protected)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).ListMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handlers.NewMaintenanceHandler(deps).GetMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).CreateMaintenance).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/maintenance", handlers.NewMaintenanceHandler(deps).UpdateMaintenance).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handlers.NewMaintenanceHandler(deps).DeleteMaintenance).Methods(http.MethodDelete)

	// Machine attachment endpoints (protected)
	api.HandleFunc("/machines/{serial_number}/attachments", handlers.NewAttachmentHandler(deps).CreateMachineAttachment).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).ReplaceMachineAttachment).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/attachments/{attachment_name}", handlers.NewAttachmentHandler(deps).DeleteMachineAttachment).Methods(http.MethodDelete)

	// Apply middleware to protected endpoints only
	api.Use(middlewares.AuthenticationMiddleware(deps.Config.JWTSecret))

	return r
}
