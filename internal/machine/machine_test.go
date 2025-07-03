package machine_test

import (
	"testing"
	"time"

	"ralts-cms/internal/machine"

	"github.com/stretchr/testify/suite"
)

// MachineTestSuite defines the test suite for machine model
type MachineTestSuite struct {
	suite.Suite
}

func (suite *MachineTestSuite) TestGetPartitionKey() {
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
		suite.Run(tt.name, func() {
			machine := &machine.Machine{
				SerialNumber: tt.serialNumber,
			}
			result := machine.GetPartitionKey()
			suite.Assert().Equal(tt.expected, result)
		})
	}
}

func (suite *MachineTestSuite) TestGetSortKey() {
	suite.Run("should return correct sort key", func() {
		machine := &machine.Machine{}
		result := machine.GetSortKey()
		suite.Assert().Equal("#", result)
	})
}

func (suite *MachineTestSuite) TestSetTimestamps() {
	suite.Run("new machine", func() {
		machine := &machine.Machine{
			SerialNumber: "TEST123",
		}

		machine.SetTimestamps()

		// Parse the timestamps
		createdAt, err := time.Parse(time.RFC3339, machine.CreatedAt)
		suite.Require().NoError(err)
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt)
		suite.Require().NoError(err)

		// Just check that timestamps are not empty and are equal
		suite.Assert().NotEmpty(createdAt)
		suite.Assert().NotEmpty(updatedAt)
		suite.Assert().Equal(machine.CreatedAt, machine.UpdatedAt)
	})

	suite.Run("existing machine", func() {
		originalCreatedAt := "2023-01-01T00:00:00Z"
		machine := &machine.Machine{
			SerialNumber: "TEST123",
			CreatedAt:    originalCreatedAt,
			UpdatedAt:    "2023-01-01T01:00:00Z",
		}

		machine.SetTimestamps()

		// CreatedAt should remain unchanged
		suite.Assert().Equal(originalCreatedAt, machine.CreatedAt)

		// UpdatedAt should be updated and parseable
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(updatedAt)
	})
}

// TestMachineTestSuite runs the test suite
func TestMachineTestSuite(t *testing.T) {
	suite.Run(t, new(MachineTestSuite))
}
