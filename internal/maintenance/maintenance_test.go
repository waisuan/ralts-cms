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

func (suite *MaintenanceTestSuite) TestSetTimestamps() {
	suite.Run("new maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
		}

		maintenance.SetTimestamps()

		suite.Assert().False(maintenance.CreatedAt.IsZero())
		suite.Assert().False(maintenance.UpdatedAt.IsZero())
		// CreatedAt and UpdatedAt should be close in time
		delta := maintenance.UpdatedAt.Sub(maintenance.CreatedAt)
		suite.Assert().True(delta >= 0 && delta < time.Second)
	})

	suite.Run("existing maintenance", func() {
		originalCreatedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO123",
			CreatedAt:           originalCreatedAt,
			UpdatedAt:           time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC),
		}

		maintenance.SetTimestamps()

		// CreatedAt should remain unchanged
		suite.Assert().Equal(originalCreatedAt, maintenance.CreatedAt)

		// UpdatedAt should be updated and valid
		suite.Assert().False(maintenance.UpdatedAt.IsZero())
		// UpdatedAt should be after CreatedAt
		suite.Assert().True(maintenance.UpdatedAt.After(maintenance.CreatedAt) || maintenance.UpdatedAt.Equal(maintenance.CreatedAt))
	})
}

// TestMaintenanceTestSuite runs the test suite
func TestMaintenanceTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceTestSuite))
}
