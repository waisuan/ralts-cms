package users

import (
	"fmt"
	"ralts-cms/pkg/auth"
	"time"
)

type User struct {
	ID        int        `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	Salt      string     `json:"-" db:"salt"`
	Role      string     `json:"role" db:"role"`
	Status    string     `json:"status" db:"status"`
	Avatar    *string    `json:"avatar,omitempty" db:"avatar"`
	LastLogin *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
// If CreatedAt is zero (new user), sets both CreatedAt and UpdatedAt to current time
// If CreatedAt is not zero (existing user), only updates UpdatedAt
func (u *User) SetTimestamps() {
	now := time.Now().UTC()

	if u.CreatedAt.IsZero() {
		// New user - set both timestamps
		u.CreatedAt = now
	}

	u.UpdatedAt = now
}

// SetDefaultRole sets the default role if not already set
func (u *User) SetDefaultRole() {
	if u.Role == "" {
		u.Role = "user"
	}
}

// SetDefaultStatus sets the default status if not already set
func (u *User) SetDefaultStatus() {
	if u.Status == "" {
		u.Status = "active"
	}
}

// SetDefaults sets all default values for optional fields
func (u *User) SetDefaults() {
	u.SetDefaultRole()
	u.SetDefaultStatus()
}

// SetPassword hashes the provided plain text password and sets the salt
// This method should be called with the plain text password before saving to database
func (u *User) SetPassword(plainPassword string) error {
	if plainPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Generate salt
	salt, err := auth.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with salt
	hashedPassword, err := auth.HashPassword(plainPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Set the hashed password and salt
	u.Password = hashedPassword
	u.Salt = salt

	return nil
}
