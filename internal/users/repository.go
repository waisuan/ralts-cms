package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, user *User) error
	// GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByID(ctx context.Context, id int) (*User, error)
	// Update(ctx context.Context, user *User) error
	// Delete(ctx context.Context, id int) error
}

type db struct {
	client *pgxpool.Pool
}

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
