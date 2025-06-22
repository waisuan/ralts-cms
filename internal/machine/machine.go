package machine

import "time"

// Machine represents a machine record in the system
type Machine struct {
	SerialNumber    string `json:"serial_number" dynamodbav:"SerialNumber"`
	Customer        string `json:"customer" dynamodbav:"Customer"`
	State           string `json:"state" dynamodbav:"State"`
	AccountType     string `json:"account_type" dynamodbav:"AccountType"`
	Model           string `json:"model" dynamodbav:"Model"`
	Status          string `json:"status" dynamodbav:"Status"`
	Brand           string `json:"brand" dynamodbav:"Brand"`
	District        string `json:"district" dynamodbav:"District"`
	PersonInCharge  string `json:"person_in_charge" dynamodbav:"PersonInCharge"`
	ReportedBy      string `json:"reported_by" dynamodbav:"ReportedBy"`
	AdditionalNotes string `json:"additional_notes" dynamodbav:"AdditionalNotes"`
	Attachment      string `json:"attachment" dynamodbav:"Attachment"`
	PpmStatus       string `json:"ppm_status" dynamodbav:"PpmStatus"`
	TncDate         string `json:"tnc_date" dynamodbav:"TncDate"`     // ISO 8601 format
	PpmDate         string `json:"ppm_date" dynamodbav:"PpmDate"`     // ISO 8601 format
	CreatedAt       string `json:"created_at" dynamodbav:"CreatedAt"` // ISO 8601 format
	UpdatedAt       string `json:"updated_at" dynamodbav:"UpdatedAt"` // ISO 8601 format
}

// GetPartitionKey returns the partition key for DynamoDB
func (m *Machine) GetPartitionKey() string {
	return "Machine#" + m.SerialNumber
}

// GetSortKey returns the sort key for DynamoDB
func (m *Machine) GetSortKey() string {
	return "#"
}

// SetTimestamps sets the CreatedAt and UpdatedAt timestamps
func (m *Machine) SetTimestamps() {
	now := time.Now().UTC().Format(time.RFC3339)
	if m.CreatedAt == "" {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
