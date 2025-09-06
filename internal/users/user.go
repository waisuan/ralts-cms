package users

import (
	"fmt"
	"ralts-cms/pkg/auth"
	"time"
)

// User status constants
const (
	StatusPendingApproval = "pending_approval"
	StatusApproved        = "approved"
	StatusSuspended       = "suspended"
	StatusInactive        = "inactive"
)

// User role constants
const (
	RoleAdmin    = "ADMIN"
	RoleNonAdmin = "NON_ADMIN"
)

// User represents a user in the Ralts-CMS system
type User struct {
	ID        int64      `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	Salt      string     `json:"-" db:"salt"`
	Role      string     `json:"role" db:"role"`
	Approved  bool       `json:"approved" db:"approved"`
	Status    *string    `json:"status,omitempty" db:"status"`
	Avatar    *string    `json:"avatar,omitempty" db:"avatar"`
	LastLogin *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
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

	u.UpdatedAt = &now
}

// SetDefaultRole sets the default role if not already set
func (u *User) SetDefaultRole() {
	if u.Role == "" {
		u.Role = RoleNonAdmin
	}
}

// SetDefaultStatus sets the default status if not already set
// New users get "pending_approval" status by default
func (u *User) SetDefaultStatus() {
	if u.Status == nil {
		status := StatusPendingApproval
		u.Status = &status
	}
}

// SetDefaultApproval sets the default approval status if not already set
// New users are not approved by default, but we don't override explicitly set values
func (u *User) SetDefaultApproval() {
	// Only set to false for new users if not explicitly set to true
	// This ensures newly registered users require approval by default
	// but allows explicit approval during creation
	if u.CreatedAt.IsZero() && !u.Approved {
		u.Approved = false
	}
}

// SetDefaults sets all default values for optional fields
func (u *User) SetDefaults() {
	u.SetDefaultRole()
	u.SetDefaultStatus()
	u.SetDefaultApproval()
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

// IsStatusActive checks if the user's status allows login
// Returns true if the user has an approved status or legacy active status
func (u *User) IsStatusActive() bool {
	if u.Status == nil {
		return false
	}

	return *u.Status == StatusApproved
}

// ValidateStatusValue validates that the status is one of the allowed constants
func ValidateStatusValue(status string) error {
	validStatuses := []string{StatusPendingApproval, StatusApproved, StatusSuspended, StatusInactive}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}
	return fmt.Errorf("invalid status: %s. Must be one of: %v", status, validStatuses)
}
