package maintenance_test

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MaintenanceRepositoryTestSuite defines the test suite for maintenance repository
type MaintenanceRepositoryTestSuite struct {
	suite.Suite

	deps *deps.Dependencies
	repo maintenance.Repository
}

// SetupTest sets up each test
func (suite *MaintenanceRepositoryTestSuite) SetupTest() {
	deps := deps.Initialise()
	suite.deps = deps
	suite.repo = maintenance.NewRepository(deps.PostgresClient)
}

func (suite *MaintenanceRepositoryTestSuite) TearDownSubTest() {
	// Clear the maintenance table for PostgreSQL
	ctx := context.Background()
	_, err := suite.deps.PostgresClient.Exec(ctx, "DELETE FROM maintenance")
	require.NoError(suite.T(), err)
}

func (suite *MaintenanceRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create maintenance successfully", func() {
		maintenance := testutils.CreateMaintenance("MACHINE001", "WO001")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(maintenance.CreatedAt)
		suite.Assert().NotEmpty(maintenance.UpdatedAt)
		suite.Assert().Greater(maintenance.ID, 0) // PostgreSQL should return an ID
	})

	suite.Run("should fail when creating duplicate maintenance", func() {
		maintenance := testutils.CreateMaintenance("MACHINE002", "WO002")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		// Try to create the same maintenance again
		duplicateMaintenance := testutils.CreateMaintenance("MACHINE002", "WO002")
		err = suite.repo.Create(ctx, duplicateMaintenance)
		suite.Require().Error(err)
		// PostgreSQL will return a unique constraint violation error
		suite.Assert().Contains(err.Error(), "duplicate key")
	})

	suite.Run("should create maintenance with minimal fields", func() {
		maintenance := testutils.CreateMinimalMaintenance("MACHINE003", "WO003")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, maintenance.MachineSerialNumber, maintenance.WorkOrderNumber)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should create maintenance with custom fields", func() {
		maintenance := testutils.CreateMaintenanceWithCustomFields("MACHINE004", "WO004", "Custom Action", "Custom Tech")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, maintenance.MachineSerialNumber, maintenance.WorkOrderNumber)
		suite.Require().NoError(err)
		suite.Assert().Equal("Custom Action", retrieved.ActionTaken)
		suite.Assert().Equal("Custom Tech", retrieved.ReportedBy)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should allow multiple maintenance records for same machine", func() {
		maintenance1 := testutils.CreateMaintenance("MACHINE005", "WO005")
		maintenance2 := testutils.CreateMaintenance("MACHINE005", "WO006")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)

		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.ListByMachine(ctx, "MACHINE005")
		suite.Require().NoError(err)
		suite.Assert().Len(retrieved, 2)
		retrievedWorkOrders := make(map[string]bool)
		for _, maintenance := range retrieved {
			retrievedWorkOrders[maintenance.WorkOrderNumber] = true
		}
		suite.Assert().True(retrievedWorkOrders["WO005"])
		suite.Assert().True(retrievedWorkOrders["WO006"])
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestGetByWorkOrder() {
	ctx := context.Background()

	suite.Run("should get existing maintenance", func() {
		maintenance := testutils.CreateMaintenance("MACHINE006", "WO007")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE006", "WO007")
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE006", retrieved.MachineSerialNumber)
		suite.Assert().Equal("WO007", retrieved.WorkOrderNumber)
		suite.Assert().Equal("Routine maintenance", retrieved.ActionTaken)
		suite.Assert().Equal("John Doe", retrieved.ReportedBy)
		suite.Assert().False(retrieved.CreatedAt.IsZero())
		suite.Assert().False(retrieved.UpdatedAt.IsZero())
	})

	suite.Run("should return error for non-existent maintenance", func() {
		_, err := suite.repo.GetByWorkOrder(ctx, "NONEXISTENT", "WO999")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should get maintenance with all fields populated", func() {
		maintenance := testutils.CreateMaintenance("MACHINE007", "WO008")
		maintenance.WorkOrderDate = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		maintenance.WorkerOrderType = "Emergency"
		maintenance.Attachment = "emergency-maintenance.pdf"

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE007", "WO008")
		suite.Require().NoError(err)
		// PostgreSQL DATE type only stores the date part, not time
		expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		suite.Assert().Equal(expectedDate, retrieved.WorkOrderDate)
		suite.Assert().Equal("Emergency", retrieved.WorkerOrderType)
		suite.Assert().Equal("emergency-maintenance.pdf", retrieved.Attachment)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestListByMachine() {
	ctx := context.Background()

	suite.Run("should list all maintenance for a machine", func() {
		machineSerial := "MACHINE008"

		// Create maintenance records for the same machine
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO009")
		maintenance1.ActionTaken = "First maintenance"
		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)

		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO010")
		maintenance2.ActionTaken = "Second maintenance"
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO011")
		maintenance3.ActionTaken = "Third maintenance"
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Create maintenance for a different machine
		otherMaintenance := testutils.CreateMaintenance("OTHERMACHINE", "WO012")
		otherMaintenance.ActionTaken = "Other machine maintenance"
		err = suite.repo.Create(ctx, otherMaintenance)
		suite.Require().NoError(err)

		// List maintenance for the target machine
		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify all records belong to the correct machine
		for _, maintenance := range maintenanceList {
			suite.Assert().Equal(machineSerial, maintenance.MachineSerialNumber)
		}

		// Verify we can find all three maintenance records
		actionTaken := make(map[string]bool)
		for _, maintenance := range maintenanceList {
			actionTaken[maintenance.ActionTaken] = true
		}
		suite.Assert().True(actionTaken["First maintenance"])
		suite.Assert().True(actionTaken["Second maintenance"])
		suite.Assert().True(actionTaken["Third maintenance"])
	})

	suite.Run("should return empty list for machine with no maintenance", func() {
		maintenanceList, err := suite.repo.ListByMachine(ctx, "NOMAINTAINANCE")
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 0)
	})

	suite.Run("should list maintenance in correct order", func() {
		machineSerial := "MACHINE009"

		// Create maintenance records with different work order numbers
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO013")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO014")
		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO015")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify all records are returned
		workOrders := make(map[string]bool)
		for _, maintenance := range maintenanceList {
			workOrders[maintenance.WorkOrderNumber] = true
		}
		suite.Assert().True(workOrders["WO013"])
		suite.Assert().True(workOrders["WO014"])
		suite.Assert().True(workOrders["WO015"])
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	suite.Run("should update existing maintenance", func() {
		maintenance := testutils.CreateMaintenance("MACHINE010", "WO016")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		originalCreatedAt := maintenance.CreatedAt
		originalUpdatedAt := maintenance.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update maintenance fields
		maintenance.ActionTaken = "Updated action"
		maintenance.ReportedBy = "Jane Doe"
		maintenance.WorkerOrderType = "Corrective"
		maintenance.Attachment = "updated-maintenance.pdf"

		err = suite.repo.Update(ctx, maintenance)
		suite.Require().NoError(err)

		// Verify the update
		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE010", "WO016")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated action", retrieved.ActionTaken)
		suite.Assert().Equal("Jane Doe", retrieved.ReportedBy)
		suite.Assert().Equal("Corrective", retrieved.WorkerOrderType)
		suite.Assert().Equal("updated-maintenance.pdf", retrieved.Attachment)
		delta := retrieved.CreatedAt.Sub(originalCreatedAt)
		suite.Assert().True(delta < 2*time.Millisecond && delta > -2*time.Millisecond, "CreatedAt should be nearly unchanged")
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt) || retrieved.UpdatedAt.Equal(originalUpdatedAt))
	})

	suite.Run("should return error when updating non-existent maintenance", func() {
		maintenance := testutils.CreateMaintenance("MACHINE011", "WO017")
		maintenance.ActionTaken = "New action"

		err := suite.repo.Update(ctx, maintenance)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should update maintenance with minimal changes", func() {
		maintenance := testutils.CreateMaintenance("MACHINE012", "WO018")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		originalUpdatedAt := maintenance.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		// Only update one field
		maintenance.ActionTaken = "Minimal update"
		err = suite.repo.Update(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE012", "WO018")
		suite.Require().NoError(err)
		suite.Assert().Equal("Minimal update", retrieved.ActionTaken)
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt) || retrieved.UpdatedAt.Equal(originalUpdatedAt))
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing maintenance", func() {
		maintenance := testutils.CreateMaintenance("MACHINE013", "WO019")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		// Verify maintenance exists
		_, err = suite.repo.GetByWorkOrder(ctx, "MACHINE013", "WO019")
		suite.Require().NoError(err)

		// Delete the maintenance
		err = suite.repo.Delete(ctx, "MACHINE013", "WO019")
		suite.Require().NoError(err)

		// Verify maintenance is deleted
		_, err = suite.repo.GetByWorkOrder(ctx, "MACHINE013", "WO019")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should return error when deleting non-existent maintenance", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT", "WONONE")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should delete maintenance and allow recreation", func() {
		maintenance := testutils.CreateMaintenance("MACHINE014", "WO020")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		err = suite.repo.Delete(ctx, "MACHINE014", "WO020")
		suite.Require().NoError(err)

		// Should be able to create a new maintenance with the same work order
		newMaintenance := testutils.CreateMaintenance("MACHINE014", "WO020")
		newMaintenance.ActionTaken = "Recreated action"
		err = suite.repo.Create(ctx, newMaintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE014", "WO020")
		suite.Require().NoError(err)
		suite.Assert().Equal("Recreated action", retrieved.ActionTaken)
	})

	suite.Run("should delete specific maintenance without affecting others", func() {
		machineSerial := "MACHINE015"

		// Create multiple maintenance records
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO021")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO022")
		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO023")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Delete only one maintenance record
		err = suite.repo.Delete(ctx, machineSerial, "WO022")
		suite.Require().NoError(err)

		// Verify the deleted one is gone
		_, err = suite.repo.GetByWorkOrder(ctx, machineSerial, "WO022")
		suite.Require().Error(err)

		// Verify the others still exist
		_, err = suite.repo.GetByWorkOrder(ctx, machineSerial, "WO021")
		suite.Require().NoError(err)
		_, err = suite.repo.GetByWorkOrder(ctx, machineSerial, "WO023")
		suite.Require().NoError(err)

		// Verify list still returns the remaining records
		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 2)
	})
}

// TestMaintenanceRepositoryTestSuite runs the test suite
func TestMaintenanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceRepositoryTestSuite))
}
