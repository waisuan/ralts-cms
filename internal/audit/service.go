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
)

// AuditService defines the interface for audit logging operations
//
//go:generate mockgen -destination=mock_audit_service.go -package=audit -source=service.go AuditService
type AuditService interface {
	LogEvent(event Event)
	Start()
	Stop(ctx context.Context)
}

// Service handles asynchronous audit logging
type Service struct {
	repo      Repository
	logger    *slog.Logger
	eventChan chan Event
	wg        sync.WaitGroup
	stopChan  chan struct{}
	started   bool
	mu        sync.Mutex
}

// NewService creates a new audit service with the given repository and logger
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:      repo,
		logger:    logger,
		eventChan: make(chan Event, DefaultBufferSize),
		stopChan:  make(chan struct{}),
	}
}

// NewServiceWithBufferSize creates a new audit service with a custom buffer size
func NewServiceWithBufferSize(repo Repository, logger *slog.Logger, bufferSize int) *Service {
	return &Service{
		repo:      repo,
		logger:    logger,
		eventChan: make(chan Event, bufferSize),
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
// This method is idempotent - calling it multiple times has no effect.
func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return
	}

	s.started = true
	s.wg.Add(1)
	go s.worker()
	s.logger.Info("Audit service started")
}

// Stop gracefully shuts down the audit service, draining remaining events.
// It blocks until all events are processed or the context is cancelled.
func (s *Service) Stop(ctx context.Context) {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	close(s.stopChan)

	// Wait for worker to finish or context timeout
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
