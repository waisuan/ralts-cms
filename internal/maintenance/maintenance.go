// Package maintenance provides maintenance record management functionality
// for tracking machine maintenance activities in the Ralts-CMS application.
package maintenance

import "time"

// Maintenance represents a maintenance record in the system
type Maintenance struct {
	ID                  int       `json:"id" db:"id"`
	MachineSerialNumber string    `json:"machine_serial_number" db:"machine_serial_number"`
	WorkOrderNumber     string    `json:"work_order_number" db:"work_order_number"`
	WorkOrderDate       time.Time `json:"work_order_date" db:"work_order_date"`
	ActionTaken         string    `json:"action_taken" db:"action_taken"`
	ReportedBy          string    `json:"reported_by" db:"reported_by"`
	WorkOrderType       string    `json:"work_order_type" db:"work_order_type"`
	Attachment          *string   `json:"attachment" db:"attachment"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Maintenance) SetTimestamps() {
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
