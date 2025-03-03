package posts_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/posts"
	"ralts-cms/internal/testutils/factory"
	"ralts-cms/internal/testutils/pg"
	"testing"
	"time"
)

type RepositoryTestSuite struct {
	suite.Suite
	d    *deps.Dependencies
	repo posts.Repository
}

func (suite *RepositoryTestSuite) SetupTest() {
	d := deps.Initialise()
	suite.d = d
	suite.repo = posts.NewRepository(d.DB)
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
		numOfPosts = 3
		t          = suite.T()
	)

	testPosts := make([]*posts.Post, 0)
	for i := 0; i < numOfPosts; i++ {
		post := factory.BuildPost()
		_, err := suite.repo.Create(context.Background(), post)
		require.NoError(t, err)

		// Ensure that there's a gap in the created_at/updated_at timestamps between data creation.
		time.Sleep(1 * time.Second)

		testPosts = append(testPosts, post)
	}

	//res, err := suite.repo.Query(context.Background(), 100, 0, "", false)
	//require.NoError(t, err)
	//assert.Len(t, res, numOfMachines)
	//// By default, the queried result is returned in descending updated_at order.
	//assert.Equal(t, testMachines[2].SerialNumber, res[0].SerialNumber)
	//assert.Equal(t, testMachines[1].SerialNumber, res[1].SerialNumber)
	//assert.Equal(t, testMachines[0].SerialNumber, res[2].SerialNumber)
}

func (suite *RepositoryTestSuite) TestQueryNoPosts() {
	var (
		t = suite.T()
	)

	posts, err := suite.repo.Query(context.Background(), 100, 0, "", false)
	require.NoError(t, err)
	assert.Len(t, posts, 0)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
