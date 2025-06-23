package machine_test

import (
	"context"
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

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
