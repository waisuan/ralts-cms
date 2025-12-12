package audit_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"ralts-cms/internal/audit"
	"ralts-cms/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AuditRepositoryTestSuite defines the test suite for audit repository
type AuditRepositoryTestSuite struct {
	suite.Suite

	db   *testutils.TestDatabase
	repo audit.Repository
}

// SetupSuite sets up the test suite
func (suite *AuditRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = audit.NewRepository(db.PostgresClient)
}

// TearDownSuite tears down the test suite
func (suite *AuditRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

// TearDownSubTest cleans up after each subtest
func (suite *AuditRepositoryTestSuite) TearDownSubTest() {
	ctx := context.Background()
	err := suite.db.CleanupAllTables(ctx)
	require.NoError(suite.T(), err)
}

// createTestEvent creates a test audit event with the given parameters
func createTestEvent(action, resourceType, resourceID string) *audit.Event {
	return &audit.Event{
		ID:           uuid.New().String(),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now().UTC(),
	}
}

// createTestEventWithUser creates a test audit event with a user ID
func createTestEventWithUser(userID, action, resourceType, resourceID string) *audit.Event {
	event := createTestEvent(action, resourceType, resourceID)
	event.UserID = &userID
	return event
}

// createTestEventWithDetails creates a test audit event with details
func createTestEventWithDetails(action, resourceType, resourceID string, details map[string]any) *audit.Event {
	event := createTestEvent(action, resourceType, resourceID)
	event.Details = details
	return event
}

func (suite *AuditRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create audit event successfully", func() {
		event := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")

		err := suite.repo.Create(ctx, event)
		suite.Require().NoError(err)
	})

	suite.Run("should create audit event with user ID", func() {
		// First create a user to satisfy foreign key constraint
		userID := suite.createTestUser(ctx)

		event := createTestEventWithUser(userID, audit.ActionUpdated, audit.ResourceMachine, "SN-002")

		err := suite.repo.Create(ctx, event)
		suite.Require().NoError(err)

		// Verify the event was created with user ID
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Assert().NotNil(events[0].UserID)
		suite.Assert().Equal(userID, *events[0].UserID)
	})

	suite.Run("should create audit event with details", func() {
		details := map[string]any{
			"customer": "Test Customer",
			"model":    "Test Model",
		}
		event := createTestEventWithDetails(audit.ActionCreated, audit.ResourceMachine, "SN-003", details)

		err := suite.repo.Create(ctx, event)
		suite.Require().NoError(err)

		// Verify the event was created with details
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Assert().Equal("Test Customer", events[0].Details["customer"])
		suite.Assert().Equal("Test Model", events[0].Details["model"])
	})

	suite.Run("should create audit event without user ID", func() {
		event := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-004")

		err := suite.repo.Create(ctx, event)
		suite.Require().NoError(err)

		// Verify the event was created without user ID
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Assert().Nil(events[0].UserID)
	})

	suite.Run("should create audit event without details", func() {
		event := createTestEvent(audit.ActionDeleted, audit.ResourceMaintenance, "WO-001")
		event.Details = nil

		err := suite.repo.Create(ctx, event)
		suite.Require().NoError(err)

		// Verify the event was created without details
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Require().Len(events, 1)
		suite.Assert().Nil(events[0].Details)
	})
}

func (suite *AuditRepositoryTestSuite) TestList() {
	ctx := context.Background()

	suite.Run("should list all audit events", func() {
		// Create multiple events
		for i := 0; i < 5; i++ {
			event := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 5)
	})

	suite.Run("should list events with pagination", func() {
		// Create 10 events
		for i := 0; i < 10; i++ {
			event := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Get first page
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  5,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 5)

		// Get second page
		events, err = suite.repo.List(ctx, &audit.ListOptions{
			Limit:  5,
			Offset: 5,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 5)

		// Get third page (should be empty)
		events, err = suite.repo.List(ctx, &audit.ListOptions{
			Limit:  5,
			Offset: 10,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 0)
	})

	suite.Run("should filter by resource type", func() {
		// Create events with different resource types
		event1 := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEvent(audit.ActionCreated, audit.ResourceMaintenance, "WO-001")
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		event3 := createTestEvent(audit.ActionCreated, audit.ResourceUser, "1")
		err = suite.repo.Create(ctx, event3)
		suite.Require().NoError(err)

		// Filter by machine resource type
		resourceType := audit.ResourceMachine
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:        10,
			Offset:       0,
			ResourceType: &resourceType,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 1)
		suite.Assert().Equal(audit.ResourceMachine, events[0].ResourceType)
	})

	suite.Run("should filter by action", func() {
		// Create events with different actions
		event1 := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEvent(audit.ActionUpdated, audit.ResourceMachine, "SN-002")
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		event3 := createTestEvent(audit.ActionDeleted, audit.ResourceMachine, "SN-003")
		err = suite.repo.Create(ctx, event3)
		suite.Require().NoError(err)

		// Filter by created action
		action := audit.ActionCreated
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
			Action: &action,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 1)
		suite.Assert().Equal(audit.ActionCreated, events[0].Action)
	})

	suite.Run("should filter by resource ID", func() {
		// Create events with different resource IDs
		event1 := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-001")
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-002")
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		// Filter by resource ID
		resourceID := "SN-001"
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:      10,
			Offset:     0,
			ResourceID: &resourceID,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 1)
		suite.Assert().Equal("SN-001", events[0].ResourceID)
	})

	suite.Run("should filter by user ID", func() {
		// Create a user first
		userID := suite.createTestUser(ctx)

		// Create events with and without user ID
		event1 := createTestEventWithUser(userID, audit.ActionCreated, audit.ResourceMachine, "SN-001")
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-002")
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		// Filter by user ID
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
			UserID: &userID,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 1)
		suite.Assert().NotNil(events[0].UserID)
		suite.Assert().Equal(userID, *events[0].UserID)
	})

	suite.Run("should order by created_at desc", func() {
		// Create events with different timestamps
		event1 := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")
		event1.CreatedAt = time.Now().Add(-2 * time.Hour)
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEvent(audit.ActionUpdated, audit.ResourceMachine, "SN-002")
		event2.CreatedAt = time.Now().Add(-1 * time.Hour)
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		event3 := createTestEvent(audit.ActionDeleted, audit.ResourceMachine, "SN-003")
		event3.CreatedAt = time.Now()
		err = suite.repo.Create(ctx, event3)
		suite.Require().NoError(err)

		// List should return newest first
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 3)
		suite.Assert().Equal("SN-003", events[0].ResourceID)
		suite.Assert().Equal("SN-002", events[1].ResourceID)
		suite.Assert().Equal("SN-001", events[2].ResourceID)
	})

	suite.Run("should return empty list when no events", func() {
		events, err := suite.repo.List(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 0)
	})
}

func (suite *AuditRepositoryTestSuite) TestCount() {
	ctx := context.Background()

	suite.Run("should count all audit events", func() {
		// Create multiple events
		for i := 0; i < 5; i++ {
			event := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		count, err := suite.repo.Count(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(5), count)
	})

	suite.Run("should count events with filter", func() {
		// Create events with different resource types
		for i := 0; i < 3; i++ {
			event := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		for i := 0; i < 2; i++ {
			event := createTestEvent(audit.ActionCreated, audit.ResourceMaintenance, "WO-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Count machine events only
		resourceType := audit.ResourceMachine
		count, err := suite.repo.Count(ctx, &audit.ListOptions{
			Limit:        10,
			Offset:       0,
			ResourceType: &resourceType,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(3), count)
	})

	suite.Run("should return zero when no events", func() {
		count, err := suite.repo.Count(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(0), count)
	})

	suite.Run("should count with multiple filters", func() {
		// Create a user
		userID := suite.createTestUser(ctx)

		// Create events
		event1 := createTestEventWithUser(userID, audit.ActionCreated, audit.ResourceMachine, "SN-001")
		err := suite.repo.Create(ctx, event1)
		suite.Require().NoError(err)

		event2 := createTestEventWithUser(userID, audit.ActionUpdated, audit.ResourceMachine, "SN-001")
		err = suite.repo.Create(ctx, event2)
		suite.Require().NoError(err)

		event3 := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-001")
		err = suite.repo.Create(ctx, event3)
		suite.Require().NoError(err)

		// Count with user and action filter
		action := audit.ActionCreated
		count, err := suite.repo.Count(ctx, &audit.ListOptions{
			Limit:  10,
			Offset: 0,
			UserID: &userID,
			Action: &action,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(1), count)
	})
}

func (suite *AuditRepositoryTestSuite) TestDeleteOlderThan() {
	ctx := context.Background()

	suite.Run("should delete events older than cutoff", func() {
		// Create old events (10 days ago)
		for i := 0; i < 3; i++ {
			event := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-OLD")
			event.CreatedAt = time.Now().AddDate(0, 0, -10)
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Create recent events (1 day ago)
		for i := 0; i < 2; i++ {
			event := createTestEvent(audit.ActionCreated, audit.ResourceMachine, "SN-NEW")
			event.CreatedAt = time.Now().AddDate(0, 0, -1)
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Delete events older than 7 days
		cutoff := time.Now().AddDate(0, 0, -7)
		deleted, err := suite.repo.DeleteOlderThan(ctx, cutoff)
		suite.Require().NoError(err)
		suite.Assert().Equal(int64(3), deleted)

		// Verify only new events remain
		events, err := suite.repo.List(ctx, &audit.ListOptions{Limit: 10, Offset: 0})
		suite.Require().NoError(err)
		suite.Assert().Len(events, 2)
		for _, e := range events {
			suite.Assert().Equal("SN-NEW", e.ResourceID)
		}
	})

	suite.Run("should return zero when no old events", func() {
		// Create only recent events
		for i := 0; i < 3; i++ {
			event := createTestEvent(audit.ActionViewed, audit.ResourceMachine, "SN-NEW")
			event.CreatedAt = time.Now()
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Try to delete events older than 7 days (none exist)
		cutoff := time.Now().AddDate(0, 0, -7)
		deleted, err := suite.repo.DeleteOlderThan(ctx, cutoff)
		suite.Require().NoError(err)
		suite.Assert().Equal(int64(0), deleted)

		// Verify all events still exist
		count, err := suite.repo.Count(ctx, &audit.ListOptions{Limit: 10, Offset: 0})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(3), count)
	})

	suite.Run("should delete all events when cutoff is in the future", func() {
		// Create events
		for i := 0; i < 3; i++ {
			event := createTestEvent(audit.ActionDeleted, audit.ResourceMachine, "SN-001")
			err := suite.repo.Create(ctx, event)
			suite.Require().NoError(err)
		}

		// Delete events with future cutoff
		cutoff := time.Now().Add(1 * time.Hour)
		deleted, err := suite.repo.DeleteOlderThan(ctx, cutoff)
		suite.Require().NoError(err)
		suite.Assert().Equal(int64(3), deleted)

		// Verify no events remain
		count, err := suite.repo.Count(ctx, &audit.ListOptions{Limit: 10, Offset: 0})
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(0), count)
	})

	suite.Run("should handle empty table", func() {
		cutoff := time.Now().AddDate(0, 0, -7)
		deleted, err := suite.repo.DeleteOlderThan(ctx, cutoff)
		suite.Require().NoError(err)
		suite.Assert().Equal(int64(0), deleted)
	})
}

// createTestUser creates a test user and returns the user ID as string
func (suite *AuditRepositoryTestSuite) createTestUser(ctx context.Context) string {
	// Insert a test user directly into the database
	var userID int64
	err := suite.db.PostgresClient.QueryRow(ctx, `
		INSERT INTO users (username, email, password, salt, role, approved, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, "testuser"+uuid.New().String()[:8], "test"+uuid.New().String()[:8]+"@example.com", "hashedpassword", "testsalt", "user", true, time.Now()).Scan(&userID)
	suite.Require().NoError(err)

	return strconv.FormatInt(userID, 10)
}

func TestAuditRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AuditRepositoryTestSuite))
}
