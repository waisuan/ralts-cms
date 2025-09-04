// Package users provides user management functionality including authentication,
// user creation, and database operations for the Ralts-CMS application.
package users

import (
	"context"
	"fmt"
	"ralts-cms/pkg/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for user data access operations
//
//go:generate mockgen -destination=../users/mock_users_repository.go -package=users -source=repository.go
type Repository interface {
	Create(ctx context.Context, user *User) error
	Login(ctx context.Context, username string, password string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
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
