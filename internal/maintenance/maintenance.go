// Package maintenance provides maintenance record management functionality
// for tracking machine maintenance activities in the Ralts-CMS application.
package maintenance

import "time"

// Maintenance represents a maintenance record in the system
type Maintenance struct {
	ID                  int64     `json:"id" db:"id"`
	MachineSerialNumber string    `json:"machine_serial_number" db:"serialNumber"`
	WorkOrderNumber     string    `json:"work_order_number" db:"workOrderNumber"`
	WorkOrderDate       time.Time `json:"work_order_date" db:"workOrderDate"`
	ActionTaken         string    `json:"action_taken" db:"actionTaken"`
	ReportedBy          string    `json:"reported_by" db:"reportedBy"`
	WorkOrderType       string    `json:"work_order_type" db:"workOrderType"`
	Attachment          *string   `json:"attachment" db:"attachment"`
	CreatedAt           time.Time `json:"created_at" db:"createdAt"`
	UpdatedAt           time.Time `json:"updated_at" db:"updatedAt"`
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Maintenance) SetTimestamps() {
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
