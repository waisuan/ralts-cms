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

func (suite *MachineTestSuite) TestSetTimestamps() {
	suite.Run("new machine", func() {
		machine := &machine.Machine{
			SerialNumber: "TEST123",
		}

		machine.SetTimestamps()

		// Parse the timestamps
		createdAt, err := time.Parse(time.RFC3339, machine.CreatedAt.Format(time.RFC3339))
		suite.Require().NoError(err)
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt.Format(time.RFC3339))
		suite.Require().NoError(err)

		// Just check that timestamps are not empty and are equal
		suite.Assert().NotEmpty(createdAt)
		suite.Assert().NotEmpty(updatedAt)
		suite.Assert().Equal(machine.CreatedAt, machine.UpdatedAt)
	})

	suite.Run("existing machine", func() {
		originalCreatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		machine := &machine.Machine{
			SerialNumber: "TEST123",
			CreatedAt:    originalCreatedAt,
			UpdatedAt:    originalCreatedAt,
		}

		machine.SetTimestamps()

		// CreatedAt should remain unchanged
		suite.Assert().Equal(originalCreatedAt, machine.CreatedAt)

		// UpdatedAt should be updated and parseable
		updatedAt, err := time.Parse(time.RFC3339, machine.UpdatedAt.Format(time.RFC3339))
		suite.Require().NoError(err)
		suite.Assert().True(updatedAt.After(originalCreatedAt))
	})
}

// TestMachineTestSuite runs the test suite
func TestMachineTestSuite(t *testing.T) {
	suite.Run(t, new(MachineTestSuite))
}
