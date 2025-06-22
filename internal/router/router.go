package router

import (
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handler"

	"github.com/gorilla/mux"
)

// NewRouter creates a new HTTP router with all endpoints using gorilla/mux
func NewRouter(deps *deps.Dependencies) http.Handler {
	r := mux.NewRouter()

	// Health endpoint (no auth required)
	r.HandleFunc("/health", handler.NewHealthHandler(deps).Health).Methods(http.MethodGet)

	// Global OPTIONS handler for CORS preflight
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	// Subrouter for protected endpoints
	api := r.PathPrefix("/").Subrouter()

	// Machine endpoints
	api.HandleFunc("/machines/{serial_number}", handler.NewMachineHandler(deps).GetMachine).Methods(http.MethodGet)
	api.HandleFunc("/machines", handler.NewMachineHandler(deps).CreateMachine).Methods(http.MethodPost)
	api.HandleFunc("/machines", handler.NewMachineHandler(deps).UpdateMachine).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}", handler.NewMachineHandler(deps).DeleteMachine).Methods(http.MethodDelete)

	// Maintenance endpoints
	api.HandleFunc("/machines/{serial_number}/maintenance", handler.NewMaintenanceHandler(deps).ListMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.NewMaintenanceHandler(deps).GetMaintenance).Methods(http.MethodGet)
	api.HandleFunc("/machines/{serial_number}/maintenance", handler.NewMaintenanceHandler(deps).CreateMaintenance).Methods(http.MethodPost)
	api.HandleFunc("/machines/{serial_number}/maintenance", handler.NewMaintenanceHandler(deps).UpdateMaintenance).Methods(http.MethodPut)
	api.HandleFunc("/machines/{serial_number}/maintenance/{work_order_number}", handler.NewMaintenanceHandler(deps).DeleteMaintenance).Methods(http.MethodDelete)

	// Apply middleware to protected endpoints
	api.Use(handler.LoggingMiddleware)
	api.Use(handler.CORSMiddleware)
	api.Use(handler.BasicAuthMiddleware(deps.Config.Credentials))

	return r
}
