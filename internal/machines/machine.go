package machines

import (
	"time"
)

// Machine represents a machine record in the system
type Machine struct {
	ID               int       `json:"id" db:"id"`
	SerialNumber     string    `json:"serial_number" db:"serial_number"`
	Customer         string    `json:"customer" db:"customer"`
	State            string    `json:"state" db:"state"`
	AccountType      string    `json:"account_type" db:"account_type"`
	Model            string    `json:"model" db:"model"`
	Status           string    `json:"status" db:"status"`
	Brand            string    `json:"brand" db:"brand"`
	District         string    `json:"district" db:"district"`
	PersonInCharge   string    `json:"person_in_charge" db:"person_in_charge"`
	ReportedBy       string    `json:"reported_by" db:"reported_by"`
	AdditionalNotes  string    `json:"additional_notes" db:"additional_notes"`
	Attachment       string    `json:"attachment" db:"attachment"`
	TncDate          time.Time `json:"tnc_date" db:"tnc_date"`
	PpmDate          time.Time `json:"ppm_date" db:"ppm_date"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	PpmStatus        string    `json:"ppm_status"`
	MaintenanceCount int       `json:"maintenance_count"`
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Machine) SetTimestamps() {
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
