package machine_test

import (
	"testing"
	"time"

	"ralts-cms/internal/machine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMachine_GetPartitionKey(t *testing.T) {
	tests := []struct {
		name         string
		serialNumber string
		expected     string
	}{
		{
			name:         "valid serial number",
			serialNumber: "MACHINE123",
			expected:     "Machine#MACHINE123",
		},
		{
			name:         "empty serial number",
			serialNumber: "",
			expected:     "Machine#",
		},
		{
			name:         "special characters",
			serialNumber: "MACHINE-123_ABC",
			expected:     "Machine#MACHINE-123_ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine := &machine.Machine{
				SerialNumber: tt.serialNumber,
			}
			result := machine.GetPartitionKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMachine_GetSortKey(t *testing.T) {
	machine := &machine.Machine{}
	result := machine.GetSortKey()
	assert.Equal(t, "#", result)
}

func TestMachine_SetTimestamps(t *testing.T) {
	t.Run("new machine", func(t *testing.T) {
		machine := &machine.Machine{
			SerialNumber: "TEST123",
		}

		machine.SetTimestamps()

		// Parse the timestamps
		createdAt, err := time.Parse(time.RFC3339, machine.CreatedAt)
		require.NoError(t, err)
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt)
		require.NoError(t, err)

		// Just check that timestamps are not empty and are equal
		assert.NotEmpty(t, createdAt)
		assert.NotEmpty(t, updatedAt)
		assert.Equal(t, machine.CreatedAt, machine.UpdatedAt)
	})

	t.Run("existing machine", func(t *testing.T) {
		originalCreatedAt := "2023-01-01T00:00:00Z"
		machine := &machine.Machine{
			SerialNumber: "TEST123",
			CreatedAt:    originalCreatedAt,
			UpdatedAt:    "2023-01-01T01:00:00Z",
		}

		machine.SetTimestamps()

		// CreatedAt should remain unchanged
		assert.Equal(t, originalCreatedAt, machine.CreatedAt)

		// UpdatedAt should be updated and parseable
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt)
		require.NoError(t, err)
		assert.NotEmpty(t, updatedAt)
	})
}

func TestMachine_Complete(t *testing.T) {
	machine := &machine.Machine{
		SerialNumber:    "MACHINE123",
		Customer:        "Test Customer",
		State:           "Active",
		AccountType:     "Premium",
		Model:           "Model X",
		Status:          "Operational",
		Brand:           "TestBrand",
		District:        "District A",
		PersonInCharge:  "John Doe",
		ReportedBy:      "Jane Smith",
		AdditionalNotes: "Test notes",
		Attachment:      "attachment.pdf",
		PpmStatus:       "Scheduled",
		TncDate:         "2024-01-15",
		PpmDate:         "2024-02-15",
	}

	// Test all fields are set correctly
	assert.Equal(t, "MACHINE123", machine.SerialNumber)
	assert.Equal(t, "Test Customer", machine.Customer)
	assert.Equal(t, "Active", machine.State)
	assert.Equal(t, "Premium", machine.AccountType)
	assert.Equal(t, "Model X", machine.Model)
	assert.Equal(t, "Operational", machine.Status)
	assert.Equal(t, "TestBrand", machine.Brand)
	assert.Equal(t, "District A", machine.District)
	assert.Equal(t, "John Doe", machine.PersonInCharge)
	assert.Equal(t, "Jane Smith", machine.ReportedBy)
	assert.Equal(t, "Test notes", machine.AdditionalNotes)
	assert.Equal(t, "attachment.pdf", machine.Attachment)
	assert.Equal(t, "Scheduled", machine.PpmStatus)
	assert.Equal(t, "2024-01-15", machine.TncDate)
	assert.Equal(t, "2024-02-15", machine.PpmDate)

	// Test partition and sort keys
	assert.Equal(t, "Machine#MACHINE123", machine.GetPartitionKey())
	assert.Equal(t, "#", machine.GetSortKey())

	// Test timestamps
	machine.SetTimestamps()
	assert.NotEmpty(t, machine.CreatedAt)
	assert.NotEmpty(t, machine.UpdatedAt)
}
