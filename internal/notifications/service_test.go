package notifications_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"ralts-cms/internal/notifications"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceNotify(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("persists a notification asynchronously", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := notifications.NewMockRepository(ctrl)
		called := make(chan struct{}, 1)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, n *notifications.Notification) error {
				assert.Equal(t, int64(42), n.UserID)
				assert.Equal(t, notifications.TypeFlagged, n.Type)
				called <- struct{}{}
				return nil
			})

		svc := notifications.NewService(mockRepo, logger)
		svc.Notify(context.Background(), notifications.Notification{
			UserID: 42,
			Type:   notifications.TypeFlagged,
			Title:  "flagged",
		})

		select {
		case <-called:
		case <-time.After(2 * time.Second):
			t.Fatal("Create was not called within timeout")
		}
	})

	t.Run("drops self-notifications where actor == recipient", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := notifications.NewMockRepository(ctrl)
		// No EXPECT() — Create must not be called.
		svc := notifications.NewService(mockRepo, logger)

		actor := int64(7)
		svc.Notify(context.Background(), notifications.Notification{
			UserID:      7,
			ActorUserID: &actor,
			Type:        notifications.TypeAssigned,
			Title:       "self",
		})

		// Give any (unwanted) goroutine time to run.
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("drops rows with zero recipient", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := notifications.NewMockRepository(ctrl)
		svc := notifications.NewService(mockRepo, logger)

		svc.Notify(context.Background(), notifications.Notification{
			Type:  notifications.TypeAssigned,
			Title: "no recipient",
		})

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("errors from Create are logged, not returned", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := notifications.NewMockRepository(ctrl)
		called := make(chan struct{}, 1)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *notifications.Notification) error {
				called <- struct{}{}
				return errors.New("boom")
			})

		svc := notifications.NewService(mockRepo, logger)
		require.NotPanics(t, func() {
			svc.Notify(context.Background(), notifications.Notification{
				UserID: 1,
				Type:   notifications.TypeAssigned,
				Title:  "t",
			})
		})
		select {
		case <-called:
		case <-time.After(2 * time.Second):
			t.Fatal("Create was not called within timeout")
		}
	})
}
