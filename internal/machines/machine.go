// Package machines provides machine management functionality for tracking
// equipment and maintenance schedules in the Ralts-CMS application.
package machines

import (
	"time"
)

// AssignedUser is a minimal projection of the user assigned to a machine.
// It is populated by joining machines against users so callers can show
// assignee username/email without a second lookup.
type AssignedUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Machine represents a machine record in the system
type Machine struct {
	ID              int64     `json:"id" db:"id"`
	SerialNumber    string    `json:"serial_number" db:"serialNumber"`
	Customer        string    `json:"customer" db:"customer"`
	State           string    `json:"state" db:"state"`
	AccountType     string    `json:"account_type" db:"accountType"`
	Model           string    `json:"model" db:"model"`
	Status          string    `json:"status" db:"status"`
	Brand           string    `json:"brand" db:"brand"`
	District        string    `json:"district" db:"district"`
	PersonInCharge  string    `json:"person_in_charge" db:"personInCharge"`
	ReportedBy      string    `json:"reported_by" db:"reportedBy"`
	AdditionalNotes string    `json:"additional_notes" db:"additionalNotes"`
	Attachment      string    `json:"attachment" db:"attachment"`
	TncDate         time.Time `json:"tnc_date" db:"tncDate"`
	PpmDate         time.Time `json:"ppm_date" db:"ppmDate"`
	CreatedAt       time.Time `json:"created_at" db:"createdAt"`
	UpdatedAt       time.Time `json:"updated_at" db:"updatedAt"`
	UpdatedBy       string    `json:"updated_by" db:"updatedBy"`
	// AssignedUserID is the FK to the user this machine is assigned to.
	// Nullable: legacy rows and unassigned machines have no assignee.
	AssignedUserID   *int64        `json:"assigned_user_id,omitempty" db:"assignedUserId"`
	AssignedUser     *AssignedUser `json:"assigned_user,omitempty"`
	PpmStatus        string        `json:"ppm_status"`
	MaintenanceCount int           `json:"maintenance_count"`
}

// IsPpmDateUnset reports whether ppmDate is a zero/sentinel value that must not receive a computed PPM status.
// Matches client logic: UTC calendar year ≤ 1 (Go zero time and legacy DB dates like 0001-01-01 / 0001-12-31).
func IsPpmDateUnset(t time.Time) bool {
	y, _, _ := t.UTC().Date()
	return y <= 1
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Machine) SetTimestamps() {
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
