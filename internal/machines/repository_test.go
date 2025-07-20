package machines_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MachineRepositoryTestSuite defines the test suite for machine repository
type MachineRepositoryTestSuite struct {
	suite.Suite

	deps *deps.Dependencies
	repo machines.Repository
}

// SetupTest sets up each test
func (suite *MachineRepositoryTestSuite) SetupTest() {
	deps := deps.Initialise()
	suite.deps = deps
	suite.repo = machines.NewRepository(deps.PostgresClient)
}

func (suite *MachineRepositoryTestSuite) TearDownSubTest() {
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
		machine.PpmStatus = string(machines.PPMStatusDue)
		machine.Attachment = "special-test.pdf"

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE006")
		suite.Require().NoError(err)
		suite.Assert().Equal("Special notes for testing", retrieved.AdditionalNotes)
		suite.Assert().Equal(string(machines.PPMStatusDue), retrieved.PpmStatus)
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
		options := machines.DefaultListOptions()
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Empty(machines)
		suite.Assert().Len(machines, 0)
	})

	suite.Run("should list all machines when limit is sufficient", func() {
		machine1 := testutils.CreateMachine("LIST001")
		machine2 := testutils.CreateMachine("LIST002")
		machine3 := testutils.CreateMachine("LIST003")
		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 50, Offset: 0, Sort: machines.SortOrderCreatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		serialNumbers := map[string]bool{}
		for _, m := range machines {
			serialNumbers[m.SerialNumber] = true
		}
		suite.Assert().True(serialNumbers["LIST001"])
		suite.Assert().True(serialNumbers["LIST002"])
		suite.Assert().True(serialNumbers["LIST003"])
	})

	suite.Run("should respect limit parameter", func() {
		for i := 1; i <= 5; i++ {
			m := testutils.CreateMachine(fmt.Sprintf("LIMIT%03d", i))
			suite.Require().NoError(suite.repo.Create(ctx, m))
		}
		options := &machines.ListOptions{Limit: 3, Offset: 0, Sort: machines.SortOrderCreatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
	})

	suite.Run("should support pagination with offset", func() {
		for i := 1; i <= 5; i++ {
			m := testutils.CreateMachine(fmt.Sprintf("PAGE%03d", i))
			suite.Require().NoError(suite.repo.Create(ctx, m))
		}
		options := &machines.ListOptions{Limit: 2, Offset: 0, Sort: machines.SortOrderCreatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 2)
		suite.Assert().Equal("PAGE005", machines[0].SerialNumber)
		suite.Assert().Equal("PAGE004", machines[1].SerialNumber)

		options.Offset = 2
		machines, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 2)
		suite.Assert().Equal("PAGE003", machines[0].SerialNumber)
		suite.Assert().Equal("PAGE002", machines[1].SerialNumber)

		options.Offset = 4
		machines, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("PAGE001", machines[0].SerialNumber)

		options.Offset = 10
		machines, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 0)
	})

	suite.Run("should support sorting by created_at ascending", func() {
		machine1 := testutils.CreateMachine("SORT001")
		time.Sleep(10 * time.Millisecond)
		machine2 := testutils.CreateMachine("SORT002")
		time.Sleep(10 * time.Millisecond)
		machine3 := testutils.CreateMachine("SORT003")
		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderCreatedAtAsc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		suite.Assert().Equal("SORT001", machines[0].SerialNumber)
		suite.Assert().Equal("SORT002", machines[1].SerialNumber)
		suite.Assert().Equal("SORT003", machines[2].SerialNumber)
	})

	suite.Run("should support sorting by created_at descending", func() {
		machine1 := testutils.CreateMachine("SORT004")
		time.Sleep(10 * time.Millisecond)
		machine2 := testutils.CreateMachine("SORT005")
		time.Sleep(10 * time.Millisecond)
		machine3 := testutils.CreateMachine("SORT006")
		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderCreatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		suite.Assert().Equal("SORT006", machines[0].SerialNumber)
		suite.Assert().Equal("SORT005", machines[1].SerialNumber)
		suite.Assert().Equal("SORT004", machines[2].SerialNumber)
	})

	suite.Run("should use default options when nil is passed", func() {
		machine := testutils.CreateMachine("DEFAULT001")
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		machines, err := suite.repo.List(ctx, nil)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("DEFAULT001", machines[0].SerialNumber)
	})

	suite.Run("should handle invalid sort order gracefully", func() {
		m := testutils.CreateMachine("INVALID001")
		suite.Require().NoError(suite.repo.Create(ctx, m))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: "invalid_sort"}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("INVALID001", machines[0].SerialNumber)
	})

	suite.Run("should return empty list when no machines are due for PPM", func() {
		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Empty(machines)
	})

	suite.Run("should return machines that are due for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM001")
		machine.PpmDate = time.Now()
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("DUEPPM001", machines[0].SerialNumber)
	})

	suite.Run("should return machines that are overdue for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM002")
		machine.PpmDate = time.Now().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("DUEPPM002", machines[0].SerialNumber)
	})

	suite.Run("should return machines that are due in 2 weeks for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM003")
		machine.PpmDate = time.Now().AddDate(0, 0, 10)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 1)
		suite.Assert().Equal("DUEPPM003", machines[0].SerialNumber)
	})

	suite.Run("should not return machines that are due in more than 2 weeks for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM004")
		machine.PpmDate = time.Now().AddDate(0, 0, 30)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Empty(machines)
	})

	suite.Run("should return multiple machines due for PPM ordered by creation date", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("DUEPPM005")
		machine1.PpmDate = time.Now().AddDate(0, 0, 5) // Due in 5 days
		suite.Require().NoError(suite.repo.Create(ctx, machine1))

		machine2 := testutils.CreateMachine("DUEPPM006")
		machine2.PpmDate = time.Now().AddDate(0, 0, -2) // Overdue by 2 days
		suite.Require().NoError(suite.repo.Create(ctx, machine2))

		machine3 := testutils.CreateMachine("DUEPPM007")
		machine3.PpmDate = time.Now().AddDate(0, 0, 12) // Due in 12 days
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)

		suite.Assert().Equal("DUEPPM007", machines[0].SerialNumber)
		suite.Assert().Equal("DUEPPM006", machines[1].SerialNumber)
		suite.Assert().Equal("DUEPPM005", machines[2].SerialNumber)
	})

	suite.Run("should respect limit and offset when filtering for PPM due", func() {
		// Create multiple machines due for PPM
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("DUEPPM%03d", i))
			machine.PpmDate = time.Now().AddDate(0, 0, i)
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		options := &machines.ListOptions{Limit: 2, Offset: 0, DuePPMOnly: true}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 2)

		options.Offset = 2
		machines, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 2)

		options.Offset = 4
		machines, err = suite.repo.List(ctx, options)
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

func (suite *MachineRepositoryTestSuite) TestCount() {
	ctx := context.Background()

	suite.Run("should return count of machines", func() {
		suite.repo.Create(ctx, testutils.CreateMachine("COUNT001"))
		suite.repo.Create(ctx, testutils.CreateMachine("COUNT002"))
		suite.repo.Create(ctx, testutils.CreateMachine("COUNT003"))

		count, err := suite.repo.Count(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(3, count)
	})

	suite.Run("should return 0 when no machines exist", func() {
		count, err := suite.repo.Count(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})
}

func (suite *MachineRepositoryTestSuite) TestDuePPM() {
	ctx := context.Background()

	suite.Run("should return machines that are overdue for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM001")
		machine.PpmDate = time.Now().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		machinesList, err := suite.repo.DuePPM(ctx)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1)
		suite.Assert().Equal("DUEPPM001", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusOverdue), machinesList[0].PpmStatus)
	})

	suite.Run("should return machines that are due for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM002")
		machine.PpmDate = time.Now()
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		machinesList, err := suite.repo.DuePPM(ctx)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1)
		suite.Assert().Equal("DUEPPM002", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusDue), machinesList[0].PpmStatus)
	})

	suite.Run("should return machines that are almost due for PPM", func() {
		machine := testutils.CreateMachine("DUEPPM003")
		machine.PpmDate = time.Now().AddDate(0, 0, 10)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		machinesList, err := suite.repo.DuePPM(ctx)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1)
		suite.Assert().Equal("DUEPPM003", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusAlmostDue), machinesList[0].PpmStatus)
	})

	suite.Run("should return empty list when no machines exist", func() {
		machinesList, err := suite.repo.DuePPM(ctx)
		suite.Require().NoError(err)
		suite.Assert().Empty(machinesList)
	})

	suite.Run("should return empty list when no machines are due for PPM", func() {
		machinesList, err := suite.repo.DuePPM(ctx)
		suite.Require().NoError(err)
		suite.Assert().Empty(machinesList)
	})
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
