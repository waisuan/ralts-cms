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
	suite.repo = machine.NewRepository(deps.PostgresClient)
}

func (suite *MachineRepositoryTestSuite) TearDownTest() {
	// Clear the machines table for PostgreSQL
	ctx := context.Background()
	_, err := suite.deps.PostgresClient.Exec(ctx, "DELETE FROM machines")
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
		suite.Assert().Greater(machine.ID, 0) // PostgreSQL should return an ID
	})

	suite.Run("should fail when creating duplicate machine", func() {
		machine := testutils.CreateMachine("MACHINE002")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Try to create the same machine again
		duplicateMachine := testutils.CreateMachine("MACHINE002")
		err = suite.repo.Create(ctx, duplicateMachine)
		suite.Require().Error(err)
		// PostgreSQL will return a unique constraint violation error
		suite.Assert().Contains(err.Error(), "duplicate key")
	})

	suite.Run("should create machine with minimal fields", func() {
		machine := testutils.CreateMinimalMachine("MACHINE003")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, machine.SerialNumber)
		suite.Require().NoError(err)
		suite.Assert().Equal("Minimal Customer", retrieved.Customer)
		suite.Assert().Equal("Active", retrieved.Status)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should create machine with custom fields", func() {
		machine := testutils.CreateMachineWithCustomFields("MACHINE004", "Custom Customer", "Maintenance")

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, machine.SerialNumber)
		suite.Require().NoError(err)
		suite.Assert().Equal("Custom Customer", retrieved.Customer)
		suite.Assert().Equal("Maintenance", retrieved.Status)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
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
		suite.Assert().Contains(err.Error(), "machine not found")
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
		suite.Assert().WithinDuration(originalCreatedAt, retrieved.CreatedAt, 1*time.Second)
		// UpdatedAt should be different or at least not older
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt))
	})

	suite.Run("should return error when updating non-existent machine", func() {
		machine := testutils.CreateMachine("MACHINE008")
		machine.Customer = "New Customer"

		err := suite.repo.Update(ctx, machine)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "machine not found")
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
		suite.Assert().True(retrieved.UpdatedAt.After(originalUpdatedAt) || retrieved.UpdatedAt.Equal(originalUpdatedAt))
	})
}

func (suite *MachineRepositoryTestSuite) TestList() {
	ctx := context.Background()

	suite.Run("should return empty list when no machines exist", func() {
		machines, err := suite.repo.List(ctx, 50)
		suite.Require().NoError(err)
		suite.Assert().Empty(machines)
		suite.Assert().Len(machines, 0)
	})

	suite.Run("should list all machines when limit is sufficient", func() {
		// Create multiple machines
		machine1 := testutils.CreateMachine("LIST001")
		machine2 := testutils.CreateMachine("LIST002")
		machine3 := testutils.CreateMachine("LIST003")

		err := suite.repo.Create(ctx, machine1)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, machine2)
		suite.Require().NoError(err)
		err = suite.repo.Create(ctx, machine3)
		suite.Require().NoError(err)

		machines, err := suite.repo.List(ctx, 50)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)

		// Verify all machines are returned
		serialNumbers := make(map[string]bool)
		for _, m := range machines {
			serialNumbers[m.SerialNumber] = true
		}
		suite.Assert().True(serialNumbers["LIST001"])
		suite.Assert().True(serialNumbers["LIST002"])
		suite.Assert().True(serialNumbers["LIST003"])
	})

	suite.Run("should respect limit parameter", func() {
		// Create 5 machines
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("LIMIT%03d", i))
			err := suite.repo.Create(ctx, machine)
			suite.Require().NoError(err)
		}

		// Test with limit of 3
		machines, err := suite.repo.List(ctx, 3)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)

		// Test with limit of 1
		machines, err = suite.repo.List(ctx, 1)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
	})
}

func (suite *MachineRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing machine", func() {
		machine := testutils.CreateMachine("DELETE001")
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Verify machine exists
		_, err = suite.repo.GetBySerialNumber(ctx, "DELETE001")
		suite.Require().NoError(err)

		// Delete the machine
		err = suite.repo.Delete(ctx, "DELETE001")
		suite.Require().NoError(err)

		// Verify machine is deleted
		_, err = suite.repo.GetBySerialNumber(ctx, "DELETE001")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "machine not found")
	})

	suite.Run("should return error when deleting non-existent machine", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "machine not found")
	})
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
