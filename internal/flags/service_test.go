package flags_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"ralts-cms/internal/flags"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/notifications"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func int64Ptr(v int64) *int64 { return &v }

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCreateManualFlag(t *testing.T) {
	t.Run("rejects requires_attention as a manual reason", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		_, _, err := svc.CreateManualFlag(context.Background(), 1, "SN-1", flags.ReasonRequiresAttention, "")
		require.Error(t, err)
	})

	t.Run("requires a note when reason is other", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())

		// Whitespace does not stand in for a note.
		for _, note := range []string{"", "   \n\t "} {
			_, _, err := svc.CreateManualFlag(context.Background(), 1, "SN-1", flags.ReasonOther, note)
			require.ErrorIs(t, err, flags.ErrNoteRequired)
		}
	})

	t.Run("rejects a note longer than the limit, for either reason", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())

		for _, reason := range []flags.Reason{flags.ReasonMissingValues, flags.ReasonOther} {
			tooLong := strings.Repeat("a", flags.MaxNoteLength+1)
			_, _, err := svc.CreateManualFlag(context.Background(), 1, "SN-1", reason, tooLong)
			require.ErrorIs(t, err, flags.ErrNoteTooLong)
		}
	})

	t.Run("counts characters rather than bytes, so multi-byte notes fit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		serial := "SN-UTF8"
		// Three bytes per rune: over the limit in bytes, at it in characters.
		note := strings.Repeat("字", flags.MaxNoteLength)
		mrepo.EXPECT().
			GetBySerialNumber(gomock.Any(), serial).
			Return(&machines.Machine{SerialNumber: serial}, nil)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(true, nil)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		created, _, err := svc.CreateManualFlag(context.Background(), 1, serial, flags.ReasonOther, note)
		require.NoError(t, err)
		assert.Equal(t, note, created.Note)
	})

	t.Run("stores the trimmed note", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		serial := "SN-TRIM"
		mrepo.EXPECT().
			GetBySerialNumber(gomock.Any(), serial).
			Return(&machines.Machine{SerialNumber: serial}, nil)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(true, nil)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		created, _, err := svc.CreateManualFlag(context.Background(), 1, serial, flags.ReasonOther, "  check cabling  ")
		require.NoError(t, err)
		assert.Equal(t, "check cabling", created.Note)
	})

	t.Run("creates flag and notifies assignee (skipping when actor==assignee)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		serial := "SN-42"
		actor := int64(9)
		assignee := int64(7)

		mrepo.EXPECT().
			GetBySerialNumber(gomock.Any(), serial).
			Return(&machines.Machine{
				SerialNumber:   serial,
				AssignedUserID: &assignee,
			}, nil)

		repo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, f *flags.Flag) (bool, error) {
				assert.Equal(t, serial, f.MachineSerialNumber)
				assert.Equal(t, flags.ReasonMissingValues, f.Reason)
				require.NotNil(t, f.CreatedBy)
				assert.Equal(t, actor, *f.CreatedBy)
				f.ID = "generated-id"
				return true, nil
			})

		notifSvc.EXPECT().
			Notify(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, n notifications.Notification) {
				assert.Equal(t, assignee, n.UserID)
				assert.Equal(t, notifications.TypeFlagged, n.Type)
				require.NotNil(t, n.ActorUserID)
				assert.Equal(t, actor, *n.ActorUserID)
			})

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		f, isNew, err := svc.CreateManualFlag(context.Background(), actor, serial, flags.ReasonMissingValues, "")
		require.NoError(t, err)
		require.NotNil(t, f)
		assert.True(t, isNew)
	})

	t.Run("replacing a machine's open flag notifies the assignee again", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		serial := "SN-REFLAG"
		assignee := int64(7)

		mrepo.EXPECT().
			GetBySerialNumber(gomock.Any(), serial).
			Return(&machines.Machine{SerialNumber: serial, AssignedUserID: &assignee}, nil)
		// The repository reports that it rewrote the existing open flag.
		repo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, f *flags.Flag) (bool, error) {
				f.ID = "existing-id"
				return false, nil
			})
		notifSvc.EXPECT().
			Notify(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, n notifications.Notification) {
				assert.Equal(t, assignee, n.UserID)
				require.NotNil(t, n.FlagID)
				assert.Equal(t, "existing-id", *n.FlagID)
			})

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		f, isNew, err := svc.CreateManualFlag(context.Background(), 9, serial, flags.ReasonOther, "still wrong")
		require.NoError(t, err)
		assert.False(t, isNew, "the caller can tell an update from a fresh flag")
		assert.Equal(t, "existing-id", f.ID)
	})

	t.Run("does not notify when machine has no assignee", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		mrepo.EXPECT().
			GetBySerialNumber(gomock.Any(), "SN-N").
			Return(&machines.Machine{SerialNumber: "SN-N"}, nil)
		repo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(true, nil)
		// notifSvc.Notify must NOT be called.

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		_, _, err := svc.CreateManualFlag(context.Background(), 5, "SN-N", flags.ReasonOther, "reason note")
		require.NoError(t, err)
	})
}

func TestResolve(t *testing.T) {
	newService := func(t *testing.T) (*flags.MockRepository, *notifications.MockService, flags.Service) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)
		return repo, notifSvc, flags.NewService(repo, mrepo, notifSvc, discardLogger())
	}

	t.Run("resolves without a note", func(t *testing.T) {
		repo, _, svc := newService(t)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(3), "").Return(nil)

		require.NoError(t, svc.Resolve(context.Background(), 3, true, "flag-1", ""))
	})

	t.Run("passes the trimmed resolution note through", func(t *testing.T) {
		repo, _, svc := newService(t)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(3), "replaced the sensor").Return(nil)

		require.NoError(t, svc.Resolve(context.Background(), 3, true, "flag-1", "  replaced the sensor \n"))
	})

	t.Run("treats a whitespace-only note as no note", func(t *testing.T) {
		repo, _, svc := newService(t)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(3), "").Return(nil)

		require.NoError(t, svc.Resolve(context.Background(), 3, true, "flag-1", "   "))
	})

	t.Run("rejects a resolution note over the limit without touching the repository", func(t *testing.T) {
		_, _, svc := newService(t)

		err := svc.Resolve(context.Background(), 3, true, "flag-1", strings.Repeat("a", flags.MaxNoteLength+1))
		require.ErrorIs(t, err, flags.ErrNoteTooLong)
	})

	t.Run("an admin resolution notifies nobody", func(t *testing.T) {
		repo, _, svc := newService(t)
		// No GetByID either: the raiser is irrelevant, so the flag is not read.
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(3), "").Return(nil)

		require.NoError(t, svc.Resolve(context.Background(), 3, true, "flag-1", ""))
	})

	t.Run("tells the raising admin when an assignee resolves the flag", func(t *testing.T) {
		repo, notifSvc, svc := newService(t)
		repo.EXPECT().GetByID(gomock.Any(), "flag-1").Return(&flags.Flag{
			ID:                  "flag-1",
			MachineSerialNumber: "SN-1",
			CreatedBy:           int64Ptr(9),
		}, nil)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(7), "swapped the sensor").Return(nil)
		notifSvc.EXPECT().Notify(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, n notifications.Notification) {
				assert.Equal(t, int64(9), n.UserID)
				assert.Equal(t, notifications.TypeFlagResolved, n.Type)
				assert.Equal(t, "Machine SN-1: flag resolved", n.Title)
				assert.Equal(t, "swapped the sensor", n.Body)
				require.NotNil(t, n.ActorUserID)
				assert.Equal(t, int64(7), *n.ActorUserID)
				require.NotNil(t, n.FlagID)
				assert.Equal(t, "flag-1", *n.FlagID)
			})

		require.NoError(t, svc.Resolve(context.Background(), 7, false, "flag-1", "swapped the sensor"))
	})

	t.Run("skips the notification for a flag with no known raiser", func(t *testing.T) {
		repo, _, svc := newService(t)
		repo.EXPECT().GetByID(gomock.Any(), "flag-1").
			Return(&flags.Flag{ID: "flag-1", MachineSerialNumber: "SN-1"}, nil)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(7), "").Return(nil)

		require.NoError(t, svc.Resolve(context.Background(), 7, false, "flag-1", ""))
	})

	t.Run("does not notify when the resolve itself fails", func(t *testing.T) {
		repo, _, svc := newService(t)
		repo.EXPECT().GetByID(gomock.Any(), "flag-1").Return(&flags.Flag{
			ID:                  "flag-1",
			MachineSerialNumber: "SN-1",
			CreatedBy:           int64Ptr(9),
		}, nil)
		repo.EXPECT().Resolve(gomock.Any(), "flag-1", int64(7), "").Return(flags.ErrNotFound)

		require.ErrorIs(t, svc.Resolve(context.Background(), 7, false, "flag-1", ""), flags.ErrNotFound)
	})
}

func TestCanResolve(t *testing.T) {
	t.Run("admins may resolve any flag without loading it", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		allowed, err := svc.CanResolve(context.Background(), 1, true, "flag-1")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("assignee may resolve flags on their own machine", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		repo.EXPECT().GetByID(gomock.Any(), "flag-1").
			Return(&flags.Flag{ID: "flag-1", MachineSerialNumber: "SN-1"}, nil)
		mrepo.EXPECT().GetBySerialNumber(gomock.Any(), "SN-1").
			Return(&machines.Machine{SerialNumber: "SN-1", AssignedUserID: int64Ptr(7)}, nil)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		allowed, err := svc.CanResolve(context.Background(), 7, false, "flag-1")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("non-assignee non-admin may not resolve", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		repo.EXPECT().GetByID(gomock.Any(), "flag-1").
			Return(&flags.Flag{ID: "flag-1", MachineSerialNumber: "SN-1"}, nil)
		mrepo.EXPECT().GetBySerialNumber(gomock.Any(), "SN-1").
			Return(&machines.Machine{SerialNumber: "SN-1", AssignedUserID: int64Ptr(7)}, nil)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		allowed, err := svc.CanResolve(context.Background(), 8, false, "flag-1")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("free-text assignee means only admins can resolve", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		repo.EXPECT().GetByID(gomock.Any(), "flag-2").
			Return(&flags.Flag{ID: "flag-2", MachineSerialNumber: "SN-G"}, nil)
		mrepo.EXPECT().GetBySerialNumber(gomock.Any(), "SN-G").
			Return(&machines.Machine{SerialNumber: "SN-G", PersonInCharge: "Not A User"}, nil)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		allowed, err := svc.CanResolve(context.Background(), 7, false, "flag-2")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("propagates ErrNotFound for unknown flags", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := flags.NewMockRepository(ctrl)
		mrepo := machines.NewMockRepository(ctrl)
		notifSvc := notifications.NewMockService(ctrl)

		repo.EXPECT().GetByID(gomock.Any(), "nope").Return(nil, flags.ErrNotFound)

		svc := flags.NewService(repo, mrepo, notifSvc, discardLogger())
		_, err := svc.CanResolve(context.Background(), 7, false, "nope")
		require.ErrorIs(t, err, flags.ErrNotFound)
	})
}
