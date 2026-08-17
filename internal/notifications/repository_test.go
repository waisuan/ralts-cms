package notifications_test

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/notifications"
	"ralts-cms/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// NotificationsRepositoryTestSuite exercises the Postgres-backed notifications
// repository against a real database in a testcontainer.
type NotificationsRepositoryTestSuite struct {
	suite.Suite

	db   *testutils.TestDatabase
	repo notifications.Repository
}

func (suite *NotificationsRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = notifications.NewRepository(db.PostgresClient)
}

func (suite *NotificationsRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *NotificationsRepositoryTestSuite) TearDownSubTest() {
	err := suite.db.CleanupAllTables(context.Background())
	require.NoError(suite.T(), err)
}

// createTestUser inserts a bare user row and returns the generated id.
func (suite *NotificationsRepositoryTestSuite) createTestUser(ctx context.Context, username string) int64 {
	var id int64
	err := suite.db.PostgresClient.QueryRow(ctx, `
		INSERT INTO users (username, email, password, salt, role, approved, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, username+"_"+uuid.NewString()[:8],
		username+"_"+uuid.NewString()[:8]+"@example.com",
		"hashedpassword", "salt", "USER", true, time.Now()).Scan(&id)
	suite.Require().NoError(err)
	return id
}

func (suite *NotificationsRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("persists a notification for a user", func() {
		userID := suite.createTestUser(ctx, "recipient")

		n := &notifications.Notification{
			UserID: userID,
			Type:   notifications.TypeAssigned,
			Title:  "You have been assigned a machine",
			Body:   "SN-001",
		}

		err := suite.repo.Create(ctx, n)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(n.ID)
		suite.Assert().False(n.CreatedAt.IsZero())

		list, err := suite.repo.ListForUser(ctx, userID, notifications.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(list, 1)
		suite.Assert().Equal(n.ID, list[0].ID)
		suite.Assert().Equal(notifications.TypeAssigned, list[0].Type)
		suite.Assert().Equal("You have been assigned a machine", list[0].Title)
		suite.Assert().Nil(list[0].ReadAt)
	})
}

func (suite *NotificationsRepositoryTestSuite) TestListForUserFlagStatus() {
	ctx := context.Background()

	suite.Run("reports the current status of the referenced flag", func() {
		userID := suite.createTestUser(ctx, "assignee")

		serial := "SN-FLAGSTATUS"
		_, err := suite.db.PostgresClient.Exec(ctx, `
			INSERT INTO machines ("serialNumber", customer, "assignedUserId")
			VALUES ($1, 'Cust', $2)
		`, serial, userID)
		suite.Require().NoError(err)

		openFlagID := uuid.NewString()
		resolvedFlagID := uuid.NewString()
		_, err = suite.db.PostgresClient.Exec(ctx, `
			INSERT INTO machine_flags (id, machine_serial_number, reason, status)
			VALUES ($1, $3, 'missing_values', 'open'), ($2, $3, 'other', 'resolved')
		`, openFlagID, resolvedFlagID, serial)
		suite.Require().NoError(err)

		for _, flagID := range []string{openFlagID, resolvedFlagID} {
			id := flagID
			suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
				UserID: userID, Type: notifications.TypeFlagged, Title: "flagged",
				MachineSerialNumber: &serial, FlagID: &id,
			}))
		}
		// A notification with no flag at all must report no status.
		suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
			UserID: userID, Type: notifications.TypeAssigned, Title: "assigned",
			MachineSerialNumber: &serial,
		}))

		list, err := suite.repo.ListForUser(ctx, userID, notifications.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(list, 3)

		statusByFlag := map[string]*string{}
		var assignedStatus *string
		for _, n := range list {
			if n.FlagID == nil {
				assignedStatus = n.FlagStatus
				continue
			}
			statusByFlag[*n.FlagID] = n.FlagStatus
		}

		suite.Require().NotNil(statusByFlag[openFlagID])
		suite.Assert().Equal("open", *statusByFlag[openFlagID])
		suite.Require().NotNil(statusByFlag[resolvedFlagID])
		suite.Assert().Equal("resolved", *statusByFlag[resolvedFlagID])
		suite.Assert().Nil(assignedStatus)
	})
}

func (suite *NotificationsRepositoryTestSuite) TestListForUserScoping() {
	ctx := context.Background()

	suite.Run("only returns rows belonging to the requested user", func() {
		aliceID := suite.createTestUser(ctx, "alice")
		bobID := suite.createTestUser(ctx, "bob")

		suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
			UserID: aliceID, Type: notifications.TypeAssigned, Title: "for alice",
		}))
		suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
			UserID: aliceID, Type: notifications.TypeFlagged, Title: "for alice 2",
		}))
		suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
			UserID: bobID, Type: notifications.TypeAssigned, Title: "for bob",
		}))

		aliceList, err := suite.repo.ListForUser(ctx, aliceID, notifications.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Assert().Len(aliceList, 2)
		for _, n := range aliceList {
			suite.Assert().Equal(aliceID, n.UserID)
		}

		bobList, err := suite.repo.ListForUser(ctx, bobID, notifications.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Assert().Len(bobList, 1)
		suite.Assert().Equal(bobID, bobList[0].UserID)
	})
}

func (suite *NotificationsRepositoryTestSuite) TestUnreadFilterAndCount() {
	ctx := context.Background()

	suite.Run("unread filter and CountUnread", func() {
		userID := suite.createTestUser(ctx, "reader")

		n1 := &notifications.Notification{UserID: userID, Type: notifications.TypeAssigned, Title: "one"}
		n2 := &notifications.Notification{UserID: userID, Type: notifications.TypeFlagged, Title: "two"}
		n3 := &notifications.Notification{UserID: userID, Type: notifications.TypeFlagged, Title: "three"}
		suite.Require().NoError(suite.repo.Create(ctx, n1))
		suite.Require().NoError(suite.repo.Create(ctx, n2))
		suite.Require().NoError(suite.repo.Create(ctx, n3))

		unread, err := suite.repo.CountUnread(ctx, userID)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, unread)

		suite.Require().NoError(suite.repo.MarkRead(ctx, userID, n2.ID))

		unread, err = suite.repo.CountUnread(ctx, userID)
		suite.Require().NoError(err)
		suite.Assert().Equal(2, unread)

		filtered, err := suite.repo.ListForUser(ctx, userID, notifications.ListOptions{Limit: 10, UnreadOnly: true})
		suite.Require().NoError(err)
		suite.Assert().Len(filtered, 2)
		for _, n := range filtered {
			suite.Assert().Nil(n.ReadAt)
		}
	})
}

func (suite *NotificationsRepositoryTestSuite) TestMarkReadScopedToUser() {
	ctx := context.Background()

	suite.Run("cannot mark another user's notification read", func() {
		aliceID := suite.createTestUser(ctx, "alice")
		bobID := suite.createTestUser(ctx, "bob")

		n := &notifications.Notification{UserID: bobID, Type: notifications.TypeAssigned, Title: "for bob"}
		suite.Require().NoError(suite.repo.Create(ctx, n))

		suite.Require().NoError(suite.repo.MarkRead(ctx, aliceID, n.ID))

		list, err := suite.repo.ListForUser(ctx, bobID, notifications.ListOptions{Limit: 10})
		suite.Require().NoError(err)
		suite.Require().Len(list, 1)
		suite.Assert().Nil(list[0].ReadAt, "bob's notification should not have been marked read by alice")
	})
}

func (suite *NotificationsRepositoryTestSuite) TestMarkAllRead() {
	ctx := context.Background()

	suite.Run("marks all unread rows for the given user", func() {
		userID := suite.createTestUser(ctx, "reader")
		other := suite.createTestUser(ctx, "someone_else")

		for i := 0; i < 3; i++ {
			suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
				UserID: userID, Type: notifications.TypeAssigned, Title: "t",
			}))
		}
		suite.Require().NoError(suite.repo.Create(ctx, &notifications.Notification{
			UserID: other, Type: notifications.TypeAssigned, Title: "other",
		}))

		updated, err := suite.repo.MarkAllRead(ctx, userID)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, updated)

		unread, err := suite.repo.CountUnread(ctx, userID)
		suite.Require().NoError(err)
		suite.Assert().Equal(0, unread)

		otherUnread, err := suite.repo.CountUnread(ctx, other)
		suite.Require().NoError(err)
		suite.Assert().Equal(1, otherUnread, "MarkAllRead must not touch other users' rows")
	})
}

func TestNotificationsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationsRepositoryTestSuite))
}
