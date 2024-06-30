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

func (suite *RepositoryTestSuite) TestQueryNoMachines() {
	var (
		limit  = 100
		offset = 0
		t      = suite.T()
	)

	res, err := suite.repo.Query(limit, offset)
	require.NoError(t, err)
	assert.Empty(t, res)
}

func (suite *RepositoryTestSuite) TestCreate() {
	var (
		machine     = factory.BuildMachine()
		testMachine = factory.BuildMachine()
		t           = suite.T()
	)

	testMachine.SerialNumber = machine.SerialNumber

	res, err := suite.repo.Create(machine)
	require.NoError(t, err)
	assert.Equal(t, testMachine.SerialNumber, res.SerialNumber)
	assert.Equal(t, testMachine.Customer, res.Customer)
	assert.NotEqual(t, testMachine.ID, res.ID)
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
