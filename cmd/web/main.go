// Package main provides the web server entry point for the Ralts-CMS application.
// It handles server startup, graceful shutdown, and dependency initialization.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/router"
)

func main() {
	// Initialize dependencies
	deps := deps.Initialise()
	logger := deps.Logger

	// Create router
	handler := router.NewRouter(deps)

	// Create server with configurable timeouts
	server := &http.Server{
		Addr:         ":" + deps.Config.ServerPort,
		Handler:      handler,
		ReadTimeout:  deps.Config.ReadTimeout,
		WriteTimeout: deps.Config.WriteTimeout,
		IdleTimeout:  deps.Config.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server",
			"port", deps.Config.ServerPort,
			"read_timeout", deps.Config.ReadTimeout,
			"write_timeout", deps.Config.WriteTimeout,
			"idle_timeout", deps.Config.IdleTimeout)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
