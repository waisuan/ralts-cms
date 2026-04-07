package audit

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultBufferSize is the default size of the event buffer channel
	DefaultBufferSize = 1000
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
	// DefaultRetentionDays is the default number of days to retain audit events
	DefaultRetentionDays = 7
	// DefaultCleanupInterval is the default interval between cleanup runs
	DefaultCleanupInterval = 1 * time.Hour
)

// AuditService defines the interface for audit logging operations.
//
//go:generate mockgen -destination=mock_audit_service.go -package=audit -source=service.go AuditService
//revive:disable-next-line:exported // Name retained for stable mockgen output and call sites.
type AuditService interface {
	LogEvent(event Event)
	Start()
	Stop(ctx context.Context)
}

// ServiceConfig holds configuration for the audit service
type ServiceConfig struct {
	BufferSize      int
	RetentionDays   int
	CleanupInterval time.Duration
}

// DefaultServiceConfig returns the default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		BufferSize:      DefaultBufferSize,
		RetentionDays:   DefaultRetentionDays,
		CleanupInterval: DefaultCleanupInterval,
	}
}

// Service handles asynchronous audit logging
type Service struct {
	repo      Repository
	logger    *slog.Logger
	config    ServiceConfig
	eventChan chan Event
	wg        sync.WaitGroup
	stopChan  chan struct{}
	started   bool
	// terminalShutdown is set when Stop closes stopChan; Start becomes a no-op afterward.
	terminalShutdown bool
	mu               sync.Mutex
	stopOnce         sync.Once
	cleanupTicker    *time.Ticker
}

// NewService creates a new audit service with the given repository and logger
func NewService(repo Repository, logger *slog.Logger) *Service {
	return NewServiceWithConfig(repo, logger, DefaultServiceConfig())
}

// NewServiceWithBufferSize creates a new audit service with a custom buffer size
func NewServiceWithBufferSize(repo Repository, logger *slog.Logger, bufferSize int) *Service {
	config := DefaultServiceConfig()
	config.BufferSize = bufferSize
	return NewServiceWithConfig(repo, logger, config)
}

// NewServiceWithConfig creates a new audit service with custom configuration
func NewServiceWithConfig(repo Repository, logger *slog.Logger, config ServiceConfig) *Service {
	if config.BufferSize <= 0 {
		config.BufferSize = DefaultBufferSize
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = DefaultRetentionDays
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = DefaultCleanupInterval
	}

	return &Service{
		repo:      repo,
		logger:    logger,
		config:    config,
		eventChan: make(chan Event, config.BufferSize),
		stopChan:  make(chan struct{}),
	}
}

// LogEvent queues an audit event for asynchronous processing.
// This method is non-blocking and will drop events if the buffer is full.
func (s *Service) LogEvent(event Event) {
	// Set defaults
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	// Non-blocking send - drop event if buffer is full
	select {
	case s.eventChan <- event:
		// Event queued successfully
	default:
		s.logger.Warn("Audit buffer full, dropping event",
			"action", event.Action,
			"resource_type", event.ResourceType,
			"resource_id", event.ResourceID,
		)
	}
}

// Start begins the background worker goroutine that processes events.
// It is idempotent while running. After a successful Stop that shut down workers,
// Start does nothing (the service is terminal; use a new Service if needed).
func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.terminalShutdown {
		s.logger.Warn("audit service: Start ignored after shutdown")
		return
	}

	if s.started {
		return
	}

	s.started = true

	// Start event worker
	s.wg.Add(1)
	go s.worker()

	// Start cleanup worker
	s.cleanupTicker = time.NewTicker(s.config.CleanupInterval)
	s.wg.Add(1)
	go s.cleanupWorker()

	s.logger.Info("Audit service started",
		"retention_days", s.config.RetentionDays,
		"cleanup_interval", s.config.CleanupInterval.String(),
	)
}

// Stop gracefully shuts down the audit service, draining remaining events.
// It blocks until all events are processed or the context is cancelled.
// Stop is idempotent: multiple calls after the first complete without panicking.
func (s *Service) Stop(ctx context.Context) {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		if !s.started {
			s.mu.Unlock()
			return
		}
		s.started = false
		s.terminalShutdown = true
		s.mu.Unlock()

		if s.cleanupTicker != nil {
			s.cleanupTicker.Stop()
		}

		close(s.stopChan)

		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			s.logger.Info("Audit service stopped gracefully")
		case <-ctx.Done():
			s.logger.Warn("Audit service shutdown timed out, some events may be lost")
		}
	})
}

// worker processes events from the channel and persists them to the database
func (s *Service) worker() {
	defer s.wg.Done()

	for {
		select {
		case event := <-s.eventChan:
			s.processEvent(event)

		case <-s.stopChan:
			// Drain remaining events before exiting
			s.drainEvents()
			return
		}
	}
}

// cleanupWorker periodically deletes old audit events
func (s *Service) cleanupWorker() {
	defer s.wg.Done()

	// Run cleanup immediately on start
	s.runCleanup()

	for {
		select {
		case <-s.cleanupTicker.C:
			s.runCleanup()

		case <-s.stopChan:
			return
		}
	}
}

// runCleanup performs the actual cleanup of old audit events
func (s *Service) runCleanup() {
	startTime := time.Now()
	cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	s.logger.Info("Audit cleanup started",
		"retention_days", s.config.RetentionDays,
		"cutoff", cutoff.Format(time.RFC3339),
	)

	deleted, err := s.repo.DeleteOlderThan(context.Background(), cutoff)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.Error("Audit cleanup failed",
			"error", err,
			"duration", duration.String(),
		)
		return
	}

	s.logger.Info("Audit cleanup completed",
		"deleted", deleted,
		"duration", duration.String(),
	)
}

// processEvent persists a single event to the database
func (s *Service) processEvent(event Event) {
	if err := s.repo.Create(context.Background(), &event); err != nil {
		s.logger.Error("Failed to persist audit event",
			"error", err,
			"action", event.Action,
			"resource_type", event.ResourceType,
			"resource_id", event.ResourceID,
		)
	}
}

// drainEvents processes all remaining events in the buffer
func (s *Service) drainEvents() {
	for {
		select {
		case event := <-s.eventChan:
			s.processEvent(event)
		default:
			// Channel is empty, we're done
			return
		}
	}
}

// BufferLength returns the current number of events waiting in the buffer.
// Useful for monitoring and testing.
func (s *Service) BufferLength() int {
	return len(s.eventChan)
}

// IsStarted returns whether the service has been started.
func (s *Service) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}
