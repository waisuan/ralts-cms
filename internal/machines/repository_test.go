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

	suite.Run("should return machine with PPM status", func() {
		machine := testutils.CreateMachine("MACHINE006")
		machine.PpmDate = time.Now().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		retrieved, err := suite.repo.GetBySerialNumber(ctx, "MACHINE006")
		suite.Require().NoError(err)
		suite.Assert().Equal(string(machines.PPMStatusOverdue), retrieved.PpmStatus)
	})

	suite.Run("should return error for non-existent machine", func() {
		_, err := suite.repo.GetBySerialNumber(ctx, "NONEXISTENT")
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "machine not found")
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

	suite.Run("should return machines with PPM status", func() {
		machine1 := testutils.CreateMachine("LIST004")
		machine1.PpmDate = time.Now().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, machine1))

		machine2 := testutils.CreateMachine("LIST005")
		machine2.PpmDate = time.Now()
		suite.Require().NoError(suite.repo.Create(ctx, machine2))

		machine3 := testutils.CreateMachine("LIST006")
		machine3.PpmDate = time.Now().AddDate(0, 0, 10)
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 50, Offset: 0, Sort: machines.SortOrderCreatedAtAsc}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 3)
		suite.Assert().Equal(string(machines.PPMStatusOverdue), machinesList[0].PpmStatus)
		suite.Assert().Equal(string(machines.PPMStatusDue), machinesList[1].PpmStatus)
		suite.Assert().Equal(string(machines.PPMStatusAlmostDue), machinesList[2].PpmStatus)
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

	suite.Run("should filter by overdue PPM status", func() {
		// Create machines with different PPM dates
		overdueMachine := testutils.CreateMachine("OVERDUE001")
		overdueMachine.PpmDate = time.Now().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE001")
		dueMachine.PpmDate = time.Now() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE001")
		almostDueMachine.PpmDate = time.Now().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now (not due)
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Should return only overdue machines

		// Verify we got the expected machine
		suite.Assert().Equal("OVERDUE001", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusOverdue), machinesList[0].PpmStatus)
	})

	suite.Run("should filter by due PPM status", func() {
		// Create machines with different PPM dates
		overdueMachine := testutils.CreateMachine("OVERDUE002")
		overdueMachine.PpmDate = time.Now().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE002")
		dueMachine.PpmDate = time.Now() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE002")
		almostDueMachine.PpmDate = time.Now().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusDue,
		}

		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Should return only due machines

		// Verify we got the expected machine
		suite.Assert().Equal("DUE002", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusDue), machinesList[0].PpmStatus)
	})

	suite.Run("should filter by almost due PPM status", func() {
		// Create machines with different PPM dates
		overdueMachine := testutils.CreateMachine("OVERDUE003")
		overdueMachine.PpmDate = time.Now().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE003")
		dueMachine.PpmDate = time.Now() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE003")
		almostDueMachine.PpmDate = time.Now().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE003")
		futureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now (not due)
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusAlmostDue,
		}

		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Should return only almost due machines

		// Verify we got the expected machine
		suite.Assert().Equal("ALMOSTDUE003", machinesList[0].SerialNumber)
		suite.Assert().Equal(string(machines.PPMStatusAlmostDue), machinesList[0].PpmStatus)
	})

	suite.Run("should respect pagination with PpmStatusFilter", func() {
		// Create multiple overdue machines
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("OVERDUE%03d", i+10))
			machine.PpmDate = time.Now().AddDate(0, 0, -i) // All overdue
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		options := &machines.ListOptions{
			Limit:           2,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 2) // Should respect limit

		options.Offset = 2
		machinesList, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 2) // Should respect offset

		options.Offset = 4
		machinesList, err = suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Should return remaining overdue machines
	})

	suite.Run("should combine PpmStatusFilter with sorting", func() {
		// Create machines with different PPM dates and creation times
		overdueMachine1 := testutils.CreateMachine("OVERDUE020")
		overdueMachine1.PpmDate = time.Now().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine1))
		time.Sleep(10 * time.Millisecond)

		overdueMachine2 := testutils.CreateMachine("OVERDUE021")
		overdueMachine2.PpmDate = time.Now().AddDate(0, 0, -2)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine2))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		}

		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 2)
		// Should be sorted by creation time descending (newest first)
		suite.Assert().Equal("OVERDUE021", machinesList[0].SerialNumber)
		suite.Assert().Equal("OVERDUE020", machinesList[1].SerialNumber)
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

func (suite *MachineRepositoryTestSuite) TestCountByStatus() {
	ctx := context.Background()

	suite.Run("should return correct counts for different PPM statuses", func() {
		// Create machines with different PPM dates
		overdueMachine := testutils.CreateMachine("OVERDUE001")
		overdueMachine.PpmDate = time.Now().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE001")
		dueMachine.PpmDate = time.Now() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE001")
		almostDueMachine.PpmDate = time.Now().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now (not due)
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		overdueCount, dueCount, almostDueCount, err := suite.repo.CountByStatus(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(1), overdueCount)
		suite.Assert().Equal(int32(1), dueCount)
		suite.Assert().Equal(int32(1), almostDueCount)
	})

	suite.Run("should return zero counts when no machines exist", func() {
		overdueCount, dueCount, almostDueCount, err := suite.repo.CountByStatus(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(0), overdueCount)
		suite.Assert().Equal(int32(0), dueCount)
		suite.Assert().Equal(int32(0), almostDueCount)
	})

	suite.Run("should return zero counts when no machines are due", func() {
		futureMachine := testutils.CreateMachine("FUTURE002")
		futureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		overdueCount, dueCount, almostDueCount, err := suite.repo.CountByStatus(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(0), overdueCount)
		suite.Assert().Equal(int32(0), dueCount)
		suite.Assert().Equal(int32(0), almostDueCount)
	})

	suite.Run("should handle multiple machines with same status", func() {
		// Create multiple overdue machines
		for i := 1; i <= 3; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("OVERDUE%03d", i))
			machine.PpmDate = time.Now().AddDate(0, 0, -i) // Different overdue dates
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		// Create multiple due machines
		for i := 1; i <= 2; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("DUE%03d", i))
			machine.PpmDate = time.Now() // All due today
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		overdueCount, dueCount, almostDueCount, err := suite.repo.CountByStatus(ctx)
		suite.Require().NoError(err)
		suite.Assert().Equal(int32(3), overdueCount)
		suite.Assert().Equal(int32(2), dueCount)
		suite.Assert().Equal(int32(0), almostDueCount)
	})
}

func (suite *MachineRepositoryTestSuite) TestSearch() {
	ctx := context.Background()

	suite.Run("should search machines by text query", func() {
		// Create test machines with different searchable content
		hpPrinter := testutils.CreateMachine("HP001")
		hpPrinter.Brand = "HP"
		hpPrinter.Model = "LaserJet Pro"
		hpPrinter.Customer = "Office Supplies Inc"
		suite.Require().NoError(suite.repo.Create(ctx, hpPrinter))

		canonPrinter := testutils.CreateMachine("CANON001")
		canonPrinter.Brand = "Canon"
		canonPrinter.Model = "Pixma Pro"
		canonPrinter.Customer = "Print Shop"
		suite.Require().NoError(suite.repo.Create(ctx, canonPrinter))

		// Search for HP printers
		results, err := suite.repo.Search(ctx, "HP", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("HP001", results[0].SerialNumber)
		suite.Assert().Equal("HP", results[0].Brand)
	})

	suite.Run("should search across multiple fields", func() {
		// Create machine with searchable content in different fields
		machine := testutils.CreateMachine("SEARCH001")
		machine.Customer = "Tech Solutions"
		machine.Brand = "Brother"
		machine.Model = "Laser Printer"
		machine.District = "Downtown"
		machine.PersonInCharge = "John Tech"
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search for "Tech" which should match customer and person_in_charge
		results, err := suite.repo.Search(ctx, "Tech", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("SEARCH001", results[0].SerialNumber)
	})

	suite.Run("should return empty results for non-matching query", func() {
		// Create a machine
		machine := testutils.CreateMachine("NOMATCH001")
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search for something that doesn't exist
		results, err := suite.repo.Search(ctx, "nonexistent", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 0)
	})

	suite.Run("should respect pagination parameters", func() {
		// Create multiple machines
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("PAGE%03d", i))
			machine.Brand = "HP"
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		// Search with limit and offset
		results, err := suite.repo.Search(ctx, "HP", &machines.ListOptions{
			Limit:  2,
			Offset: 1,
			Sort:   machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
	})

	suite.Run("should filter by PPM status overdue", func() {
		// Create overdue machine
		overdueMachine := testutils.CreateMachine("OVERDUE001")
		overdueMachine.Brand = "HP"
		overdueMachine.PpmDate = time.Now().AddDate(0, 0, -1) // Yesterday
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		// Create future machine (not overdue)
		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.Brand = "HP"
		futureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		// Search for HP machines that are overdue
		results, err := suite.repo.Search(ctx, "HP", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("OVERDUE001", results[0].SerialNumber)
	})

	suite.Run("should filter by PPM status due today", func() {
		// Create machine due today
		dueMachine := testutils.CreateMachine("DUE001")
		dueMachine.Brand = "Canon"
		dueMachine.PpmDate = time.Now() // Today
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		// Create machine due tomorrow
		tomorrowMachine := testutils.CreateMachine("TOMORROW001")
		tomorrowMachine.Brand = "Canon"
		tomorrowMachine.PpmDate = time.Now().AddDate(0, 0, 1) // Tomorrow
		suite.Require().NoError(suite.repo.Create(ctx, tomorrowMachine))

		// Search for Canon machines due today
		results, err := suite.repo.Search(ctx, "Canon", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusDue,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("DUE001", results[0].SerialNumber)
	})

	suite.Run("should filter by PPM status almost due", func() {
		// Create machine almost due (within 2 weeks)
		almostDueMachine := testutils.CreateMachine("ALMOSTDUE001")
		almostDueMachine.Brand = "Brother"
		almostDueMachine.PpmDate = time.Now().AddDate(0, 0, 10) // 10 days from now
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		// Create machine far in future (not almost due)
		farFutureMachine := testutils.CreateMachine("FARFUTURE001")
		farFutureMachine.Brand = "Brother"
		farFutureMachine.PpmDate = time.Now().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, farFutureMachine))

		// Search for Brother machines almost due
		results, err := suite.repo.Search(ctx, "Brother", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderCreatedAtDesc,
			PpmStatusFilter: machines.PPMStatusAlmostDue,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("ALMOSTDUE001", results[0].SerialNumber)
	})

	suite.Run("should use default options when none provided", func() {
		// Create a machine
		machine := testutils.CreateMachine("DEFAULT001")
		machine.Brand = "HP"
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search without providing options
		results, err := suite.repo.Search(ctx, "HP", nil)
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("DEFAULT001", results[0].SerialNumber)
	})

	suite.Run("should handle case-insensitive search", func() {
		// Create machine with mixed case
		machine := testutils.CreateMachine("CASE001")
		machine.Brand = "HP"
		machine.Customer = "Office Supplies"
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search with lowercase
		results, err := suite.repo.Search(ctx, "office", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("CASE001", results[0].SerialNumber)
	})

	suite.Run("should handle empty search query", func() {
		// Create a machine
		machine := testutils.CreateMachine("EMPTY001")
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search with empty query
		results, err := suite.repo.Search(ctx, "", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		// Should return no results for empty query
		suite.Assert().Len(results, 0)
	})

	suite.Run("should handle special characters in search query", func() {
		// Create machine with special characters
		machine := testutils.CreateMachine("SPECIAL001")
		machine.Customer = "Tech & Solutions"
		machine.AdditionalNotes = "Maintenance (urgent)"
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search with special characters
		results, err := suite.repo.Search(ctx, "Tech Solutions", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("SPECIAL001", results[0].SerialNumber)
	})

	suite.Run("should return multiple results with proper ranking", func() {
		// Create multiple HP machines with different search relevance
		hpLaser := testutils.CreateMachine("HP001")
		hpLaser.Brand = "HP"
		hpLaser.Model = "LaserJet Pro"
		hpLaser.Customer = "Office Supplies"
		suite.Require().NoError(suite.repo.Create(ctx, hpLaser))

		hpInkjet := testutils.CreateMachine("HP002")
		hpInkjet.Brand = "HP"
		hpInkjet.Model = "Inkjet Pro"
		hpInkjet.Customer = "Print Shop"
		suite.Require().NoError(suite.repo.Create(ctx, hpInkjet))

		hpScanner := testutils.CreateMachine("HP003")
		hpScanner.Brand = "HP"
		hpScanner.Model = "Scanner Pro"
		hpScanner.Customer = "Document Center"
		suite.Require().NoError(suite.repo.Create(ctx, hpScanner))

		// Create a non-HP machine to ensure it's not included
		canonMachine := testutils.CreateMachine("CANON001")
		canonMachine.Brand = "Canon"
		canonMachine.Model = "Pixma"
		suite.Require().NoError(suite.repo.Create(ctx, canonMachine))

		// Search for HP machines
		results, err := suite.repo.Search(ctx, "HP", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)

		// Verify all results are HP machines
		for _, result := range results {
			suite.Assert().Equal("HP", result.Brand)
		}

		// Verify we got the expected serial numbers
		serialNumbers := make([]string, len(results))
		for i, result := range results {
			serialNumbers[i] = result.SerialNumber
		}
		suite.Assert().Contains(serialNumbers, "HP001")
		suite.Assert().Contains(serialNumbers, "HP002")
		suite.Assert().Contains(serialNumbers, "HP003")
		suite.Assert().NotContains(serialNumbers, "CANON001")
	})

	suite.Run("should respect limit when multiple results exist", func() {
		// Create multiple machines with same brand
		for i := 1; i <= 5; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("LIMIT%03d", i))
			machine.Brand = "Brother"
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		// Search with limit less than total matches
		results, err := suite.repo.Search(ctx, "Brother", &machines.ListOptions{
			Limit:  3,
			Offset: 0,
			Sort:   machines.SortOrderCreatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)

		// Verify all results are Brother machines
		for _, result := range results {
			suite.Assert().Equal("Brother", result.Brand)
		}
	})
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
