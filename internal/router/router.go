package router

import (
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/handler"
	"strings"
)

// NewRouter creates a new HTTP router with all endpoints
func NewRouter(deps *deps.Dependencies) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	machineHandler := handler.NewMachineHandler(deps)
	maintenanceHandler := handler.NewMaintenanceHandler(deps)
	healthHandler := handler.NewHealthHandler(deps)

	// Health endpoint (no auth required)
	mux.HandleFunc("/health", healthHandler.Health)

	// Machine endpoints (protected)
	mux.HandleFunc("/machines/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		if strings.HasPrefix(path, "/machines/") && strings.Contains(path, "/maintenance") {
			// Maintenance endpoints
			if strings.HasSuffix(path, "/maintenance") && method == http.MethodGet {
				maintenanceHandler.ListMaintenance(w, r)
				return
			}
			if strings.HasSuffix(path, "/maintenance") && method == http.MethodPost {
				maintenanceHandler.CreateMaintenance(w, r)
				return
			}
			if strings.HasSuffix(path, "/maintenance") && method == http.MethodPut {
				maintenanceHandler.UpdateMaintenance(w, r)
				return
			}
			if strings.Contains(path, "/maintenance/") && method == http.MethodGet {
				maintenanceHandler.GetMaintenance(w, r)
				return
			}
			if strings.Contains(path, "/maintenance/") && method == http.MethodDelete {
				maintenanceHandler.DeleteMaintenance(w, r)
				return
			}
		} else {
			// Machine endpoints
			if method == http.MethodGet {
				machineHandler.GetMachine(w, r)
				return
			}
			if method == http.MethodPost && path == "/machines" {
				machineHandler.CreateMachine(w, r)
				return
			}
			if method == http.MethodPut && path == "/machines" {
				machineHandler.UpdateMachine(w, r)
				return
			}
			if method == http.MethodDelete {
				machineHandler.DeleteMachine(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})

	// Apply middleware to all except /health
	protected := handler.LoggingMiddleware(mux)
	protected = handler.CORSMiddleware(protected)
	protected = handler.BasicAuthMiddleware(deps.Config.Credentials)(protected)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			mux.ServeHTTP(w, r)
			return
		}
		protected.ServeHTTP(w, r)
	})
}
