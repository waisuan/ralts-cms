// Package users provides user management functionality including authentication,
// user creation, and database operations for the Ralts-CMS application.
package users

import (
	"context"
	"fmt"
	"ralts-cms/pkg/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -destination=../users/mock_users_repository.go -package=users -source=repository.go
// Repository defines the interface for user data access operations
type Repository interface {
	Create(ctx context.Context, user *User) error
	Login(ctx context.Context, email string, password string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByID(ctx context.Context, id int) (*User, error)
	// Update(ctx context.Context, user *User) error
	// Delete(ctx context.Context, id int) error
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
	if user.Name == "" {
		return fmt.Errorf("name is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Set default values for optional fields
	user.SetDefaults()

	// Hash password and set salt
	if err := user.SetPassword(user.Password); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	// Set timestamps
	user.SetTimestamps()

	// Insert into database
	query := `
		INSERT INTO users (
			name, email, password, salt, role, status, avatar,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id
	`

	err := r.client.QueryRow(ctx, query,
		user.Name, user.Email, user.Password, user.Salt, user.Role, user.Status, user.Avatar,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *db) Login(ctx context.Context, email string, password string) (*User, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err := auth.VerifyPassword(password, user.Password, user.Salt); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return user, nil
}

func (r *db) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, email, password, salt, role, status, avatar, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.client.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Salt, &user.Role, &user.Status, &user.Avatar, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}
