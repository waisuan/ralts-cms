package notifications

import (
	"context"
	"log/slog"
)

// Service is the interface used by callers that want to emit notifications
// without caring about persistence. Notifications are best-effort: failures
// are logged but never returned so the caller's main flow is not blocked.
//
//go:generate mockgen -destination=mock_service.go -package=notifications -source=service.go
type Service interface {
	// Notify persists a notification for the recipient set on the Notification.
	// The call is asynchronous — errors are logged, not returned. Notifications
	// where ActorUserID equals the recipient UserID are dropped so users are
	// not notified about their own actions.
	Notify(ctx context.Context, n Notification)
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService returns a Service backed by the given repository.
func NewService(repo Repository, logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{repo: repo, logger: logger}
}

func (s *service) Notify(ctx context.Context, n Notification) {
	if n.UserID == 0 {
		return
	}
	// Skip self-notifications: the actor doesn't need to hear about their own action.
	if n.ActorUserID != nil && *n.ActorUserID == n.UserID {
		return
	}

	// Detach the parent context so the write survives request cancellation but
	// still uses a fresh background context. Notifications are important enough
	// that we don't want them to be dropped when a client disconnects.
	go func(n Notification) {
		if err := s.repo.Create(context.Background(), &n); err != nil {
			s.logger.Error("failed to persist notification",
				"error", err,
				"user_id", n.UserID,
				"type", n.Type,
			)
		}
	}(n)
}
