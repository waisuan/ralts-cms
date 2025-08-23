package maintenance_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	TestMachineOne = "MAINTENANCE_MACHINE001"
	TestMachineTwo = "MAINTENANCE_MACHINE002"
)

// MaintenanceRepositoryTestSuite defines the test suite for maintenance repository
type MaintenanceRepositoryTestSuite struct {
	suite.Suite

	db          *testutils.TestDatabase
	repo        maintenance.Repository
	machineRepo machines.Repository
}

// SetupTest sets up each test
func (suite *MaintenanceRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = maintenance.NewRepository(db.PostgresClient)
	suite.machineRepo = machines.NewRepository(db.PostgresClient)
}

func (suite *MaintenanceRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *MaintenanceRepositoryTestSuite) SetupSubTest() {
	ctx := context.Background()

	machineSerials := []string{
		TestMachineOne,
		TestMachineTwo,
	}

	for _, serial := range machineSerials {
		machine := testutils.CreateMachine(serial)
		err := suite.machineRepo.Create(ctx, machine)
		require.NoError(suite.T(), err)
	}
}

func (suite *MaintenanceRepositoryTestSuite) TearDownSubTest() {
	// Clear tables for PostgreSQL (maintenance first due to foreign key constraint)
	ctx := context.Background()
	err := suite.db.CleanupAllTables(ctx)
	require.NoError(suite.T(), err)
}

func (suite *MaintenanceRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create maintenance successfully", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO001")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(maintenance.CreatedAt)
		suite.Assert().NotEmpty(maintenance.UpdatedAt)
		suite.Assert().Greater(maintenance.ID, int64(0)) // PostgreSQL should return an ID
	})

	suite.Run("should fail when machine does not exist", func() {
		maintenance := testutils.CreateMaintenance("NONEXISTENT_MACHINE", "WO001")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "foreign key constraint")
	})

	suite.Run("should fail when creating duplicate maintenance", func() {
		// Create maintenance first
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO002")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		// Try to create the same maintenance again
		duplicateMaintenance := testutils.CreateMaintenance(TestMachineOne, "WO002")
		err = suite.repo.Create(ctx, duplicateMaintenance)
		suite.Require().Error(err)
		// PostgreSQL will return a unique constraint violation error
		suite.Assert().Contains(err.Error(), "duplicate key")
	})

	suite.Run("should create maintenance with minimal fields", func() {
		maintenance := testutils.CreateMinimalMaintenance(TestMachineOne, "WO003")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, maintenance.MachineSerialNumber, maintenance.WorkOrderNumber)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should create maintenance with custom fields", func() {
		maintenance := testutils.CreateMaintenanceWithCustomFields(TestMachineOne, "WO004", "Custom Action", "Custom Tech")
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
		// Create multiple maintenance records for the same machine
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO005")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO006")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)

		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
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
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO007")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO007")
		suite.Require().NoError(err)
		suite.Assert().Equal(TestMachineOne, retrieved.MachineSerialNumber)
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
		// Create maintenance with custom fields
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO008")
		maintenance.WorkOrderDate = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
		maintenance.WorkOrderType = "Emergency"
		maintenance.Attachment = testutils.StringPtr("emergency-maintenance.pdf")

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO008")
		suite.Require().NoError(err)
		// PostgreSQL DATE type only stores the date part, not time
		expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		suite.Assert().Equal(expectedDate, retrieved.WorkOrderDate)
		suite.Assert().Equal("Emergency", retrieved.WorkOrderType)
		suite.Assert().NotNil(retrieved.Attachment)
		suite.Assert().Equal("emergency-maintenance.pdf", *retrieved.Attachment)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestListByMachine() {
	ctx := context.Background()

	suite.Run("should list all maintenance for a machine", func() {
		// Create maintenance records for the same machine
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO009")
		maintenance1.ActionTaken = "First maintenance"
		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO010")
		maintenance2.ActionTaken = "Second maintenance"
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO011")
		maintenance3.ActionTaken = "Third maintenance"
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// List maintenance for the target machine
		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify all records belong to the correct machine
		for _, maintenance := range maintenanceList {
			suite.Assert().Equal(TestMachineOne, maintenance.MachineSerialNumber)
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
		// Create maintenance records with different work order numbers
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO013")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO014")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO015")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
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

		// Create 5 maintenance records
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(TestMachineOne, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test with limit of 3
		options := &maintenance.ListOptions{
			Limit:  3,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)
	})

	suite.Run("should respect offset parameter", func() {
		// Create 5 maintenance records
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(TestMachineOne, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Get all records first to establish order
		allRecords, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
		suite.Require().NoError(err)
		suite.Require().Len(allRecords, 5)

		// Test with offset of 2
		options := &maintenance.ListOptions{
			Limit:  10,
			Offset: 2,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3) // Should return 3 records (5 total - 2 offset)

		// Verify the returned records are the correct ones (skipping first 2)
		for i, record := range maintenanceList {
			suite.Assert().Equal(allRecords[i+2].WorkOrderNumber, record.WorkOrderNumber)
		}
	})

	suite.Run("should sort by created_at in descending order by default", func() {
		// Create 3 maintenance records
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO001")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO002")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO003")

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

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify descending order (most recent first)
		suite.Assert().Equal("WO003", maintenanceList[0].WorkOrderNumber)
		suite.Assert().Equal("WO002", maintenanceList[1].WorkOrderNumber)
		suite.Assert().Equal("WO001", maintenanceList[2].WorkOrderNumber)
	})

	suite.Run("should sort by created_at in ascending order", func() {
		// Create 3 maintenance records
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO001")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO002")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO003")

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

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3)

		// Verify ascending order (oldest first)
		suite.Assert().Equal("WO001", maintenanceList[0].WorkOrderNumber)
		suite.Assert().Equal("WO002", maintenanceList[1].WorkOrderNumber)
		suite.Assert().Equal("WO003", maintenanceList[2].WorkOrderNumber)
	})

	suite.Run("should handle pagination with limit and offset", func() {
		// Create 10 maintenance records
		for i := 1; i <= 10; i++ {
			maintenance := testutils.CreateMaintenance(TestMachineOne, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test first page (limit 3, offset 0)
		options1 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page1, err := suite.repo.ListByMachine(ctx, TestMachineOne, options1)
		suite.Require().NoError(err)
		suite.Assert().Len(page1, 3)

		// Test second page (limit 3, offset 3)
		options2 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 3,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page2, err := suite.repo.ListByMachine(ctx, TestMachineOne, options2)
		suite.Require().NoError(err)
		suite.Assert().Len(page2, 3)

		// Test third page (limit 3, offset 6)
		options3 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 6,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page3, err := suite.repo.ListByMachine(ctx, TestMachineOne, options3)
		suite.Require().NoError(err)
		suite.Assert().Len(page3, 3)

		// Test last page (limit 3, offset 9)
		options4 := &maintenance.ListOptions{
			Limit:  3,
			Offset: 9,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		page4, err := suite.repo.ListByMachine(ctx, TestMachineOne, options4)
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
		// Create 2 maintenance records
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO001")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO002")

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

		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, options)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 0) // Should return empty list

		// Test with zero limit
		optionsZero := &maintenance.ListOptions{
			Limit:  0,
			Offset: 0,
			Sort:   maintenance.SortOrderCreatedAtDesc,
		}

		maintenanceListZero, err := suite.repo.ListByMachine(ctx, TestMachineOne, optionsZero)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceListZero, 0) // Should return empty list
	})

	suite.Run("should use default options when nil is passed", func() {
		// Create 3 maintenance records
		for i := 1; i <= 3; i++ {
			maintenance := testutils.CreateMaintenance(TestMachineOne, fmt.Sprintf("WO%d", i))
			err := suite.repo.Create(ctx, maintenance)
			suite.Require().NoError(err)
		}

		// Test with nil options (should use defaults)
		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 3) // Should return all records with default sorting
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	suite.Run("should update existing maintenance", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO016")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		originalCreatedAt := maintenance.CreatedAt
		originalUpdatedAt := maintenance.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update maintenance fields
		maintenance.ActionTaken = "Updated action"
		maintenance.ReportedBy = "Jane Doe"
		maintenance.WorkOrderType = "Corrective"
		maintenance.Attachment = testutils.StringPtr("updated-maintenance.pdf")

		err = suite.repo.Update(ctx, maintenance)
		suite.Require().NoError(err)

		// Verify the update
		retrieved, err := suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO016")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated action", retrieved.ActionTaken)
		suite.Assert().Equal("Jane Doe", retrieved.ReportedBy)
		suite.Assert().Equal("Corrective", retrieved.WorkOrderType)
		suite.Assert().NotNil(retrieved.Attachment)
		suite.Assert().Equal("updated-maintenance.pdf", *retrieved.Attachment)
		delta := retrieved.CreatedAt.Sub(originalCreatedAt)
		suite.Assert().True(delta < 2*time.Millisecond && delta > -2*time.Millisecond, "CreatedAt should be nearly unchanged")
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt) || retrieved.UpdatedAt.Equal(originalUpdatedAt))
	})

	suite.Run("should return error when updating non-existent maintenance", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO017")
		maintenance.ActionTaken = "New action"

		err := suite.repo.Update(ctx, maintenance)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should update maintenance with minimal changes", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO018")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		originalUpdatedAt := maintenance.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		// Only update one field
		maintenance.ActionTaken = "Minimal update"
		err = suite.repo.Update(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO018")
		suite.Require().NoError(err)
		suite.Assert().Equal("Minimal update", retrieved.ActionTaken)
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt) || retrieved.UpdatedAt.Equal(originalUpdatedAt))
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing maintenance", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO019")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		// Verify maintenance exists
		_, err = suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO019")
		suite.Require().NoError(err)

		// Delete the maintenance
		err = suite.repo.Delete(ctx, TestMachineOne, "WO019")
		suite.Require().NoError(err)

		// Verify maintenance is deleted
		_, err = suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO019")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should return error when deleting non-existent maintenance", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT", "WONONE")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "maintenance not found")
	})

	suite.Run("should delete maintenance and allow recreation", func() {
		maintenance := testutils.CreateMaintenance(TestMachineOne, "WO020")
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		err = suite.repo.Delete(ctx, TestMachineOne, "WO020")
		suite.Require().NoError(err)

		// Should be able to create a new maintenance with the same work order
		newMaintenance := testutils.CreateMaintenance(TestMachineOne, "WO020")
		newMaintenance.ActionTaken = "Recreated action"
		err = suite.repo.Create(ctx, newMaintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO020")
		suite.Require().NoError(err)
		suite.Assert().Equal("Recreated action", retrieved.ActionTaken)
	})

	suite.Run("should delete specific maintenance without affecting others", func() {
		// Create multiple maintenance records
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO021")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO022")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO023")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		// Delete only one maintenance record
		err = suite.repo.Delete(ctx, TestMachineOne, "WO022")
		suite.Require().NoError(err)

		// Verify the deleted one is gone
		_, err = suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO022")
		suite.Require().Error(err)

		// Verify the others still exist
		_, err = suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO021")
		suite.Require().NoError(err)
		_, err = suite.repo.GetByWorkOrder(ctx, TestMachineOne, "WO023")
		suite.Require().NoError(err)

		// Verify list still returns the remaining records
		maintenanceList, err := suite.repo.ListByMachine(ctx, TestMachineOne, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 2)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCount() {
	ctx := context.Background()

	suite.Run("should return count of all maintenance", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO024")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO025")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO026")

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
		// Create maintenance records for the target machine
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO001")
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO002")
		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO003")

		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		count, err := suite.repo.CountByMachine(ctx, TestMachineOne)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should return 0 for machine with no maintenance records", func() {
		count, err := suite.repo.CountByMachine(ctx, TestMachineOne)
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
		// Create maintenance records with different work order types
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO001")
		maintenance1.WorkOrderType = "Preventive"

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO002")
		maintenance2.WorkOrderType = "Corrective"

		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO003")
		maintenance3.WorkOrderType = "Emergency"

		maintenance4 := testutils.CreateMaintenance(TestMachineOne, "WO004")
		maintenance4.WorkOrderType = "Inspection"

		maintenance5 := testutils.CreateMaintenance(TestMachineOne, "WO005")
		maintenance5.WorkOrderType = "Preventive"

		maintenance6 := testutils.CreateMaintenance(TestMachineOne, "WO006")
		maintenance6.WorkOrderType = "Corrective"

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

		// Get counts for the target machine
		preventativeCount, correctiveCount, emergencyCount, inspectionCount, err := suite.repo.CountByWorkOrderType(ctx, TestMachineOne)
		suite.Require().NoError(err)

		// Verify counts
		suite.Assert().Equal(2, preventativeCount) // WO001, WO005
		suite.Assert().Equal(2, correctiveCount)   // WO002, WO006
		suite.Assert().Equal(1, emergencyCount)    // WO003
		suite.Assert().Equal(1, inspectionCount)   // WO004
	})

	suite.Run("should return 0 for all types when machine has no maintenance records", func() {
		preventativeCount, correctiveCount, emergencyCount, inspectionCount, err := suite.repo.CountByWorkOrderType(ctx, TestMachineOne)
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

func (suite *MaintenanceRepositoryTestSuite) TestSearchByMachine() {
	ctx := context.Background()

	suite.Run("should search by work order number within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO004")
		maintenance1.ReportedBy = "Alice Tech"
		maintenance1.WorkOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineTwo, "WO004")
		maintenance2.ReportedBy = "Bob Tech"
		maintenance2.WorkOrderType = "Inspection"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "WO004", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("WO004", results[0].WorkOrderNumber)
		suite.Assert().Equal(TestMachineOne, results[0].MachineSerialNumber)
		suite.Assert().Equal("Alice Tech", results[0].ReportedBy)
	})

	suite.Run("should search by reported by name within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO005")
		maintenance1.ReportedBy = "Charlie Tech"
		maintenance1.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO006")
		maintenance2.ReportedBy = "David Tech"
		maintenance2.WorkOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance(TestMachineTwo, "WO007")
		maintenance3.ReportedBy = "Charlie Tech"
		maintenance3.WorkOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "Charlie", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("WO005", results[0].WorkOrderNumber)
		suite.Assert().Equal(TestMachineOne, results[0].MachineSerialNumber)
		suite.Assert().Equal("Charlie Tech", results[0].ReportedBy)
	})

	suite.Run("should search by work order type within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO008")
		maintenance1.ReportedBy = "Eve Tech"
		maintenance1.WorkOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO009")
		maintenance2.ReportedBy = "Frank Tech"
		maintenance2.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance(TestMachineTwo, "WO010")
		maintenance3.ReportedBy = "Grace Tech"
		maintenance3.WorkOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "Emergency", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("WO008", results[0].WorkOrderNumber)
		suite.Assert().Equal(TestMachineOne, results[0].MachineSerialNumber)
		suite.Assert().Equal("Emergency", results[0].WorkOrderType)
	})

	suite.Run("should search with pagination", func() {
		for i := 1; i <= 5; i++ {
			maintenance := testutils.CreateMaintenance(TestMachineOne, fmt.Sprintf("WO%03d", i+10))
			maintenance.ReportedBy = "Test Tech"
			maintenance.WorkOrderType = "Preventive"
			maintenance.ActionTaken = fmt.Sprintf("Action %d", i)
			suite.Require().NoError(suite.repo.Create(ctx, maintenance))
		}

		// Search with limit 2, offset 0
		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "Test", &maintenance.ListOptions{
			Limit:  2,
			Offset: 0,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
		suite.Assert().Equal(TestMachineOne, results[0].MachineSerialNumber)
		suite.Assert().Equal(TestMachineOne, results[1].MachineSerialNumber)

		// Search with limit 2, offset 2
		results, err = suite.repo.SearchByMachine(ctx, TestMachineOne, "Test", &maintenance.ListOptions{
			Limit:  2,
			Offset: 2,
			Sort:   maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
		suite.Assert().Equal(TestMachineOne, results[0].MachineSerialNumber)
		suite.Assert().Equal(TestMachineOne, results[1].MachineSerialNumber)
	})

	suite.Run("should return empty results for non-matching query", func() {
		maintenanceRecord := testutils.CreateMaintenance(TestMachineOne, "WO020")
		maintenanceRecord.ReportedBy = "Henry Tech"
		maintenanceRecord.WorkOrderType = "Inspection"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Search for non-matching term
		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "NonExistentTerm", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 0)
	})

	suite.Run("should handle different sort orders", func() {
		// Create maintenance records with different dates
		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO022")
		maintenance1.ReportedBy = "Jack Tech"
		maintenance1.WorkOrderType = "Preventive"
		maintenance1.WorkOrderDate = baseTime.AddDate(0, 0, 1) // Jan 2nd
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO023")
		maintenance2.ReportedBy = "Jack Tech"
		maintenance2.WorkOrderType = "Corrective"
		maintenance2.WorkOrderDate = baseTime.AddDate(0, 0, 3) // Jan 4th
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO024")
		maintenance3.ReportedBy = "Jack Tech"
		maintenance3.WorkOrderType = "Emergency"
		maintenance3.WorkOrderDate = baseTime.AddDate(0, 0, 2) // Jan 3rd
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		// Search with work order date descending (newest first)
		results, err := suite.repo.SearchByMachine(ctx, TestMachineOne, "Jack", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		suite.Assert().Equal("WO023", results[0].WorkOrderNumber) // Jan 4th
		suite.Assert().Equal("WO024", results[1].WorkOrderNumber) // Jan 3rd
		suite.Assert().Equal("WO022", results[2].WorkOrderNumber) // Jan 2nd

		// Search with work order date ascending (oldest first)
		results, err = suite.repo.SearchByMachine(ctx, TestMachineOne, "Jack", &maintenance.ListOptions{
			Limit: 10,
			Sort:  maintenance.SortOrderWorkOrderDateAsc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		suite.Assert().Equal("WO022", results[0].WorkOrderNumber) // Jan 2nd
		suite.Assert().Equal("WO024", results[1].WorkOrderNumber) // Jan 3rd
		suite.Assert().Equal("WO023", results[2].WorkOrderNumber) // Jan 4th
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestCountSearchByMachine() {
	ctx := context.Background()

	suite.Run("should count search results correctly for specific machine", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO030")
		maintenance1.ReportedBy = "Alpha Tech"
		maintenance1.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		// Create test maintenance records for different machines
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO031")
		maintenance2.ReportedBy = "Beta Tech"
		maintenance2.WorkOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance(TestMachineOne, "WO032")
		maintenance3.ReportedBy = "Gamma Tech"
		maintenance3.WorkOrderType = "Emergency"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		maintenance4 := testutils.CreateMaintenance(TestMachineTwo, "WO033")
		maintenance4.ReportedBy = "Delta Tech"
		maintenance4.WorkOrderType = "Inspection"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance4))

		// Create a record for a different machine
		maintenance5 := testutils.CreateMaintenance(TestMachineTwo, "WO034")
		maintenance5.ReportedBy = "Echo Tech"
		maintenance5.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance5))

		count, err := suite.repo.CountSearchByMachine(ctx, TestMachineOne, "Tech")
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should count search results by work order number within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO034")
		maintenance1.ReportedBy = "Echo Tech"
		maintenance1.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineTwo, "WO034")
		maintenance2.ReportedBy = "Foxtrot Tech"
		maintenance2.WorkOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		count, err := suite.repo.CountSearchByMachine(ctx, TestMachineOne, "WO034")
		suite.Require().NoError(err)
		suite.Assert().Equal(1, count)
	})

	suite.Run("should return zero count for non-matching query", func() {
		maintenanceRecord := testutils.CreateMaintenance(TestMachineOne, "WO035")
		maintenanceRecord.ReportedBy = "Golf Tech"
		maintenanceRecord.WorkOrderType = "Inspection"
		suite.Require().NoError(suite.repo.Create(ctx, maintenanceRecord))

		// Count search results for non-matching term
		count, err := suite.repo.CountSearchByMachine(ctx, TestMachineOne, "NonExistentTerm")
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})

	suite.Run("should count search results by reported by name within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO037")
		maintenance1.ReportedBy = "India Tech"
		maintenance1.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO038")
		maintenance2.ReportedBy = "India Tech"
		maintenance2.WorkOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		count, err := suite.repo.CountSearchByMachine(ctx, TestMachineOne, "India")
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count)
	})

	suite.Run("should count search results by work order type within machine scope", func() {
		maintenance1 := testutils.CreateMaintenance(TestMachineOne, "WO040")
		maintenance1.ReportedBy = "Juliet Tech"
		maintenance1.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance1))

		// Create test maintenance records
		maintenance2 := testutils.CreateMaintenance(TestMachineOne, "WO041")
		maintenance2.ReportedBy = "Kilo Tech"
		maintenance2.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance2))

		maintenance3 := testutils.CreateMaintenance(TestMachineTwo, "WO042")
		maintenance3.ReportedBy = "Lima Tech"
		maintenance3.WorkOrderType = "Corrective"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance3))

		maintenance4 := testutils.CreateMaintenance(TestMachineTwo, "WO043")
		maintenance4.ReportedBy = "Mike Tech"
		maintenance4.WorkOrderType = "Preventive"
		suite.Require().NoError(suite.repo.Create(ctx, maintenance4))

		count, err := suite.repo.CountSearchByMachine(ctx, TestMachineOne, "Preventive")
		suite.Require().NoError(err)
		suite.Assert().Equal(2, count)
	})
}

// TestMaintenanceRepositoryTestSuite runs the test suite
func TestMaintenanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceRepositoryTestSuite))
}
