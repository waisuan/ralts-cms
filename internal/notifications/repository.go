package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for notifications.
//
//go:generate mockgen -destination=mock_repository.go -package=notifications -source=repository.go
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	ListForUser(ctx context.Context, userID int64, opts ListOptions) ([]*Notification, error)
	CountForUser(ctx context.Context, userID int64, opts ListOptions) (int, error)
	CountUnread(ctx context.Context, userID int64) (int, error)
	MarkRead(ctx context.Context, userID int64, id string) error
	MarkAllRead(ctx context.Context, userID int64) (int, error)
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository returns a Postgres-backed notifications repository.
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{client: client}
}

func (r *db) Create(ctx context.Context, n *Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO notifications (
			id, user_id, type, machine_serial_number, flag_id,
			title, body, actor_user_id, read_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.client.Exec(ctx, query,
		n.ID, n.UserID, string(n.Type), n.MachineSerialNumber, n.FlagID,
		n.Title, n.Body, n.ActorUserID, n.ReadAt, n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *db) ListForUser(ctx context.Context, userID int64, opts ListOptions) ([]*Notification, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	// machine_flags is joined so the inbox knows whether a flag notification is
	// still actionable without a follow-up request per row.
	query := `
		SELECT n.id, n.user_id, n.type, n.machine_serial_number, n.flag_id,
		       f.status, n.title, COALESCE(n.body, ''), n.actor_user_id, u.username,
		       n.read_at, n.created_at
		FROM notifications n
		LEFT JOIN users u ON u.id = n.actor_user_id
		LEFT JOIN machine_flags f ON f.id = n.flag_id
		WHERE n.user_id = $1
	`
	args := []any{userID}

	if opts.UnreadOnly {
		query += " AND n.read_at IS NULL"
	}

	query += " ORDER BY n.created_at DESC LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var results []*Notification
	for rows.Next() {
		var n Notification
		var typeStr string
		if err := rows.Scan(
			&n.ID, &n.UserID, &typeStr, &n.MachineSerialNumber, &n.FlagID,
			&n.FlagStatus, &n.Title, &n.Body, &n.ActorUserID, &n.ActorUsername,
			&n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		n.Type = Type(typeStr)
		results = append(results, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}
	return results, nil
}

func (r *db) CountForUser(ctx context.Context, userID int64, opts ListOptions) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	args := []any{userID}
	if opts.UnreadOnly {
		query += " AND read_at IS NULL"
	}

	var count int
	if err := r.client.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}
	return count, nil
}

func (r *db) CountUnread(ctx context.Context, userID int64) (int, error) {
	return r.CountForUser(ctx, userID, ListOptions{UnreadOnly: true})
}

func (r *db) MarkRead(ctx context.Context, userID int64, id string) error {
	query := `
		UPDATE notifications
		SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`
	if _, err := r.client.Exec(ctx, query, id, userID); err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}
	return nil
}

func (r *db) MarkAllRead(ctx context.Context, userID int64) (int, error) {
	query := `
		UPDATE notifications
		SET read_at = NOW()
		WHERE user_id = $1 AND read_at IS NULL
	`
	result, err := r.client.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications read: %w", err)
	}
	return int(result.RowsAffected()), nil
}
