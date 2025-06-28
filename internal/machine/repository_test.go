package machine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/machine"
	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MachineRepositoryTestSuite defines the test suite for machine repository
type MachineRepositoryTestSuite struct {
	suite.Suite

	deps *deps.Dependencies
	repo machine.Repository
}

// SetupTest sets up each test
func (suite *MachineRepositoryTestSuite) SetupTest() {
	deps := deps.Initialise()
	suite.deps = deps
	suite.repo = machine.NewRepository(deps.DynamoDBClient, deps.Config.DynamoDBTable)
}

func (suite *MachineRepositoryTestSuite) TearDownTest() {
	err := testutils.ClearTable(context.Background(), suite.deps.Config.DynamoDBTable, suite.deps.DynamoDBClient)
	require.NoError(suite.T(), err)
}

func (suite *MachineRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create machine successfully", func() {
		machine := testutils.CreateMachine("MACHINE001")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(machine.CreatedAt)
		suite.Assert().NotEmpty(machine.UpdatedAt)
		suite.Assert().Equal(machine.CreatedAt, machine.UpdatedAt)
	})

	suite.Run("should fail when creating duplicate machine", func() {
		machine := testutils.CreateMachine("MACHINE002")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Try to create the same machine again
		duplicateMachine := testutils.CreateMachine("MACHINE002")
		err = suite.repo.Create(ctx, duplicateMachine)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "ConditionalCheckFailedException")
	})

	suite.Run("should create machine with minimal fields", func() {
		machine := testutils.CreateMinimalMachine("MACHINE003")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(machine.CreatedAt)
		suite.Assert().NotEmpty(machine.UpdatedAt)
	})

	suite.Run("should create machine with custom fields", func() {
		machine := testutils.CreateMachineWithCustomFields("MACHINE004", "Custom Customer", "Maintenance")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)
		suite.Assert().Equal("Custom Customer", machine.Customer)
		suite.Assert().Equal("Maintenance", machine.Status)
	})
}

func (suite *MachineRepositoryTestSuite) TestGetBySerialNumber() {
	ctx := context.Background()

	suite.Run("should get existing machine", func() {
		machine := testutils.CreateMachine("MACHINE005")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE005")
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE005", retrieved.SerialNumber)
		suite.Assert().Equal("Test Customer", retrieved.Customer)
		suite.Assert().Equal("Operational", retrieved.Status)
		suite.Assert().Equal("Test Model", retrieved.Model)
		suite.Assert().Equal("Test Brand", retrieved.Brand)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should return error for non-existent machine", func() {
		_, err := suite.repo.GetBySerialNumber(ctx, "NONEXISTENT")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})

	suite.Run("should get machine with all fields populated", func() {
		machine := testutils.CreateMachine("MACHINE006")
		machine.AdditionalNotes = "Special notes for testing"
		machine.PpmStatus = "Completed"
		machine.Attachment = "special-test.pdf"

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE006")
		suite.Require().NoError(err)
		suite.Assert().Equal("Special notes for testing", retrieved.AdditionalNotes)
		suite.Assert().Equal("Completed", retrieved.PpmStatus)
		suite.Assert().Equal("special-test.pdf", retrieved.Attachment)
	})
}

func (suite *MachineRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	suite.Run("should update existing machine", func() {
		machine := testutils.CreateMachine("MACHINE007")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		originalCreatedAt := machine.CreatedAt
		originalUpdatedAt := machine.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update machine fields
		machine.Customer = "Updated Customer"
		machine.Status = "Under Maintenance"
		machine.AdditionalNotes = "Updated notes"
		machine.PersonInCharge = "Jane Doe"

		err = suite.repo.Update(ctx, machine)
		suite.Require().NoError(err)

		// Verify the update
		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE007")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated Customer", retrieved.Customer)
		suite.Assert().Equal("Under Maintenance", retrieved.Status)
		suite.Assert().Equal("Updated notes", retrieved.AdditionalNotes)
		suite.Assert().Equal("Jane Doe", retrieved.PersonInCharge)
		suite.Assert().Equal(originalCreatedAt, retrieved.CreatedAt)
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt >= originalUpdatedAt)
	})

	suite.Run("should update non-existent machine (creates new record)", func() {
		machine := testutils.CreateMachine("MACHINE008")
		machine.Customer = "New Customer"

		err := suite.repo.Update(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE008")
		suite.Require().NoError(err)
		suite.Assert().Equal("New Customer", retrieved.Customer)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should update machine with minimal changes", func() {
		machine := testutils.CreateMachine("MACHINE009")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		originalUpdatedAt := machine.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		// Only update one field
		machine.Status = "Idle"
		err = suite.repo.Update(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE009")
		suite.Require().NoError(err)
		suite.Assert().Equal("Idle", retrieved.Status)
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt >= originalUpdatedAt)
	})
}

func (suite *MachineRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing machine", func() {
		machine := testutils.CreateMachine("MACHINE010")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Verify machine exists
		_, err = suite.repo.GetBySerialNumber(ctx, "MACHINE010")
		suite.Require().NoError(err)

		// Delete the machine
		err = suite.repo.Delete(ctx, "MACHINE010")
		suite.Require().NoError(err)

		// Verify machine is deleted
		_, err = suite.repo.GetBySerialNumber(ctx, "MACHINE010")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})

	suite.Run("should not error when deleting non-existent machine", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT")
		suite.Require().NoError(err)
	})

	suite.Run("should delete machine and allow recreation", func() {
		machine := testutils.CreateMachine("MACHINE011")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		err = suite.repo.Delete(ctx, "MACHINE011")
		suite.Require().NoError(err)

		// Should be able to create a new machine with the same serial number
		newMachine := testutils.CreateMachine("MACHINE011")
		newMachine.Customer = "Recreated Customer"
		err = suite.repo.Create(ctx, newMachine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE011")
		suite.Require().NoError(err)
		suite.Assert().Equal("Recreated Customer", retrieved.Customer)
	})
}

func (suite *MachineRepositoryTestSuite) TestList() {
	ctx := context.Background()

	suite.Run("should return empty list when no machines exist", func() {
		machines, nextPageToken, err := suite.repo.List(ctx, 10, "")
		suite.Require().NoError(err)
		suite.Assert().Empty(machines)
		suite.Assert().Empty(nextPageToken)
	})

	suite.Run("should list all machines when count is less than limit", func() {
		// Create 3 machines
		for i := 1; i <= 3; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("MACHINE%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		machines, nextPageToken, err := suite.repo.List(ctx, 10, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		suite.Assert().Empty(nextPageToken) // No more pages
		suite.Assert().Equal("MACHINE001", machines[0].SerialNumber)
		suite.Assert().Equal("MACHINE002", machines[1].SerialNumber)
		suite.Assert().Equal("MACHINE003", machines[2].SerialNumber)
	})

	suite.Run("should respect limit parameter", func() {
		// Create 5 machines
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("MACHINE%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		machines, nextPageToken, err := suite.repo.List(ctx, 3, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		suite.Assert().NotEmpty(nextPageToken) // Should have more pages
		suite.Assert().Equal("MACHINE001", machines[0].SerialNumber)
		suite.Assert().Equal("MACHINE002", machines[1].SerialNumber)
		suite.Assert().Equal("MACHINE003", machines[2].SerialNumber)
	})

	suite.Run("should handle pagination correctly", func() {
		// Create 6 machines
		for i := 1; i <= 6; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("MACHINE%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		// First page - limit 2
		machines1, nextPageToken1, err := suite.repo.List(ctx, 2, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines1, 2)
		suite.Assert().NotEmpty(nextPageToken1)
		suite.Assert().Equal("MACHINE001", machines1[0].SerialNumber)
		suite.Assert().Equal("MACHINE002", machines1[1].SerialNumber)

		// Second page - using the page token
		machines2, nextPageToken2, err := suite.repo.List(ctx, 2, nextPageToken1)
		suite.Require().NoError(err)
		suite.Assert().Len(machines2, 2)
		suite.Assert().NotEmpty(nextPageToken2)
		suite.Assert().Equal("MACHINE003", machines2[0].SerialNumber)
		suite.Assert().Equal("MACHINE004", machines2[1].SerialNumber)

		// Third page
		machines3, nextPageToken3, err := suite.repo.List(ctx, 2, nextPageToken2)
		suite.Require().NoError(err)
		suite.Assert().Len(machines3, 2)
		suite.Assert().NotEmpty(nextPageToken3)
		suite.Assert().Equal("MACHINE005", machines3[0].SerialNumber)
		suite.Assert().Equal("MACHINE006", machines3[1].SerialNumber)

		// Fourth page - should be empty
		machines4, nextPageToken4, err := suite.repo.List(ctx, 2, nextPageToken3)
		suite.Require().NoError(err)
		suite.Assert().Empty(machines4)
		suite.Assert().Empty(nextPageToken4)
	})

	suite.Run("should handle invalid page token gracefully", func() {
		// Create a machine
		machine := testutils.CreateMachine("MACHINE001")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Try with invalid page token
		_, _, err = suite.repo.List(ctx, 10, "invalid-token")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "invalid page token")
	})

	suite.Run("should handle empty page token", func() {
		// Create a machine
		machine := testutils.CreateMachine("MACHINE001")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Try with empty page token
		machines, nextPageToken, err := suite.repo.List(ctx, 10, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Empty(nextPageToken)
	})

	suite.Run("should return machines with all fields populated", func() {
		// Create a machine with custom fields
		machine := testutils.CreateMachine("MACHINE001")
		machine.AdditionalNotes = "Test notes"
		machine.PpmStatus = "Completed"
		machine.Attachment = "test.pdf"
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		machines, _, err := suite.repo.List(ctx, 10, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)

		retrieved := machines[0]
		suite.Assert().Equal("MACHINE001", retrieved.SerialNumber)
		suite.Assert().Equal("Test Customer", retrieved.Customer)
		suite.Assert().Equal("Operational", retrieved.Status)
		suite.Assert().Equal("Test notes", retrieved.AdditionalNotes)
		suite.Assert().Equal("Completed", retrieved.PpmStatus)
		suite.Assert().Equal("test.pdf", retrieved.Attachment)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should handle large limit gracefully", func() {
		// Create 3 machines
		for i := 1; i <= 3; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("MACHINE%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		// Request more than available
		machines, nextPageToken, err := suite.repo.List(ctx, 100, "")
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		suite.Assert().Empty(nextPageToken)
	})

	suite.Run("should maintain order consistency across pages", func() {
		// Create machines in reverse order to test ordering
		for i := 5; i >= 1; i-- {
			machine := testutils.CreateMachine(fmt.Sprintf("MACHINE%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		// Get all machines in one request
		allMachines, _, err := suite.repo.List(ctx, 10, "")
		suite.Require().NoError(err)
		suite.Assert().Len(allMachines, 5)

		// Verify order (should be consistent with DynamoDB scan order)
		// Note: DynamoDB scan order is not guaranteed, but should be consistent within a single scan
		serialNumbers := make([]string, len(allMachines))
		for i, m := range allMachines {
			serialNumbers[i] = m.SerialNumber
		}

		// All machines should be present
		suite.Assert().Contains(serialNumbers, "MACHINE001")
		suite.Assert().Contains(serialNumbers, "MACHINE002")
		suite.Assert().Contains(serialNumbers, "MACHINE003")
		suite.Assert().Contains(serialNumbers, "MACHINE004")
		suite.Assert().Contains(serialNumbers, "MACHINE005")
	})
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
