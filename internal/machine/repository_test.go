package machine

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MachineRepositoryTestSuite defines the test suite for machine repository
type MachineRepositoryTestSuite struct {
	suite.Suite
	*testutils.DynamoDBTestSuite
	repo Repository
}

// SetupSuite sets up the test suite
func (suite *MachineRepositoryTestSuite) SetupSuite() {
	// Skip if DynamoDB is not available
	testutils.SkipIfDynamoDBUnavailable(suite.T())

	// Initialize DynamoDB test suite
	suite.DynamoDBTestSuite = testutils.NewDynamoDBTestSuite("test-ralts")
	err := suite.DynamoDBTestSuite.SetupSuite()
	require.NoError(suite.T(), err)

	// Initialize repository
	suite.repo = NewRepository(suite.Client, suite.TableName)
}

// TearDownSuite cleans up the test suite
func (suite *MachineRepositoryTestSuite) TearDownSuite() {
	if suite.DynamoDBTestSuite != nil {
		err := suite.DynamoDBTestSuite.TearDownSuite()
		if err != nil {
			suite.T().Logf("Failed to teardown DynamoDB test suite: %v", err)
		}
	}
}

// SetupTest sets up each test
func (suite *MachineRepositoryTestSuite) SetupTest() {
	if suite.DynamoDBTestSuite != nil {
		err := suite.DynamoDBTestSuite.SetupTest()
		require.NoError(suite.T(), err)
	}
}

// TestMachineRepositoryTestSuite runs the test suite
func TestMachineRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineRepositoryTestSuite))
}

// TestCreate tests the Create method
func (suite *MachineRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create machine successfully", func() {
		machine := &Machine{
			SerialNumber:    "TEST123",
			Customer:        "Test Customer",
			State:           "Active",
			AccountType:     "Premium",
			Model:           "Test Model",
			Status:          "Operational",
			Brand:           "Test Brand",
			District:        "Test District",
			PersonInCharge:  "John Doe",
			ReportedBy:      "Jane Smith",
			AdditionalNotes: "Test notes",
			Attachment:      "test.pdf",
			PpmStatus:       "Scheduled",
			TncDate:         "2025-05-12",
			PpmDate:         "2025-06-12",
		}

		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Verify timestamps were set
		suite.Assert().NotEmpty(machine.CreatedAt)
		suite.Assert().NotEmpty(machine.UpdatedAt)
	})

	suite.Run("should fail when creating duplicate machine", func() {
		machine := &Machine{
			SerialNumber: "TEST123",
			Customer:     "Test Customer",
		}

		err := suite.repo.Create(ctx, machine)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "ConditionalCheckFailedException")
	})
}

// TestGetBySerialNumber tests the GetBySerialNumber method
func (suite *MachineRepositoryTestSuite) TestGetBySerialNumber() {
	ctx := context.Background()

	suite.Run("should get existing machine", func() {
		// Create a test machine first
		machine := &Machine{
			SerialNumber: "GET123",
			Customer:     "Get Test Customer",
			State:        "Active",
		}
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Retrieve the machine
		retrieved, err := suite.repo.GetBySerialNumber(ctx, "GET123")
		suite.Require().NoError(err)
		suite.Assert().Equal("GET123", retrieved.SerialNumber)
		suite.Assert().Equal("Get Test Customer", retrieved.Customer)
		suite.Assert().Equal("Active", retrieved.State)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should return error for non-existent machine", func() {
		_, err := suite.repo.GetBySerialNumber(ctx, "NONEXISTENT")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})
}

// TestUpdate tests the Update method
func (suite *MachineRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	suite.Run("should update existing machine", func() {
		// Create a test machine first
		machine := &Machine{
			SerialNumber: "UPDATE123",
			Customer:     "Original Customer",
			State:        "Inactive",
		}
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		originalCreatedAt := machine.CreatedAt
		originalUpdatedAt := machine.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		// Update the machine
		machine.Customer = "Updated Customer"
		machine.State = "Active"
		err = suite.repo.Update(ctx, machine)
		suite.Require().NoError(err)

		// Verify the update
		retrieved, err := suite.repo.GetBySerialNumber(ctx, "UPDATE123")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated Customer", retrieved.Customer)
		suite.Assert().Equal("Active", retrieved.State)
		suite.Assert().Equal(originalCreatedAt, retrieved.CreatedAt)    // CreatedAt should not change
		suite.Assert().NotEqual(originalUpdatedAt, retrieved.UpdatedAt) // UpdatedAt should change
	})

	suite.Run("should return error for non-existent machine", func() {
		machine := &Machine{
			SerialNumber: "NONEXISTENT",
			Customer:     "Non-existent Customer",
		}

		err := suite.repo.Update(ctx, machine)
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})
}

// TestDelete tests the Delete method
func (suite *MachineRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing machine", func() {
		// Create a test machine first
		machine := &Machine{
			SerialNumber: "DELETE123",
			Customer:     "Delete Test Customer",
			State:        "Active",
		}
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)

		// Verify machine exists
		_, err = suite.repo.GetBySerialNumber(ctx, "DELETE123")
		suite.Require().NoError(err)

		// Delete the machine
		err = suite.repo.Delete(ctx, "DELETE123")
		suite.Require().NoError(err)

		// Verify machine is deleted
		_, err = suite.repo.GetBySerialNumber(ctx, "DELETE123")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})

	suite.Run("should return error for non-existent machine", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})
}

// TestIntegration tests the full CRUD workflow
func (suite *MachineRepositoryTestSuite) TestIntegration() {
	ctx := context.Background()

	suite.Run("should handle full CRUD workflow", func() {
		// Create
		machine := &Machine{
			SerialNumber: "INTEGRATION123",
			Customer:     "Integration Test Customer",
			State:        "New",
		}
		err := suite.repo.Create(ctx, machine)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(machine.CreatedAt)
		suite.Assert().NotEmpty(machine.UpdatedAt)

		// Read
		retrieved, err := suite.repo.GetBySerialNumber(ctx, "INTEGRATION123")
		suite.Require().NoError(err)
		suite.Assert().Equal("Integration Test Customer", retrieved.Customer)
		suite.Assert().Equal("New", retrieved.State)

		// Update
		retrieved.State = "Active"
		retrieved.Customer = "Updated Integration Customer"
		err = suite.repo.Update(ctx, retrieved)
		suite.Require().NoError(err)

		// Read again to verify update
		updated, err := suite.repo.GetBySerialNumber(ctx, "INTEGRATION123")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated Integration Customer", updated.Customer)
		suite.Assert().Equal("Active", updated.State)

		// Delete
		err = suite.repo.Delete(ctx, "INTEGRATION123")
		suite.Require().NoError(err)

		// Verify deletion
		_, err = suite.repo.GetBySerialNumber(ctx, "INTEGRATION123")
		suite.Require().Error(err)
		suite.Assert().Equal("machine not found", err.Error())
	})
}
