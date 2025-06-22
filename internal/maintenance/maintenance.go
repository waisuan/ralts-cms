package maintenance

import "time"

// Maintenance represents a maintenance record in the system
type Maintenance struct {
	MachineSerialNumber string `json:"machine_serial_number" dynamodbav:"MachineSerialNumber"`
	WorkOrderNumber     string `json:"work_order_number" dynamodbav:"WorkOrderNumber"`
	WorkOrderDate       string `json:"work_order_date" dynamodbav:"WorkOrderDate"` // ISO 8601 format
	ActionTaken         string `json:"action_taken" dynamodbav:"ActionTaken"`
	ReportedBy          string `json:"reported_by" dynamodbav:"ReportedBy"`
	WorkerOrderType     string `json:"worker_order_type" dynamodbav:"WorkerOrderType"`
	Attachment          string `json:"attachment" dynamodbav:"Attachment"`
	CreatedAt           string `json:"created_at" dynamodbav:"CreatedAt"` // ISO 8601 format
	UpdatedAt           string `json:"updated_at" dynamodbav:"UpdatedAt"` // ISO 8601 format
}

// GetPartitionKey returns the partition key for DynamoDB
func (m *Maintenance) GetPartitionKey() string {
	return "Machine#" + m.MachineSerialNumber
}

// GetSortKey returns the sort key for DynamoDB
func (m *Maintenance) GetSortKey() string {
	return "Maintenance#" + m.WorkOrderNumber
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Maintenance) SetTimestamps() {
	now := time.Now().UTC().Format(time.RFC3339)
	if m.CreatedAt == "" {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
