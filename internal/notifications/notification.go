// Package notifications provides an in-app inbox for users.
//
// A notification row represents a single event delivered to a recipient user.
// Rows are created by the machine handlers (on assignment changes) and by the
// flags service (on flag creation and on resolution by a non-admin) and are
// read via the /api/v1/notifications endpoints backing the frontend bell/inbox
// UI.
package notifications

import "time"

// Type identifies why a notification was created.
type Type string

const (
	// TypeAssigned indicates that the recipient became the assignee of a machine.
	TypeAssigned Type = "assigned"
	// TypeFlagged indicates that a machine assigned to the recipient was flagged.
	TypeFlagged Type = "flagged"
	// TypeFlagResolved indicates that a flag the recipient raised was resolved
	// by somebody else.
	TypeFlagResolved Type = "flag_resolved"
)

// Notification is a single row in the recipient's inbox.
type Notification struct {
	ID string `json:"id"`
	// UserID is the recipient user id. Notifications are strictly scoped to this
	// user by the repository/handlers.
	UserID              int64   `json:"user_id"`
	Type                Type    `json:"type"`
	MachineSerialNumber *string `json:"machine_serial_number,omitempty"`
	FlagID              *string `json:"flag_id,omitempty"`
	// FlagStatus is the current status of the referenced flag (open/resolved).
	// Read-only: populated when listing so the inbox can offer a resolve action
	// only while the flag is still open. Nil for non-flag notifications.
	FlagStatus    *string    `json:"flag_status,omitempty"`
	Title         string     `json:"title"`
	Body          string     `json:"body,omitempty"`
	ActorUserID   *int64     `json:"actor_user_id,omitempty"`
	ActorUsername *string    `json:"actor_username,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ListOptions parameterises listing notifications for a user.
type ListOptions struct {
	Limit      int32
	Offset     int32
	UnreadOnly bool
}
