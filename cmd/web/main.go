// Package main provides the web server entry point for the Ralts-CMS application.
// It handles server startup, graceful shutdown, and dependency initialization.
package main

import (
	"context"
	"log"
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
		log.Printf("Starting server on port %s", deps.Config.ServerPort)
		log.Printf("HTTP timeouts - Read: %v, Write: %v, Idle: %v",
			deps.Config.ReadTimeout, deps.Config.WriteTimeout, deps.Config.IdleTimeout)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
