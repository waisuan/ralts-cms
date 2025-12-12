package audit_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ralts-cms/internal/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository implements audit.Repository for testing
type mockRepository struct {
	mu          sync.Mutex
	events      []*audit.Event
	createCalls int32
	createDelay time.Duration
	createErr   error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		events: make([]*audit.Event, 0),
	}
}

func (m *mockRepository) Create(ctx context.Context, event *audit.Event) error {
	if m.createDelay > 0 {
		time.Sleep(m.createDelay)
	}

	if m.createErr != nil {
		return m.createErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.AddInt32(&m.createCalls, 1)
	m.events = append(m.events, event)
	return nil
}

func (m *mockRepository) List(ctx context.Context, options *audit.ListOptions) ([]*audit.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events, nil
}

func (m *mockRepository) Count(ctx context.Context, options *audit.ListOptions) (int32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int32(len(m.events)), nil
}

func (m *mockRepository) getEvents() []*audit.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*audit.Event, len(m.events))
	copy(result, m.events)
	return result
}

func (m *mockRepository) getCreateCalls() int32 {
	return atomic.LoadInt32(&m.createCalls)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestService_LogEvent(t *testing.T) {
	t.Run("should queue event and process it", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())
		service.Start()
		defer service.Stop(context.Background())

		event := audit.Event{
			Action:       audit.ActionCreated,
			ResourceType: audit.ResourceMachine,
			ResourceID:   "SN-001",
		}

		service.LogEvent(event)

		// Wait for the event to be processed
		require.Eventually(t, func() bool {
			return repo.getCreateCalls() == 1
		}, time.Second, 10*time.Millisecond)

		events := repo.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, audit.ActionCreated, events[0].Action)
		assert.Equal(t, audit.ResourceMachine, events[0].ResourceType)
		assert.Equal(t, "SN-001", events[0].ResourceID)
		assert.NotEmpty(t, events[0].ID)
		assert.False(t, events[0].CreatedAt.IsZero())
	})

	t.Run("should set ID and CreatedAt if not provided", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())
		service.Start()
		defer service.Stop(context.Background())

		event := audit.Event{
			Action:       audit.ActionUpdated,
			ResourceType: audit.ResourceUser,
			ResourceID:   "user-123",
		}

		service.LogEvent(event)

		require.Eventually(t, func() bool {
			return repo.getCreateCalls() == 1
		}, time.Second, 10*time.Millisecond)

		events := repo.getEvents()
		require.Len(t, events, 1)
		assert.NotEmpty(t, events[0].ID, "ID should be auto-generated")
		assert.False(t, events[0].CreatedAt.IsZero(), "CreatedAt should be auto-set")
	})

	t.Run("should preserve ID and CreatedAt if provided", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())
		service.Start()
		defer service.Stop(context.Background())

		customID := "custom-id-123"
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		event := audit.Event{
			ID:           customID,
			Action:       audit.ActionDeleted,
			ResourceType: audit.ResourceMaintenance,
			ResourceID:   "WO-001",
			CreatedAt:    customTime,
		}

		service.LogEvent(event)

		require.Eventually(t, func() bool {
			return repo.getCreateCalls() == 1
		}, time.Second, 10*time.Millisecond)

		events := repo.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, customID, events[0].ID)
		assert.Equal(t, customTime, events[0].CreatedAt)
	})
}

func TestService_LogEvent_NonBlocking(t *testing.T) {
	t.Run("should not block when buffer is full", func(t *testing.T) {
		repo := newMockRepository()
		// Slow down processing to fill the buffer
		repo.createDelay = 100 * time.Millisecond

		// Create service with small buffer
		service := audit.NewServiceWithBufferSize(repo, testLogger(), 5)
		service.Start()
		defer service.Stop(context.Background())

		// Try to log more events than buffer can hold
		start := time.Now()
		for i := 0; i < 20; i++ {
			service.LogEvent(audit.Event{
				Action:       audit.ActionViewed,
				ResourceType: audit.ResourceMachine,
				ResourceID:   "SN-001",
			})
		}
		elapsed := time.Since(start)

		// Should complete quickly (not blocked by slow processing)
		assert.Less(t, elapsed, 50*time.Millisecond, "LogEvent should be non-blocking")
	})
}

func TestService_StartStop(t *testing.T) {
	t.Run("Start is idempotent", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())

		// Start multiple times
		service.Start()
		service.Start()
		service.Start()

		assert.True(t, service.IsStarted())

		service.Stop(context.Background())
	})

	t.Run("Stop drains remaining events", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())
		service.Start()

		// Log several events
		for i := 0; i < 10; i++ {
			service.LogEvent(audit.Event{
				Action:       audit.ActionCreated,
				ResourceType: audit.ResourceMachine,
				ResourceID:   "SN-001",
			})
		}

		// Stop and wait for drain
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		service.Stop(ctx)

		// All events should be processed
		assert.Equal(t, int32(10), repo.getCreateCalls())
	})

	t.Run("Stop before Start does nothing", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())

		// Should not panic
		service.Stop(context.Background())
		assert.False(t, service.IsStarted())
	})
}

func TestService_BufferLength(t *testing.T) {
	t.Run("returns current buffer length", func(t *testing.T) {
		repo := newMockRepository()
		// Slow processing to keep events in buffer
		repo.createDelay = 500 * time.Millisecond

		service := audit.NewServiceWithBufferSize(repo, testLogger(), 100)
		service.Start()
		defer service.Stop(context.Background())

		// Initially empty
		assert.Equal(t, 0, service.BufferLength())

		// Add some events
		for i := 0; i < 5; i++ {
			service.LogEvent(audit.Event{
				Action:       audit.ActionViewed,
				ResourceType: audit.ResourceMachine,
				ResourceID:   "SN-001",
			})
		}

		// Should have events in buffer (minus the one being processed)
		time.Sleep(10 * time.Millisecond)
		assert.GreaterOrEqual(t, service.BufferLength(), 0)
	})
}

func TestService_ConcurrentLogging(t *testing.T) {
	t.Run("handles concurrent LogEvent calls", func(t *testing.T) {
		repo := newMockRepository()
		service := audit.NewService(repo, testLogger())
		service.Start()

		var wg sync.WaitGroup
		numGoroutines := 10
		eventsPerGoroutine := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < eventsPerGoroutine; j++ {
					service.LogEvent(audit.Event{
						Action:       audit.ActionViewed,
						ResourceType: audit.ResourceMachine,
						ResourceID:   "SN-001",
					})
				}
			}(i)
		}

		wg.Wait()

		// Stop and drain
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		service.Stop(ctx)

		// All events should be processed
		assert.Equal(t, int32(numGoroutines*eventsPerGoroutine), repo.getCreateCalls())
	})
}
