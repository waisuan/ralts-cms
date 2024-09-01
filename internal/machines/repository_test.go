package machines_test

// Basic imports
import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/testutils/factory"
	"ralts-cms/internal/testutils/pg"
	pkgpg "ralts-cms/pkg/pg"
	"testing"
	"time"
)

type RepositoryTestSuite struct {
	suite.Suite
	d    *deps.Dependencies
	repo machines.Repository
}

func (suite *RepositoryTestSuite) SetupTest() {
	d := deps.Initialise()
	suite.d = d
	suite.repo = machines.NewRepository(d.DB)
}

func (suite *RepositoryTestSuite) TearDownTest() {
	var (
		t = suite.T()
	)

	err := pg.TruncateTables(suite.d.DB)
	require.NoError(t, err)
}

func (suite *RepositoryTestSuite) TestQuery() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(100, 0, "", false)
	require.NoError(t, err)
	assert.Len(t, res, numOfMachines)
	// By default, the queried result is returned in descending updated_at order.
	assert.Equal(t, testMachines[2].SerialNumber, res[0].SerialNumber)
	assert.Equal(t, testMachines[1].SerialNumber, res[1].SerialNumber)
	assert.Equal(t, testMachines[0].SerialNumber, res[2].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithLimit() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(1, 0, "", false)
	require.NoError(t, err)
	assert.Len(t, res, 1)
	// By default, the queried result is returned in descending updated_at order.
	assert.Equal(t, testMachines[2].SerialNumber, res[0].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithOffset() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(1, 2, "", false)
	require.NoError(t, err)
	assert.Len(t, res, 1)
	// By default, the queried result is returned in descending updated_at order.
	assert.Equal(t, testMachines[0].SerialNumber, res[0].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithReversedSortOrder() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(100, 0, "", true)
	require.NoError(t, err)
	assert.Len(t, res, numOfMachines)
	assert.Equal(t, testMachines[0].SerialNumber, res[0].SerialNumber)
	assert.Equal(t, testMachines[1].SerialNumber, res[1].SerialNumber)
	assert.Equal(t, testMachines[2].SerialNumber, res[2].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithSortField() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		d := time.Date(2024, time.November, 5-i, 0, 0, 0, 0, time.Local)
		machine.PpmDate = &d
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(100, 0, "ppm_date", false)
	require.NoError(t, err)
	assert.Len(t, res, numOfMachines)
	assert.Equal(t, testMachines[0].SerialNumber, res[0].SerialNumber)
	assert.Equal(t, testMachines[1].SerialNumber, res[1].SerialNumber)
	assert.Equal(t, testMachines[2].SerialNumber, res[2].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithSortFieldAndReversedOrder() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		d := time.Date(2024, time.November, 5-i, 0, 0, 0, 0, time.Local)
		machine.PpmDate = &d
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(100, 0, "ppm_date", true)
	require.NoError(t, err)
	assert.Len(t, res, numOfMachines)
	assert.Equal(t, testMachines[2].SerialNumber, res[0].SerialNumber)
	assert.Equal(t, testMachines[1].SerialNumber, res[1].SerialNumber)
	assert.Equal(t, testMachines[0].SerialNumber, res[2].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryWithInvalidSortField() {
	var (
		numOfMachines = 3
		t             = suite.T()
	)

	testMachines := make([]*machines.Machine, 0)
	for i := 0; i < numOfMachines; i++ {
		machine := factory.BuildMachine()
		_, err := suite.repo.Create(machine)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testMachines = append(testMachines, machine)
	}

	res, err := suite.repo.Query(100, 0, "dummy_field", false)
	require.ErrorContains(t, err, "does not exist")
	assert.Empty(t, res, numOfMachines)
}

func (suite *RepositoryTestSuite) TestQueryNoMachines() {
	var (
		limit  = 100
		offset = 0
		t      = suite.T()
	)

	res, err := suite.repo.Query(limit, offset, "", false)
	require.NoError(t, err)
	assert.Empty(t, res)
}

func (suite *RepositoryTestSuite) TestGet() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	_, err := suite.repo.Create(machine)
	require.NoError(t, err)

	res, err := suite.repo.GetBySerialNumber(machine.SerialNumber)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)
}

func (suite *RepositoryTestSuite) TestGetNotFound() {
	var (
		t = suite.T()
	)

	res, err := suite.repo.GetBySerialNumber("test")
	require.ErrorIs(t, err, pkgpg.ErrNotFound)
	assert.Nil(t, res)
}

func (suite *RepositoryTestSuite) TestCreate() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	res, err := suite.repo.Create(machine)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)
	assert.Equal(t, machine.Customer, res.Customer)
	assert.Equal(t, machine.State, res.State)
	assert.Equal(t, machine.AccountType, res.AccountType)
	assert.Equal(t, machine.Model, res.Model)
	assert.Equal(t, machine.Status, res.Status)
	assert.Equal(t, machine.Brand, res.Brand)
	assert.Equal(t, machine.District, res.District)
	assert.Equal(t, machine.PersonInCharge, res.PersonInCharge)
	assert.Equal(t, machine.ReportedBy, res.ReportedBy)
	assert.Equal(t, machine.AdditionalNotes, res.AdditionalNotes)
	assert.Equal(t, machine.Attachment, res.Attachment)
	assert.Equal(t, machine.PpmStatus, res.PpmStatus)
	assert.Equal(t, machine.TncDate, res.TncDate)
	assert.Equal(t, machine.PpmDate, res.PpmDate)
	assert.Equal(t, machine.ID, res.ID)
	assert.WithinDuration(t, time.Now(), res.CreatedAt, 10*time.Second)
	assert.WithinDuration(t, time.Now(), res.UpdatedAt, 10*time.Second)

	res, err = suite.repo.GetBySerialNumber(machine.SerialNumber)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)
	assert.Equal(t, machine.Customer, res.Customer)
	assert.Equal(t, machine.State, res.State)
	assert.Equal(t, machine.AccountType, res.AccountType)
	assert.Equal(t, machine.Model, res.Model)
	assert.Equal(t, machine.Status, res.Status)
	assert.Equal(t, machine.Brand, res.Brand)
	assert.Equal(t, machine.District, res.District)
	assert.Equal(t, machine.PersonInCharge, res.PersonInCharge)
	assert.Equal(t, machine.ReportedBy, res.ReportedBy)
	assert.Equal(t, machine.AdditionalNotes, res.AdditionalNotes)
	assert.Equal(t, machine.Attachment, res.Attachment)
	assert.Equal(t, machine.PpmStatus, res.PpmStatus)
	assert.Equal(t, machine.FormattedTncDate(), res.FormattedTncDate())
	assert.Equal(t, machine.FormattedPpmDate(), res.FormattedPpmDate())
	assert.Equal(t, machine.ID, res.ID)
	assert.WithinDuration(t, time.Now(), res.CreatedAt, 10*time.Second)
	assert.WithinDuration(t, time.Now(), res.UpdatedAt, 10*time.Second)
}

func (suite *RepositoryTestSuite) TestCreateEmptySerialNum() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	machine.SerialNumber = ""

	res, err := suite.repo.Create(machine)
	require.ErrorContains(t, err, "violates not-null constraint")
	assert.Nil(t, res)
}

func (suite *RepositoryTestSuite) TestCreateDupSerialNum() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	_, err := suite.repo.Create(machine)
	require.NoError(t, err)

	_, err = suite.repo.Create(machine)
	require.ErrorContains(t, err, "duplicate key")
}

func (suite *RepositoryTestSuite) TestUpdate() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	res, err := suite.repo.Create(machine)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)

	targetID := res.ID

	machine.Status = "WIP"
	machine.PpmDate = nil
	res, err = suite.repo.Update(machine)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, targetID, res.ID)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)
	assert.Equal(t, machine.Status, res.Status)
	assert.Nil(t, res.PpmDate)
	assert.Greater(t, res.UpdatedAt, machine.UpdatedAt)

	res, err = suite.repo.GetBySerialNumber(machine.SerialNumber)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, targetID, res.ID)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)
	assert.Equal(t, machine.Status, res.Status)
	assert.Nil(t, res.PpmDate)
	assert.Greater(t, res.UpdatedAt, machine.UpdatedAt)
}

func (suite *RepositoryTestSuite) TestUpdateNotFound() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	res, err := suite.repo.Create(machine)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)

	err = suite.repo.DeleteBySerialNumber(machine.SerialNumber)
	require.NoError(t, err)

	res, err = suite.repo.GetBySerialNumber(machine.SerialNumber)
	require.ErrorIs(t, err, pkgpg.ErrNotFound)
	assert.Nil(t, res)
}

func (suite *RepositoryTestSuite) TestDeleteBySerialNumber() {
	var (
		machine = factory.BuildMachine()
		t       = suite.T()
	)

	res, err := suite.repo.Create(machine)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, machine.SerialNumber, res.SerialNumber)

	err = suite.repo.DeleteBySerialNumber(machine.SerialNumber)
	require.NoError(t, err)

	res, err = suite.repo.GetBySerialNumber(machine.SerialNumber)
	require.ErrorIs(t, err, pkgpg.ErrNotFound)
	assert.Nil(t, res)
}

func (suite *RepositoryTestSuite) TestDeleteBySerialNumberNotFound() {
	var (
		t = suite.T()
	)

	err := suite.repo.DeleteBySerialNumber("testing")
	require.ErrorIs(t, err, pkgpg.ErrNotFound)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
