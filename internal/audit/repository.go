package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for audit log data access operations
//
//go:generate mockgen -destination=mock_audit_repository.go -package=audit -source=repository.go
type Repository interface {
	Create(ctx context.Context, event *Event) error
	List(ctx context.Context, options *ListOptions) ([]*Event, error)
	Count(ctx context.Context, options *ListOptions) (int32, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// ListOptions defines parameters for listing audit events
type ListOptions struct {
	Limit        int32   `json:"limit"`
	Offset       int32   `json:"offset"`
	UserID       *string `json:"user_id,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Action       *string `json:"action,omitempty"`
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository creates a new audit repository instance with the given database connection
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{
		client: client,
	}
}

// Create inserts a new audit event into the database
func (r *db) Create(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource_type, resource_id, details, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	var detailsJSON []byte
	var err error
	if event.Details != nil {
		detailsJSON, err = json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	// Convert string UserID to int64 for database storage
	var userID *int64
	if event.UserID != nil && *event.UserID != "" {
		parsed, err := strconv.ParseInt(*event.UserID, 10, 64)
		if err == nil {
			userID = &parsed
		}
	}

	_, err = r.client.Exec(ctx, query,
		event.ID,
		userID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		detailsJSON,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// List retrieves audit events based on the provided options
func (r *db) List(ctx context.Context, options *ListOptions) ([]*Event, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if options.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *options.UserID)
		argIndex++
	}

	if options.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *options.ResourceType)
		argIndex++
	}

	if options.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *options.ResourceID)
		argIndex++
	}

	if options.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *options.Action)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, options.Limit, options.Offset)

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		var detailsJSON []byte
		var userID *int64

		err := rows.Scan(
			&event.ID,
			&userID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&detailsJSON,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Convert int64 userID back to string for the Event struct
		if userID != nil {
			userIDStr := strconv.FormatInt(*userID, 10)
			event.UserID = &userIDStr
		}

		if detailsJSON != nil {
			if err := json.Unmarshal(detailsJSON, &event.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return events, nil
}

// Count returns the total number of audit events matching the provided options
func (r *db) Count(ctx context.Context, options *ListOptions) (int32, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Apply filters (same as List)
	if options.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *options.UserID)
		argIndex++
	}

	if options.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *options.ResourceType)
		argIndex++
	}

	if options.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *options.ResourceID)
		argIndex++
	}

	if options.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *options.Action)
	}

	var count int32
	err := r.client.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes audit events older than the specified cutoff time.
// Returns the number of deleted events.
func (r *db) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := r.client.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return result.RowsAffected(), nil
}
