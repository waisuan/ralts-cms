package maintenance_test

import (
	"context"
	"fmt"
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

		retrieved, err := suite.repo.ListByMachine(ctx, "MACHINE005", nil)
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
		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, nil)
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
		maintenanceList, err := suite.repo.ListByMachine(ctx, "NOMAINTAINANCE", nil)
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

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, nil)
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

	suite.Run("should respect limit parameter", func() {
		machineSerial := "MACHINE_LIMIT_TEST"

		// Create 5 maintenance records
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(machineSerial, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test with limit of 3
		options := &maintenance.ListOptions{
			Limit:  3,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)
	})

	suite.Run("should respect offset parameter", func() {
		machineSerial := "MACHINE_OFFSET_TEST"

		// Create 5 maintenance records
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(machineSerial, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Get all records first to establish order
		allRecords, err := suite.repo.ListByMachine(ctx, machineSerial, nil)
		suite.Require().NoError(err)
		suite.Require().Len(allRecords, 5)

		// Test with offset of 2
		options := &maintenance.ListOptions{
			Limit:  10,
			Offset: 2,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3) // Should return 3 records (5 total - 2 offset)

		// Verify the returned records are the correct ones (skipping first 2)
		for i, record := range maintenanceList {
			suite.Assert().Equal(allRecords[i+2].WorkOrderNumber, record.WorkOrderNumber)
		}
	})

	suite.Run("should sort by created_at in descending order by default", func() {
		machineSerial := "MACHINE_SORT_DESC_TEST"

		// Create 3 maintenance records
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO001")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO002")
		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO003")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Test with default sort (descending)
		options := &maintenance.ListOptions{
			Limit:  10,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify descending order (most recent first)
		suite.Assert().Equal("WO003", maintenanceList[0].WorkOrderNumber)
		suite.Assert().Equal("WO002", maintenanceList[1].WorkOrderNumber)
		suite.Assert().Equal("WO001", maintenanceList[2].WorkOrderNumber)
	})

	suite.Run("should sort by created_at in ascending order", func() {
		machineSerial := "MACHINE_SORT_ASC_TEST"

		// Create 3 maintenance records
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO001")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO002")
		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO003")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Test with ascending sort
		options := &maintenance.ListOptions{
			Limit:  10,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtAsc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify ascending order (oldest first)
		suite.Assert().Equal("WO001", maintenanceList[0].WorkOrderNumber)
		suite.Assert().Equal("WO002", maintenanceList[1].WorkOrderNumber)
		suite.Assert().Equal("WO003", maintenanceList[2].WorkOrderNumber)
	})

	suite.Run("should handle pagination with limit and offset", func() {
		machineSerial := "MACHINE_PAGINATION_TEST"

		// Create 10 maintenance records
		for i := 1; i <= 10; i++ {
			maintenance := testutils.CreateMaintenance(machineSerial, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test first page (limit 3, offset 0)
		options1 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page1, err := suite.repo.ListByMachine(ctx, machineSerial, options1)
		suite.Require().NoError(err)
		suite.Assert().Len(page1, 3)

		// Test second page (limit 3, offset 3)
		options2 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 3,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page2, err := suite.repo.ListByMachine(ctx, machineSerial, options2)
		suite.Require().NoError(err)
		suite.Assert().Len(page2, 3)

		// Test third page (limit 3, offset 6)
		options3 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 6,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page3, err := suite.repo.ListByMachine(ctx, machineSerial, options3)
		suite.Require().NoError(err)
		suite.Assert().Len(page3, 3)

		// Test last page (limit 3, offset 9)
		options4 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 9,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page4, err := suite.repo.ListByMachine(ctx, machineSerial, options4)
		suite.Require().NoError(err)
		suite.Assert().Len(page4, 1) // Only 1 record left

		// Verify no overlap between pages
		allWorkOrders := make(map[string]bool)
		for _, record := range append(append(append(page1, page2...), page3...), page4...) {
			allWorkOrders[record.WorkOrderNumber] = true
		}
		suite.Assert().Len(allWorkOrders, 10) // All 10 records should be unique
	})

	suite.Run("should handle out of bounds offset and limit", func() {
		machineSerial := "MACHINE_EDGE_CASES"

		// Create 2 maintenance records
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO001")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO002")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		// Test offset beyond available records
		options := &maintenance.ListOptions{
			Limit:  10,
			Offset: 5, // Beyond the 2 records we have
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 0) // Should return empty list

		// Test with zero limit
		optionsZero := &maintenance.ListOptions{
			Limit:  0,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceListZero, err := suite.repo.ListByMachine(ctx, machineSerial, optionsZero)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceListZero, 0) // Should return empty list
	})

	suite.Run("should use default options when nil is passed", func() {
		machineSerial := "MACHINE_DEFAULT_OPTIONS"

		// Create 3 maintenance records
		for i := 1; i <= 3; i++ {
			maintenance := testutils.CreateMaintenance(machineSerial, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test with nil options (should use defaults)
		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3) // Should return all records with default sorting
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
		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 2)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCount() {
	ctx := context.Background()

	suite.Run("should return count of all maintenance", func() {
		maintenance1 := testutils.CreateMaintenance("MACHINE016", "WO024")
		maintenance2 := testutils.CreateMaintenance("MACHINE016", "WO025")
		maintenance3 := testutils.CreateMaintenance("MACHINE016", "WO026")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		count, err := suite.repo.Count(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should return 0 when no maintenance records exist", func() {
		count, err := suite.repo.Count(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCountByMachine() {
	ctx := context.Background()

	suite.Run("should return count of maintenance for specific machine", func() {
		machineSerial := "MACHINE_COUNT_TEST"

		// Create maintenance records for the target machine
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO001")
		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO002")
		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO003")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Create maintenance for a different machine
		otherMaintenance := testutils.CreateMaintenance("OTHER_MACHINE", "WO004")
		err = suite.repo.Create(ctx, otherMaintenance)
		suite.Require().NoError(err)

		count, err := suite.repo.CountByMachine(ctx, machineSerial)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should return 0 for machine with no maintenance records", func() {
		count, err := suite.repo.CountByMachine(ctx, "NO_MAINTENANCE_MACHINE")
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})

	suite.Run("should return 0 for non-existent machine", func() {
		count, err := suite.repo.CountByMachine(ctx, "NON_EXISTENT_MACHINE")
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCountByWorkOrderType() {
	ctx := context.Background()

	suite.Run("should return correct counts for all work order types", func() {
		machineSerial := "MACHINE_WORK_ORDER_TYPE_TEST"

		// Create maintenance records with different work order types
		maintenance1 := testutils.CreateMaintenance(machineSerial, "WO001")
		maintenance1.WorkerOrderType = "Preventive"

		maintenance2 := testutils.CreateMaintenance(machineSerial, "WO002")
		maintenance2.WorkerOrderType = "Corrective"

		maintenance3 := testutils.CreateMaintenance(machineSerial, "WO003")
		maintenance3.WorkerOrderType = "Emergency"

		maintenance4 := testutils.CreateMaintenance(machineSerial, "WO004")
		maintenance4.WorkerOrderType = "Inspection"

		maintenance5 := testutils.CreateMaintenance(machineSerial, "WO005")
		maintenance5.WorkerOrderType = "Preventive"

		maintenance6 := testutils.CreateMaintenance(machineSerial, "WO006")
		maintenance6.WorkerOrderType = "Corrective"

		// Create maintenance for a different machine
		otherMaintenance := testutils.CreateMaintenance("OTHER_MACHINE", "WO007")
		otherMaintenance.WorkerOrderType = "Preventive"

		// Create all maintenance records
		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance4)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance5)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance6)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, otherMaintenance)
		suite.Require().NoError(err)

		// Get counts for the target machine
		preventativeCount, correctiveCount, emergencyCount, inspectionCount, err := suite.repo.CountByWorkOrderType(ctx, machineSerial)
		suite.Require().NoError(err)

		// Verify counts
		suite.Assert().Equal(2, preventativeCount) // WO001, WO005
		suite.Assert().Equal(2, correctiveCount)   // WO002, WO006
		suite.Assert().Equal(1, emergencyCount)    // WO003
		suite.Assert().Equal(1, inspectionCount)   // WO004
	})

	suite.Run("should return 0 for all types when machine has no maintenance records", func() {
		preventativeCount, correctiveCount, emergencyCount, inspectionCount, err := suite.repo.CountByWorkOrderType(ctx, "NO_MAINTENANCE_MACHINE")
		suite.Require().NoError(err)

		suite.Assert().Equal(0, preventativeCount)
		suite.Assert().Equal(0, correctiveCount)
		suite.Assert().Equal(0, emergencyCount)
		suite.Assert().Equal(0, inspectionCount)
	})

	suite.Run("should return 0 for non-existent machine", func() {
		preventativeCount, correctiveCount, emergencyCount, inspectionCount, err := suite.repo.CountByWorkOrderType(ctx, "NON_EXISTENT_MACHINE")
		suite.Require().NoError(err)

		suite.Assert().Equal(0, preventativeCount)
		suite.Assert().Equal(0, correctiveCount)
		suite.Assert().Equal(0, emergencyCount)
		suite.Assert().Equal(0, inspectionCount)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestSearch() {
	ctx := context.Background()

	suite.Run("should search maintenance by work order number", func() {
		// Create test maintenance records with different work order numbers
		maintenance1 := testutils.CreateMaintenance("MACHINE001", "WO001")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE002", "WO002")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		// Search for WO001
		results, err := suite.repo.Search(ctx, "WO001", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("WO001", results[0].WorkOrderNumber)
		suite.Assert().Equal("MACHINE001", results[0].MachineSerialNumber)
	})

	suite.Run("should search maintenance by reported by name", func() {
		// Create test maintenance records with different reported by names
		maintenance1 := testutils.CreateMaintenance("MACHINE003", "WO003")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE004", "WO004")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		// Search for "John"
		results, err := suite.repo.Search(ctx, "John", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("John Tech", results[0].ReportedBy)
		suite.Assert().Equal("WO003", results[0].WorkOrderNumber)
	})

	suite.Run("should search maintenance by worker order type", func() {
		// Create test maintenance records with different worker order types
		maintenance1 := testutils.CreateMaintenance("MACHINE005", "WO005")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE006", "WO006")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		// Search for "Preventive"
		results, err := suite.repo.Search(ctx, "Preventive", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("Preventive", results[0].WorkerOrderType)
		suite.Assert().Equal("WO005", results[0].WorkOrderNumber)
	})

	suite.Run("should search with pagination", func() {
		// Create multiple maintenance records
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(fmt.Sprintf("MACHINE%03d", i), fmt.Sprintf("WO%03d", i))
			maintenance.ReportedBy = "John Tech"
			maintenance.WorkerOrderType = "Preventive"
			suite.Require().NoError(suite.repo.Create(ctx, maintenance))
		}

		// Search with limit 2
		results, err := suite.repo.Search(ctx, "John", &maintenance.ListOptions{
			Limit:  2,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
	})

	suite.Run("should return empty results for non-matching query", func() {
		// Create a maintenance record
		maintenanceRecord := testutils.CreateMaintenance("MACHINE007", "WO007")
		maintenanceRecord.ReportedBy = "John Tech"
		maintenanceRecord.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Search for non-matching term
		results, err := suite.repo.Search(ctx, "NonExistentTerm", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 0)
	})

	suite.Run("should search with different sort orders", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE008", "WO008")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE009", "WO009")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		// Search with different sort orders
		results, err := suite.repo.Search(ctx, "Tech", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateAsc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCountSearch() {
	ctx := context.Background()

	suite.Run("should count search results correctly", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE010", "WO010")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE011", "WO011")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE012", "WO012")
		maintenance3.ReportedBy = "Bob Tech"
		maintenance3.WorkerOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Count search results for "Tech"
		count, err := suite.repo.CountSearch(ctx, "Tech", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should return zero for non-matching query", func() {
		// Create a maintenance record
		maintenanceRecord := testutils.CreateMaintenance("MACHINE013", "WO013")
		maintenanceRecord.ReportedBy = "John Tech"
		maintenanceRecord.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Count search results for non-matching term
		count, err := suite.repo.CountSearch(ctx, "NonExistentTerm", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestSearchByFields() {
	ctx := context.Background()

	suite.Run("should search by work order number", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE020", "WO020")
		maintenance1.ReportedBy = "John Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE020", "WO021")
		maintenance2.ReportedBy = "Jane Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE020", "WO999")
		maintenance3.ReportedBy = "Bob Tech"
		maintenance3.WorkerOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Search by work order number
		filters := &maintenance.SearchFilters{
			WorkOrderQuery: "WO02",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE020", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
		suite.Assert().Contains([]string{results[0].WorkOrderNumber, results[1].WorkOrderNumber}, "WO020")
		suite.Assert().Contains([]string{results[0].WorkOrderNumber, results[1].WorkOrderNumber}, "WO021")
	})

	suite.Run("should search by reported by name", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE021", "WO030")
		maintenance1.ReportedBy = "Alice Johnson"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE021", "WO031")
		maintenance2.ReportedBy = "Bob Johnson"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE021", "WO032")
		maintenance3.ReportedBy = "Charlie Smith"
		maintenance3.WorkerOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Search by reported by name
		filters := &maintenance.SearchFilters{
			ReportedByQuery: "Johnson",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE021", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
		suite.Assert().Contains([]string{results[0].ReportedBy, results[1].ReportedBy}, "Alice Johnson")
		suite.Assert().Contains([]string{results[0].ReportedBy, results[1].ReportedBy}, "Bob Johnson")
	})

	suite.Run("should search by worker order type", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE022", "WO040")
		maintenance1.ReportedBy = "Tech A"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE022", "WO041")
		maintenance2.ReportedBy = "Tech B"
		maintenance2.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE022", "WO042")
		maintenance3.ReportedBy = "Tech C"
		maintenance3.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Search by worker order type
		filters := &maintenance.SearchFilters{
			WorkerOrderType: "Preventive",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE022", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
		suite.Assert().Equal("Preventive", results[0].WorkerOrderType)
		suite.Assert().Equal("Preventive", results[1].WorkerOrderType)
	})

	suite.Run("should search with multiple filters", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE023", "WO050")
		maintenance1.ReportedBy = "John Smith"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE023", "WO051")
		maintenance2.ReportedBy = "John Doe"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE023", "WO052")
		maintenance3.ReportedBy = "Jane Smith"
		maintenance3.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Search with multiple filters
		filters := &maintenance.SearchFilters{
			ReportedByQuery: "John",
			WorkerOrderType: "Preventive",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE023", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("John Smith", results[0].ReportedBy)
		suite.Assert().Equal("Preventive", results[0].WorkerOrderType)
	})

	suite.Run("should return empty results for no matches", func() {
		// Create test maintenance record
		maintenanceRecord := testutils.CreateMaintenance("MACHINE024", "WO060")
		maintenanceRecord.ReportedBy = "Tech Person"
		maintenanceRecord.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Search with non-matching filters
		filters := &maintenance.SearchFilters{
			WorkOrderQuery: "NONEXISTENT",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE024", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 0)
	})

	suite.Run("should handle pagination", func() {
		// Create multiple test maintenance records
		for i := 0; i < 5; i++ {
			maintenanceRecord := testutils.CreateMaintenance("MACHINE025", fmt.Sprintf("WO%03d", 70+i))
			maintenanceRecord.ReportedBy = "Pagination Tech"
			maintenanceRecord.WorkerOrderType = "Preventive"
			suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))
		}

		// Search with pagination
		filters := &maintenance.SearchFilters{
			ReportedByQuery: "Pagination",
		}
		results, err := suite.repo.SearchByFields(ctx, "MACHINE025", filters, &maintenance.ListOptions{
			Limit:  2,
			Offset: 1,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCountSearchByFields() {
	ctx := context.Background()

	suite.Run("should count search results by work order number", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE030", "WO080")
		maintenance1.ReportedBy = "Count Tech"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE030", "WO081")
		maintenance2.ReportedBy = "Count Tech"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE030", "WO999")
		maintenance3.ReportedBy = "Count Tech"
		maintenance3.WorkerOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Count search results by work order number
		filters := &maintenance.SearchFilters{
			WorkOrderQuery: "WO08",
		}
		count, err := suite.repo.CountSearchByFields(ctx, "MACHINE030", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count)
	})

	suite.Run("should count search results by reported by name", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE031", "WO090")
		maintenance1.ReportedBy = "Alice Counter"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE031", "WO091")
		maintenance2.ReportedBy = "Bob Counter"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE031", "WO092")
		maintenance3.ReportedBy = "Charlie Smith"
		maintenance3.WorkerOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Count search results by reported by name
		filters := &maintenance.SearchFilters{
			ReportedByQuery: "Counter",
		}
		count, err := suite.repo.CountSearchByFields(ctx, "MACHINE031", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count)
	})

	suite.Run("should count search results by worker order type", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE032", "WO100")
		maintenance1.ReportedBy = "Type Counter"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE032", "WO101")
		maintenance2.ReportedBy = "Type Counter"
		maintenance2.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE032", "WO102")
		maintenance3.ReportedBy = "Type Counter"
		maintenance3.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Count search results by worker order type
		filters := &maintenance.SearchFilters{
			WorkerOrderType: "Preventive",
		}
		count, err := suite.repo.CountSearchByFields(ctx, "MACHINE032", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count)
	})

	suite.Run("should count with multiple filters", func() {
		// Create test maintenance records
		maintenance1 := testutils.CreateMaintenance("MACHINE033", "WO110")
		maintenance1.ReportedBy = "Multi Filter"
		maintenance1.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance("MACHINE033", "WO111")
		maintenance2.ReportedBy = "Multi Filter"
		maintenance2.WorkerOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance("MACHINE033", "WO112")
		maintenance3.ReportedBy = "Other Person"
		maintenance3.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Count with multiple filters
		filters := &maintenance.SearchFilters{
			ReportedByQuery: "Multi",
			WorkerOrderType: "Preventive",
		}
		count, err := suite.repo.CountSearchByFields(ctx, "MACHINE033", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, count)
	})

	suite.Run("should return zero for no matches", func() {
		// Create test maintenance record
		maintenanceRecord := testutils.CreateMaintenance("MACHINE034", "WO120")
		maintenanceRecord.ReportedBy = "Zero Counter"
		maintenanceRecord.WorkerOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Count with non-matching filters
		filters := &maintenance.SearchFilters{
			WorkOrderQuery: "NONEXISTENT",
		}
		count, err := suite.repo.CountSearchByFields(ctx, "MACHINE034", filters, &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})
}

// TestMaintenanceRepositoryTestSuite runs the test suite
func TestMaintenanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceRepositoryTestSuite))
}
