package audit

import (
	"net/http"
	"time"

	appctx "ralts-cms/internal/context"
)

// Action constants define the types of actions that can be audited
const (
	ActionCreated         = "created"
	ActionUpdated         = "updated"
	ActionDeleted         = "deleted"
	ActionViewed          = "viewed"
	ActionListed          = "listed"
	ActionLogin           = "login"
	ActionLogout          = "logout"
	ActionPasswordChanged = "password_changed"
)

// Resource type constants define the types of resources that can be audited
const (
	ResourceMachine     = "machine"
	ResourceMaintenance = "maintenance"
	ResourceUser        = "user"
	ResourceSession     = "session"
	ResourceAttachment  = "attachment"
)

// Event represents an audit log entry
type Event struct {
	ID           string         `json:"id"`
	UserID       *string        `json:"user_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Details      map[string]any `json:"details,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// NewEvent creates a new audit event with common fields populated from the request.
// This is a convenience constructor to reduce boilerplate in handlers.
func NewEvent(r *http.Request, action, resourceType, resourceID string, details map[string]any) Event {
	return Event{
		UserID:       appctx.GetUserIDFromContext(r.Context()),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
}
