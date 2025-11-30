package machines_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ralts-cms/internal/machines"
	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MachineRepositoryTestSuite defines the test suite for machine repository
type MachineRepositoryTestSuite struct {
	suite.Suite

	db   *testutils.TestDatabase
	repo machines.Repository
}

// SetupTest sets up each test
func (suite *MachineRepositoryTestSuite) SetupSuite() {
	db := testutils.SetupTestDatabase(suite.T())
	suite.db = db
	suite.repo = machines.NewRepository(db.PostgresClient)
}

func (suite *MachineRepositoryTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *MachineRepositoryTestSuite) TearDownSubTest() {
	// Clear the machines table for PostgreSQL
	ctx := context.Background()
	err := suite.db.CleanupAllTables(ctx)
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
		suite.Assert().Greater(machine.ID, int64(0)) // PostgreSQL should return an ID
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
		machine.PpmDate = time.Now().UTC().AddDate(0, 0, -1)
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

	suite.Run("should handle NULL values in nullable fields without deserialization errors", func() {
		// Create a machine directly in the database with NULL values in the originally nullable fields
		_, err := suite.db.PostgresClient.Exec(ctx, `
			INSERT INTO machines (
				"serialNumber", customer, state, "accountType", model, status, brand,
				district, "personInCharge", "reportedBy", "additionalNotes", attachment,
				"tncDate", "ppmDate", "createdAt", "updatedAt"
			) VALUES (
				$1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL,
				NULL, NULL, $2, $3
			)
		`, "MACHINE_WITH_NULLS", time.Now().UTC(), time.Now().UTC())
		suite.Require().NoError(err)

		// This should not cause a deserialization error thanks to COALESCE
		machine, err := suite.repo.GetBySerialNumber(ctx, "MACHINE_WITH_NULLS")
		suite.Require().NoError(err)
		suite.Assert().NotNil(machine)

		// Verify that NULL values are converted to empty strings by COALESCE
		suite.Assert().Equal("", machine.Customer)
		suite.Assert().Equal("", machine.State)
		suite.Assert().Equal("", machine.AccountType)
		suite.Assert().Equal("", machine.Model)
		suite.Assert().Equal("", machine.Status)
		suite.Assert().Equal("", machine.Brand)
		suite.Assert().Equal("", machine.District)
		suite.Assert().Equal("", machine.PersonInCharge)
		suite.Assert().Equal("", machine.ReportedBy)
		suite.Assert().Equal("", machine.AdditionalNotes)
		suite.Assert().Equal("", machine.Attachment)
		suite.Assert().Equal(time.Time{}, machine.TncDate)
		suite.Assert().Equal(time.Time{}, machine.PpmDate)
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

		options := &machines.ListOptions{Limit: 50, Offset: 0, Sort: machines.SortOrderUpdatedAtDesc}
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
		machine1.PpmDate = time.Now().UTC().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, machine1))

		machine2 := testutils.CreateMachine("LIST005")
		machine2.PpmDate = time.Now().UTC()
		suite.Require().NoError(suite.repo.Create(ctx, machine2))

		machine3 := testutils.CreateMachine("LIST006")
		machine3.PpmDate = time.Now().UTC().AddDate(0, 0, 10)
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 50, Offset: 0, Sort: machines.SortOrderUpdatedAtAsc}
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
		options := &machines.ListOptions{Limit: 3, Offset: 0, Sort: machines.SortOrderUpdatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
	})

	suite.Run("should support pagination with offset", func() {
		for i := 1; i <= 5; i++ {
			m := testutils.CreateMachine(fmt.Sprintf("PAGE%03d", i))
			suite.Require().NoError(suite.repo.Create(ctx, m))
		}
		options := &machines.ListOptions{Limit: 2, Offset: 0, Sort: machines.SortOrderUpdatedAtDesc}
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

	suite.Run("should support sorting by updated_at ascending", func() {
		machine1 := testutils.CreateMachine("SORT001")
		machine2 := testutils.CreateMachine("SORT002")
		machine3 := testutils.CreateMachine("SORT003")
		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		// Update machines with different timestamps to test updated_at sorting
		time.Sleep(10 * time.Millisecond)
		machine3.Customer = "Updated Customer 3"
		suite.Require().NoError(suite.repo.Update(ctx, machine3))
		time.Sleep(10 * time.Millisecond)
		machine1.Customer = "Updated Customer 1"
		suite.Require().NoError(suite.repo.Update(ctx, machine1))
		time.Sleep(10 * time.Millisecond)
		machine2.Customer = "Updated Customer 2"
		suite.Require().NoError(suite.repo.Update(ctx, machine2))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderUpdatedAtAsc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		// Should be sorted by update time ascending (machine3 updated first, machine2 last)
		suite.Assert().Equal("SORT003", machines[0].SerialNumber)
		suite.Assert().Equal("SORT001", machines[1].SerialNumber)
		suite.Assert().Equal("SORT002", machines[2].SerialNumber)
	})

	suite.Run("should support sorting by updated_at descending", func() {
		machine1 := testutils.CreateMachine("SORT004")
		machine2 := testutils.CreateMachine("SORT005")
		machine3 := testutils.CreateMachine("SORT006")
		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		// Update machines with different timestamps to test updated_at sorting
		time.Sleep(10 * time.Millisecond)
		machine1.Customer = "Updated Customer 1"
		suite.Require().NoError(suite.repo.Update(ctx, machine1))
		time.Sleep(10 * time.Millisecond)
		machine3.Customer = "Updated Customer 3"
		suite.Require().NoError(suite.repo.Update(ctx, machine3))
		time.Sleep(10 * time.Millisecond)
		machine2.Customer = "Updated Customer 2"
		suite.Require().NoError(suite.repo.Update(ctx, machine2))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderUpdatedAtDesc}
		machines, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machines, 3)
		// Should be sorted by update time descending (machine2 updated last, machine1 first)
		suite.Assert().Equal("SORT005", machines[0].SerialNumber)
		suite.Assert().Equal("SORT006", machines[1].SerialNumber)
		suite.Assert().Equal("SORT004", machines[2].SerialNumber)
	})

	suite.Run("should support sorting by ppm_date ascending", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMSORT001")
		machine1.PpmDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("PPMSORT002")
		machine2.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("PPMSORT003")
		machine3.PpmDate = time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderPpmDateAsc}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 3)
		// Should be sorted by PPM date ascending (earliest first)
		suite.Assert().Equal("PPMSORT002", machinesList[0].SerialNumber) // Jan 10
		suite.Assert().Equal("PPMSORT003", machinesList[1].SerialNumber) // Feb 20
		suite.Assert().Equal("PPMSORT001", machinesList[2].SerialNumber) // Mar 15
	})

	suite.Run("should support sorting by ppm_date descending", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMSORT004")
		machine1.PpmDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("PPMSORT005")
		machine2.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("PPMSORT006")
		machine3.PpmDate = time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderPpmDateDesc}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 3)
		// Should be sorted by PPM date descending (latest first)
		suite.Assert().Equal("PPMSORT004", machinesList[0].SerialNumber) // Mar 15
		suite.Assert().Equal("PPMSORT006", machinesList[1].SerialNumber) // Feb 20
		suite.Assert().Equal("PPMSORT005", machinesList[2].SerialNumber) // Jan 10
	})

	suite.Run("should support sorting by tnc_date ascending", func() {
		// Create machines with different TNC dates
		machine1 := testutils.CreateMachine("TNCSORT001")
		machine1.TncDate = time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("TNCSORT002")
		machine2.TncDate = time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("TNCSORT003")
		machine3.TncDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderTncDateAsc}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 3)
		// Should be sorted by TNC date ascending (earliest first)
		suite.Assert().Equal("TNCSORT002", machinesList[0].SerialNumber) // Feb 5
		suite.Assert().Equal("TNCSORT003", machinesList[1].SerialNumber) // Mar 15
		suite.Assert().Equal("TNCSORT001", machinesList[2].SerialNumber) // Apr 25
	})

	suite.Run("should support sorting by tnc_date descending", func() {
		// Create machines with different TNC dates
		machine1 := testutils.CreateMachine("TNCSORT004")
		machine1.TncDate = time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("TNCSORT005")
		machine2.TncDate = time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("TNCSORT006")
		machine3.TncDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		options := &machines.ListOptions{Limit: 10, Offset: 0, Sort: machines.SortOrderTncDateDesc}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 3)
		// Should be sorted by TNC date descending (latest first)
		suite.Assert().Equal("TNCSORT004", machinesList[0].SerialNumber) // Apr 25
		suite.Assert().Equal("TNCSORT006", machinesList[1].SerialNumber) // Mar 15
		suite.Assert().Equal("TNCSORT005", machinesList[2].SerialNumber) // Feb 5
	})

	suite.Run("should filter by ppm_date_from", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMRANGE001")
		machine1.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		machine2 := testutils.CreateMachine("PPMRANGE002")
		machine2.PpmDate = time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		machine3 := testutils.CreateMachine("PPMRANGE003")
		machine3.PpmDate = time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		fromDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		options := &machines.ListOptions{
			Limit:       10,
			Offset:      0,
			Sort:        machines.SortOrderPpmDateAsc,
			PpmDateFrom: &fromDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 2) // Only Feb 15 and Mar 20
		suite.Assert().Equal("PPMRANGE002", machinesList[0].SerialNumber)
		suite.Assert().Equal("PPMRANGE003", machinesList[1].SerialNumber)
	})

	suite.Run("should filter by ppm_date_to", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMRANGE004")
		machine1.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		machine2 := testutils.CreateMachine("PPMRANGE005")
		machine2.PpmDate = time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		machine3 := testutils.CreateMachine("PPMRANGE006")
		machine3.PpmDate = time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		toDate := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
		options := &machines.ListOptions{
			Limit:     10,
			Offset:    0,
			Sort:      machines.SortOrderPpmDateAsc,
			PpmDateTo: &toDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 2) // Only Jan 10 and Feb 15
		suite.Assert().Equal("PPMRANGE004", machinesList[0].SerialNumber)
		suite.Assert().Equal("PPMRANGE005", machinesList[1].SerialNumber)
	})

	suite.Run("should filter by ppm_date range", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMRANGE007")
		machine1.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		machine2 := testutils.CreateMachine("PPMRANGE008")
		machine2.PpmDate = time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		machine3 := testutils.CreateMachine("PPMRANGE009")
		machine3.PpmDate = time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		fromDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
		options := &machines.ListOptions{
			Limit:       10,
			Offset:      0,
			Sort:        machines.SortOrderPpmDateAsc,
			PpmDateFrom: &fromDate,
			PpmDateTo:   &toDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Only Feb 15
		suite.Assert().Equal("PPMRANGE008", machinesList[0].SerialNumber)
	})

	suite.Run("should filter by single ppm_date when from equals to", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("PPMRANGE010")
		machine1.PpmDate = time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC)
		machine2 := testutils.CreateMachine("PPMRANGE011")
		machine2.PpmDate = time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		machine3 := testutils.CreateMachine("PPMRANGE012")
		machine3.PpmDate = time.Date(2025, 2, 16, 0, 0, 0, 0, time.UTC)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		exactDate := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		options := &machines.ListOptions{
			Limit:       10,
			Offset:      0,
			Sort:        machines.SortOrderPpmDateAsc,
			PpmDateFrom: &exactDate,
			PpmDateTo:   &exactDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Only exact match
		suite.Assert().Equal("PPMRANGE011", machinesList[0].SerialNumber)
	})

	suite.Run("should filter by tnc_date range", func() {
		// Create machines with different TNC dates
		machine1 := testutils.CreateMachine("TNCRANGE001")
		machine1.TncDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		machine2 := testutils.CreateMachine("TNCRANGE002")
		machine2.TncDate = time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
		machine3 := testutils.CreateMachine("TNCRANGE003")
		machine3.TncDate = time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		fromDate := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
		options := &machines.ListOptions{
			Limit:       10,
			Offset:      0,
			Sort:        machines.SortOrderTncDateAsc,
			TncDateFrom: &fromDate,
			TncDateTo:   &toDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Only Feb 15
		suite.Assert().Equal("TNCRANGE002", machinesList[0].SerialNumber)
	})

	suite.Run("should combine date range with ppm status filter", func() {
		// Create machines with different PPM dates
		machine1 := testutils.CreateMachine("COMBINED001")
		machine1.PpmDate = time.Now().UTC().AddDate(0, 0, -5) // 5 days ago (overdue)
		machine2 := testutils.CreateMachine("COMBINED002")
		machine2.PpmDate = time.Now().UTC().AddDate(0, 0, -10) // 10 days ago (overdue)
		machine3 := testutils.CreateMachine("COMBINED003")
		machine3.PpmDate = time.Now().UTC().AddDate(0, 0, 5) // 5 days from now (not overdue)

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		// Filter overdue machines from 7 days ago onwards
		fromDate := time.Now().UTC().AddDate(0, 0, -7)
		options := &machines.ListOptions{
			Limit:           10,
			Offset:          0,
			Sort:            machines.SortOrderPpmDateAsc,
			PpmStatusFilter: machines.PPMStatusOverdue,
			PpmDateFrom:     &fromDate,
		}
		machinesList, err := suite.repo.List(ctx, options)
		suite.Require().NoError(err)
		suite.Assert().Len(machinesList, 1) // Only machine1 (5 days ago is within 7 days)
		suite.Assert().Equal("COMBINED001", machinesList[0].SerialNumber)
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
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE001")
		dueMachine.PpmDate = time.Now().UTC() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE001")
		almostDueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now (not due)
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE002")
		dueMachine.PpmDate = time.Now().UTC() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE002")
		almostDueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE003")
		dueMachine.PpmDate = time.Now().UTC() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE003")
		almostDueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE003")
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now (not due)
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
			machine.PpmDate = time.Now().UTC().AddDate(0, 0, -i) // All overdue
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		options := &machines.ListOptions{
			Limit:           2,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		overdueMachine1.PpmDate = time.Now().UTC().AddDate(0, 0, -1)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine1))
		time.Sleep(10 * time.Millisecond)

		overdueMachine2 := testutils.CreateMachine("OVERDUE021")
		overdueMachine2.PpmDate = time.Now().UTC().AddDate(0, 0, -2)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine2))

		options := &machines.ListOptions{
			Limit:           50,
			Offset:          0,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday (overdue)
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		dueMachine := testutils.CreateMachine("DUE001")
		dueMachine.PpmDate = time.Now().UTC() // Today (due)
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		almostDueMachine := testutils.CreateMachine("ALMOSTDUE001")
		almostDueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 10) // 10 days from now (almost due)
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now (not due)
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
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now
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
			machine.PpmDate = time.Now().UTC().AddDate(0, 0, -i) // Different overdue dates
			suite.Require().NoError(suite.repo.Create(ctx, machine))
		}

		// Create multiple due machines
		for i := 1; i <= 2; i++ {
			machine := testutils.CreateMachine(fmt.Sprintf("DUE%03d", i))
			machine.PpmDate = time.Now().UTC() // All due today
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
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:   machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 2)
	})

	suite.Run("should filter by PPM status overdue", func() {
		// Create overdue machine
		overdueMachine := testutils.CreateMachine("OVERDUE001")
		overdueMachine.Brand = "HP"
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		// Create future machine (not overdue)
		futureMachine := testutils.CreateMachine("FUTURE001")
		futureMachine.Brand = "HP"
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		// Search for HP machines that are overdue
		results, err := suite.repo.Search(ctx, "HP", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		dueMachine.PpmDate = time.Now().UTC() // Today
		suite.Require().NoError(suite.repo.Create(ctx, dueMachine))

		// Create machine due tomorrow
		tomorrowMachine := testutils.CreateMachine("TOMORROW001")
		tomorrowMachine.Brand = "Canon"
		tomorrowMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 1) // Tomorrow
		suite.Require().NoError(suite.repo.Create(ctx, tomorrowMachine))

		// Search for Canon machines due today
		results, err := suite.repo.Search(ctx, "Canon", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
		almostDueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 10) // 10 days from now
		suite.Require().NoError(suite.repo.Create(ctx, almostDueMachine))

		// Create machine far in future (not almost due)
		farFutureMachine := testutils.CreateMachine("FARFUTURE001")
		farFutureMachine.Brand = "Brother"
		farFutureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, farFutureMachine))

		// Search for Brother machines almost due
		results, err := suite.repo.Search(ctx, "Brother", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderUpdatedAtDesc,
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
			Sort:  machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 1)
		suite.Assert().Equal("CASE001", results[0].SerialNumber)
	})

	suite.Run("should support sorting search results by ppm_date ascending", func() {
		// Create machines with same searchable content but different PPM dates
		machine1 := testutils.CreateMachine("SRCHPPM001")
		machine1.Brand = "Xerox"
		machine1.PpmDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("SRCHPPM002")
		machine2.Brand = "Xerox"
		machine2.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("SRCHPPM003")
		machine3.Brand = "Xerox"
		machine3.PpmDate = time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		results, err := suite.repo.Search(ctx, "Xerox", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderPpmDateAsc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		// Should be sorted by PPM date ascending (earliest first)
		suite.Assert().Equal("SRCHPPM002", results[0].SerialNumber) // Jan 10
		suite.Assert().Equal("SRCHPPM003", results[1].SerialNumber) // Feb 20
		suite.Assert().Equal("SRCHPPM001", results[2].SerialNumber) // Mar 15
	})

	suite.Run("should support sorting search results by ppm_date descending", func() {
		// Create machines with same searchable content but different PPM dates
		machine1 := testutils.CreateMachine("SRCHPPM004")
		machine1.Brand = "Ricoh"
		machine1.PpmDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("SRCHPPM005")
		machine2.Brand = "Ricoh"
		machine2.PpmDate = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("SRCHPPM006")
		machine3.Brand = "Ricoh"
		machine3.PpmDate = time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		results, err := suite.repo.Search(ctx, "Ricoh", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderPpmDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		// Should be sorted by PPM date descending (latest first)
		suite.Assert().Equal("SRCHPPM004", results[0].SerialNumber) // Mar 15
		suite.Assert().Equal("SRCHPPM006", results[1].SerialNumber) // Feb 20
		suite.Assert().Equal("SRCHPPM005", results[2].SerialNumber) // Jan 10
	})

	suite.Run("should support sorting search results by tnc_date ascending", func() {
		// Create machines with same searchable content but different TNC dates
		machine1 := testutils.CreateMachine("SRCHTNC001")
		machine1.Brand = "Konica"
		machine1.TncDate = time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("SRCHTNC002")
		machine2.Brand = "Konica"
		machine2.TncDate = time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("SRCHTNC003")
		machine3.Brand = "Konica"
		machine3.TncDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		results, err := suite.repo.Search(ctx, "Konica", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderTncDateAsc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		// Should be sorted by TNC date ascending (earliest first)
		suite.Assert().Equal("SRCHTNC002", results[0].SerialNumber) // Feb 5
		suite.Assert().Equal("SRCHTNC003", results[1].SerialNumber) // Mar 15
		suite.Assert().Equal("SRCHTNC001", results[2].SerialNumber) // Apr 25
	})

	suite.Run("should support sorting search results by tnc_date descending", func() {
		// Create machines with same searchable content but different TNC dates
		machine1 := testutils.CreateMachine("SRCHTNC004")
		machine1.Brand = "Epson"
		machine1.TncDate = time.Date(2025, 4, 25, 0, 0, 0, 0, time.UTC) // Latest
		machine2 := testutils.CreateMachine("SRCHTNC005")
		machine2.Brand = "Epson"
		machine2.TncDate = time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC) // Earliest
		machine3 := testutils.CreateMachine("SRCHTNC006")
		machine3.Brand = "Epson"
		machine3.TncDate = time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC) // Middle

		suite.Require().NoError(suite.repo.Create(ctx, machine1))
		suite.Require().NoError(suite.repo.Create(ctx, machine2))
		suite.Require().NoError(suite.repo.Create(ctx, machine3))

		results, err := suite.repo.Search(ctx, "Epson", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderTncDateDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)
		// Should be sorted by TNC date descending (latest first)
		suite.Assert().Equal("SRCHTNC004", results[0].SerialNumber) // Apr 25
		suite.Assert().Equal("SRCHTNC006", results[1].SerialNumber) // Mar 15
		suite.Assert().Equal("SRCHTNC005", results[2].SerialNumber) // Feb 5
	})

	suite.Run("should handle empty search query", func() {
		// Create a machine
		machine := testutils.CreateMachine("EMPTY001")
		suite.Require().NoError(suite.repo.Create(ctx, machine))

		// Search with empty query
		results, err := suite.repo.Search(ctx, "", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:  machines.SortOrderUpdatedAtDesc,
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
			Sort:   machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Len(results, 3)

		// Verify all results are Brother machines
		for _, result := range results {
			suite.Assert().Equal("Brother", result.Brand)
		}
	})

	suite.Run("should count search results correctly", func() {
		// Create test machines
		hpMachine := testutils.CreateMachine("HP001")
		hpMachine.Brand = "HP"
		suite.Require().NoError(suite.repo.Create(ctx, hpMachine))

		canonMachine := testutils.CreateMachine("CANON001")
		canonMachine.Brand = "Canon"
		suite.Require().NoError(suite.repo.Create(ctx, canonMachine))

		// Count search results for "HP"
		count, err := suite.repo.CountSearch(ctx, "HP", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, count)

		// Count search results for "Canon"
		count, err = suite.repo.CountSearch(ctx, "Canon", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, count)

		// Count search results for non-existent term
		count, err = suite.repo.CountSearch(ctx, "nonexistent", &machines.ListOptions{
			Limit: 10,
			Sort:  machines.SortOrderUpdatedAtDesc,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(0, count)
	})

	suite.Run("should count search results with PPM status filter", func() {
		// Create overdue HP machine
		overdueMachine := testutils.CreateMachine("HP001")
		overdueMachine.Brand = "HP"
		overdueMachine.PpmDate = time.Now().UTC().AddDate(0, 0, -1) // Yesterday
		suite.Require().NoError(suite.repo.Create(ctx, overdueMachine))

		// Create future HP machine (not overdue)
		futureMachine := testutils.CreateMachine("HP002")
		futureMachine.Brand = "HP"
		futureMachine.PpmDate = time.Now().UTC().AddDate(0, 0, 30) // 30 days from now
		suite.Require().NoError(suite.repo.Create(ctx, futureMachine))

		// Count HP machines that are overdue
		count, err := suite.repo.CountSearch(ctx, "HP", &machines.ListOptions{
			Limit:           10,
			Sort:            machines.SortOrderUpdatedAtDesc,
			PpmStatusFilter: machines.PPMStatusOverdue,
		})
		suite.Require().NoError(err)
		suite.Assert().Equal(1, count)
	})
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}
