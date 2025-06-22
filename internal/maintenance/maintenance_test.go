package maintenance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaintenance_GetPartitionKey(t *testing.T) {
	tests := []struct {
		name                string
		machineSerialNumber string
		expected            string
	}{
		{
			name:                "valid machine serial number",
			machineSerialNumber: "MACHINE123",
			expected:            "Machine#MACHINE123",
		},
		{
			name:                "empty machine serial number",
			machineSerialNumber: "",
			expected:            "Machine#",
		},
		{
			name:                "special characters",
			machineSerialNumber: "MACHINE-123_ABC",
			expected:            "Machine#MACHINE-123_ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maintenance := &Maintenance{
				MachineSerialNumber: tt.machineSerialNumber,
			}
			result := maintenance.GetPartitionKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaintenance_GetSortKey(t *testing.T) {
	tests := []struct {
		name            string
		workOrderNumber string
		expected        string
	}{
		{
			name:            "valid work order number",
			workOrderNumber: "WO123",
			expected:        "Maintenance#WO123",
		},
		{
			name:            "empty work order number",
			workOrderNumber: "",
			expected:        "Maintenance#",
		},
		{
			name:            "special characters",
			workOrderNumber: "WO-123_ABC",
			expected:        "Maintenance#WO-123_ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maintenance := &Maintenance{
				WorkOrderNumber: tt.workOrderNumber,
			}
			result := maintenance.GetSortKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaintenance_SetTimestamps(t *testing.T) {
	t.Run("new maintenance", func(t *testing.T) {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
		}

		maintenance.SetTimestamps()

		// Parse the timestamps
		createdAt, err := time.Parse(time.RFC3339, maintenance.CreatedAt)
		require.NoError(t, err)
		updatedAt, err := time.Parse(time.RFC3339, maintenance.UpdatedAt)
		require.NoError(t, err)

		// Just check that timestamps are not empty and are equal
		assert.NotEmpty(t, createdAt)
		assert.NotEmpty(t, updatedAt)
		assert.Equal(t, maintenance.CreatedAt, maintenance.UpdatedAt)
	})

	t.Run("existing maintenance", func(t *testing.T) {
		originalCreatedAt := "2023-01-01T00:00:00Z"
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
			CreatedAt:           originalCreatedAt,
			UpdatedAt:           "2023-01-01T01:00:00Z",
		}

		maintenance.SetTimestamps()

		// CreatedAt should remain unchanged
		assert.Equal(t, originalCreatedAt, maintenance.CreatedAt)

		// UpdatedAt should be updated and parseable
		updatedAt, err := time.Parse(time.RFC3339, maintenance.UpdatedAt)
		require.NoError(t, err)
		assert.NotEmpty(t, updatedAt)
	})
}

func TestMaintenance_Complete(t *testing.T) {
	maintenance := &Maintenance{
		MachineSerialNumber: "MACHINE123",
		WorkOrderNumber:     "WO456",
		WorkOrderDate:       "2024-01-15",
		ActionTaken:         "Replaced filter",
		ReportedBy:          "John Doe",
		WorkerOrderType:     "Preventive",
		Attachment:          "work_order.pdf",
	}

	// Test all fields are set correctly
	assert.Equal(t, "MACHINE123", maintenance.MachineSerialNumber)
	assert.Equal(t, "WO456", maintenance.WorkOrderNumber)
	assert.Equal(t, "2024-01-15", maintenance.WorkOrderDate)
	assert.Equal(t, "Replaced filter", maintenance.ActionTaken)
	assert.Equal(t, "John Doe", maintenance.ReportedBy)
	assert.Equal(t, "Preventive", maintenance.WorkerOrderType)
	assert.Equal(t, "work_order.pdf", maintenance.Attachment)

	// Test partition and sort keys
	assert.Equal(t, "Machine#MACHINE123", maintenance.GetPartitionKey())
	assert.Equal(t, "Maintenance#WO456", maintenance.GetSortKey())

	// Test timestamps
	maintenance.SetTimestamps()
	assert.NotEmpty(t, maintenance.CreatedAt)
	assert.NotEmpty(t, maintenance.UpdatedAt)
}
