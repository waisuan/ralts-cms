package maintenance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// MaintenanceTestSuite defines the test suite for maintenance model
type MaintenanceTestSuite struct {
	suite.Suite
}

func (suite *MaintenanceTestSuite) TestGetPartitionKey() {
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
		suite.Run(tt.name, func() {
			maintenance := &Maintenance{
				MachineSerialNumber: tt.machineSerialNumber,
			}
			result := maintenance.GetPartitionKey()
			suite.Assert().Equal(tt.expected, result)
		})
	}
}

func (suite *MaintenanceTestSuite) TestGetSortKey() {
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
		suite.Run(tt.name, func() {
			maintenance := &Maintenance{
				WorkOrderNumber: tt.workOrderNumber,
			}
			result := maintenance.GetSortKey()
			suite.Assert().Equal(tt.expected, result)
		})
	}
}

func (suite *MaintenanceTestSuite) TestSetTimestamps() {
	suite.Run("new maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
		}

		maintenance.SetTimestamps()

		// Parse the timestamps
		createdAt, err := time.Parse(time.RFC3339, maintenance.CreatedAt)
		suite.Require().NoError(err)
		updatedAt, err := time.Parse(time.RFC3339, maintenance.UpdatedAt)
		suite.Require().NoError(err)

		// Just check that timestamps are not empty and are equal
		suite.Assert().NotEmpty(createdAt)
		suite.Assert().NotEmpty(updatedAt)
		suite.Assert().Equal(maintenance.CreatedAt, maintenance.UpdatedAt)
	})

	suite.Run("existing maintenance", func() {
		originalCreatedAt := "2023-01-01T00:00:00Z"
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
			CreatedAt:           originalCreatedAt,
			UpdatedAt:           "2023-01-01T01:00:00Z",
		}

		maintenance.SetTimestamps()

		// CreatedAt should remain unchanged
		suite.Assert().Equal(originalCreatedAt, maintenance.CreatedAt)

		// UpdatedAt should be updated and parseable
		updatedAt, err := time.Parse(time.RFC3339, maintenance.UpdatedAt)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(updatedAt)
	})
}

// TestMaintenanceTestSuite runs the test suite
func TestMaintenanceTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceTestSuite))
}
