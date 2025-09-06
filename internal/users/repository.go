// Package users provides user management functionality including authentication,
// user creation, and database operations for the Ralts-CMS application.
package users

import (
	"context"
	"fmt"
	"ralts-cms/pkg/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for user data access operations
//
//go:generate mockgen -destination=../users/mock_users_repository.go -package=users -source=repository.go
type Repository interface {
	Create(ctx context.Context, user *User) error
	Login(ctx context.Context, username string, password string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateStatus(ctx context.Context, userID int64, status string) error
	UpdateMultipleStatuses(ctx context.Context, userIDs []int64, status string) error
}

type db struct {
	client *pgxpool.Pool
}

// NewRepository creates a new user repository instance with the given database connection
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{
		client: client,
	}
}

func (r *db) Create(ctx context.Context, user *User) error {
	// Validate required fields
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Set default values for optional fields
	user.SetDefaults()

	// Set timestamps
	user.SetTimestamps()

	// Hash password and set salt
	if err := user.SetPassword(user.Password); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO users (
			username, email, password, salt, role, approved, status, avatar,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		user.Username, user.Email, user.Password, user.Salt, user.Role, user.Approved, user.Status, user.Avatar,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *db) Login(ctx context.Context, username string, password string) (*User, error) {
	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if err := auth.VerifyPassword(password, user.Password, user.Salt); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Check if user is approved
	if !user.Approved {
		return nil, fmt.Errorf("account pending approval: your account is awaiting administrator approval")
	}

	// Check if user status allows login
	if !user.IsStatusActive() {
		return nil, fmt.Errorf("account not active: your account status does not allow login")
	}

	return user, nil
}

func (r *db) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password, salt, role, approved, status, avatar, last_login, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user User
	err := r.client.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Approved, &user.Status, &user.Avatar, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *db) ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error) {
	// Validate pagination parameters
	if limit <= 0 || limit > 100 {
		return nil, 0, fmt.Errorf("limit must be between 1 and 100")
	}
	if offset < 0 {
		return nil, 0, fmt.Errorf("offset must be non-negative")
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var totalCount int
	err := r.client.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users with pagination
	query := `
		SELECT id, username, email, password, salt, role, approved, status, avatar, last_login, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.client.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Approved, &user.Status, &user.Avatar, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, totalCount, nil
}

func (r *db) GetByID(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("user ID must be positive")
	}

	query := `
		SELECT id, username, email, password, salt, role, approved, status, avatar, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.client.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Approved, &user.Status, &user.Avatar, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// buildStatusUpdateQuery builds the appropriate SQL query and arguments for status updates
func buildStatusUpdateQuery(status string, userID int64) (string, []interface{}) {
	if status == StatusApproved {
		query := `UPDATE users SET status = $1, updated_at = NOW(), approved = true WHERE id = $2`
		args := []interface{}{status, userID}
		return query, args
	}

	query := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`
	args := []interface{}{status, userID}
	return query, args
}

func (r *db) UpdateStatus(ctx context.Context, userID int64, status string) error {
	if err := ValidateStatusValue(status); err != nil {
		return err
	}

	query, args := buildStatusUpdateQuery(status, userID)

	result, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	return nil
}

func (r *db) UpdateMultipleStatuses(ctx context.Context, userIDs []int64, status string) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("userIDs cannot be empty")
	}

	// Validate all user IDs upfront
	for _, userID := range userIDs {
		if userID <= 0 {
			return fmt.Errorf("all user IDs must be positive")
		}
	}

	// Validate status using domain validation
	if err := ValidateStatusValue(status); err != nil {
		return err
	}

	// Use a transaction for bulk update
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Reuse the query building logic
	query, baseArgs := buildStatusUpdateQuery(status, 0) // userID will be replaced per iteration

	var totalRowsAffected int64
	for _, userID := range userIDs {
		// Replace the userID in the arguments (last argument is always userID)
		args := make([]interface{}, len(baseArgs))
		copy(args, baseArgs)
		args[len(args)-1] = userID

		result, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update user status for ID %d: %w", userID, err)
		}
		totalRowsAffected += result.RowsAffected()
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if totalRowsAffected != int64(len(userIDs)) {
		return fmt.Errorf("expected to update %d users, but updated %d", len(userIDs), totalRowsAffected)
	}

	return nil
}
