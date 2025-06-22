package maintenance

import (
	"context"
	"testing"
	"time"

	"ralts-cms/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MaintenanceRepositoryTestSuite defines the test suite for maintenance repository
type MaintenanceRepositoryTestSuite struct {
	suite.Suite
	*testutils.DynamoDBTestSuite
	repo Repository
}

func (suite *MaintenanceRepositoryTestSuite) SetupSuite() {
	testutils.SkipIfDynamoDBUnavailable(suite.T())
	suite.DynamoDBTestSuite = testutils.NewDynamoDBTestSuite("test-ralts")
	err := suite.DynamoDBTestSuite.SetupSuite()
	require.NoError(suite.T(), err)
	suite.repo = NewRepository(suite.Client, suite.TableName)
}

func (suite *MaintenanceRepositoryTestSuite) TearDownSuite() {
	if suite.DynamoDBTestSuite != nil {
		err := suite.DynamoDBTestSuite.TearDownSuite()
		if err != nil {
			suite.T().Logf("Failed to teardown DynamoDB test suite: %v", err)
		}
	}
}

func (suite *MaintenanceRepositoryTestSuite) SetupTest() {
	if suite.DynamoDBTestSuite != nil {
		err := suite.DynamoDBTestSuite.SetupTest()
		require.NoError(suite.T(), err)
	}
}

func TestMaintenanceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(MaintenanceRepositoryTestSuite))
}

func (suite *MaintenanceRepositoryTestSuite) TestCreate() {
	ctx := context.Background()

	suite.Run("should create maintenance successfully", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			WorkOrderDate:       "2025-05-12",
			ActionTaken:         "Routine maintenance",
			ReportedBy:          "John Doe",
			WorkerOrderType:     "Preventive",
			Attachment:          "maintenance.pdf",
		}

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(maintenance.CreatedAt)
		suite.Assert().NotEmpty(maintenance.UpdatedAt)
	})

	suite.Run("should fail when creating duplicate maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE123",
			WorkOrderNumber:     "WO001",
			ActionTaken:         "Another maintenance",
		}

		err := suite.repo.Create(ctx, maintenance)
		suite.Require().Error(err)
		suite.Assert().Contains(err.Error(), "ConditionalCheckFailedException")
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestGetByWorkOrder() {
	ctx := context.Background()

	suite.Run("should get existing maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE456",
			WorkOrderNumber:     "WO002",
			ActionTaken:         "Get test maintenance",
			ReportedBy:          "Jane Smith",
		}
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE456", "WO002")
		suite.Require().NoError(err)
		suite.Assert().Equal("MACHINE456", retrieved.MachineSerialNumber)
		suite.Assert().Equal("WO002", retrieved.WorkOrderNumber)
		suite.Assert().Equal("Get test maintenance", retrieved.ActionTaken)
		suite.Assert().Equal("Jane Smith", retrieved.ReportedBy)
		suite.Assert().NotEmpty(retrieved.CreatedAt)
		suite.Assert().NotEmpty(retrieved.UpdatedAt)
	})

	suite.Run("should return error for non-existent maintenance", func() {
		_, err := suite.repo.GetByWorkOrder(ctx, "NONEXISTENT", "WO999")
		suite.Require().Error(err)
		suite.Assert().Equal("maintenance not found", err.Error())
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestListByMachine() {
	ctx := context.Background()

	suite.Run("should list all maintenance for a machine", func() {
		machineSerial := "MACHINE789"

		maintenance1 := &Maintenance{
			MachineSerialNumber: machineSerial,
			WorkOrderNumber:     "WO003",
			ActionTaken:         "First maintenance",
			ReportedBy:          "Tech1",
		}
		err := suite.repo.Create(ctx, maintenance1)
		suite.Require().NoError(err)

		maintenance2 := &Maintenance{
			MachineSerialNumber: machineSerial,
			WorkOrderNumber:     "WO004",
			ActionTaken:         "Second maintenance",
			ReportedBy:          "Tech2",
		}
		err = suite.repo.Create(ctx, maintenance2)
		suite.Require().NoError(err)

		maintenance3 := &Maintenance{
			MachineSerialNumber: "OTHERMACHINE",
			WorkOrderNumber:     "WO005",
			ActionTaken:         "Other machine maintenance",
			ReportedBy:          "Tech3",
		}
		err = suite.repo.Create(ctx, maintenance3)
		suite.Require().NoError(err)

		maintenanceList, err := suite.repo.ListByMachine(ctx, machineSerial)
		suite.Require().NoError(err)
		suite.Assert().Len(maintenanceList, 2)
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	suite.Run("should update existing maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINE999",
			WorkOrderNumber:     "WO010",
			ActionTaken:         "Original action",
			ReportedBy:          "TechX",
		}
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		originalCreatedAt := maintenance.CreatedAt
		originalUpdatedAt := maintenance.UpdatedAt

		time.Sleep(1 * time.Millisecond)

		maintenance.ActionTaken = "Updated action"
		maintenance.ReportedBy = "TechY"
		err = suite.repo.Update(ctx, maintenance)
		suite.Require().NoError(err)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "MACHINE999", "WO010")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated action", retrieved.ActionTaken)
		suite.Assert().Equal("TechY", retrieved.ReportedBy)
		suite.Assert().Equal(originalCreatedAt, retrieved.CreatedAt)
		suite.Assert().NotEqual(originalUpdatedAt, retrieved.UpdatedAt)
	})

	suite.Run("should return error for non-existent maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "NONEXISTENT",
			WorkOrderNumber:     "WONONE",
			ActionTaken:         "No action",
		}
		err := suite.repo.Update(ctx, maintenance)
		suite.Require().Error(err)
		suite.Assert().Equal("maintenance not found", err.Error())
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	suite.Run("should delete existing maintenance", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "MACHINEDEL",
			WorkOrderNumber:     "WODEL",
			ActionTaken:         "Delete action",
			ReportedBy:          "TechDel",
		}
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)

		_, err = suite.repo.GetByWorkOrder(ctx, "MACHINEDEL", "WODEL")
		suite.Require().NoError(err)

		err = suite.repo.Delete(ctx, "MACHINEDEL", "WODEL")
		suite.Require().NoError(err)

		_, err = suite.repo.GetByWorkOrder(ctx, "MACHINEDEL", "WODEL")
		suite.Require().Error(err)
		suite.Assert().Equal("maintenance not found", err.Error())
	})

	suite.Run("should return error for non-existent maintenance", func() {
		err := suite.repo.Delete(ctx, "NONEXISTENT", "WONONE")
		suite.Require().Error(err)
		suite.Assert().Equal("maintenance not found", err.Error())
	})
}

func (suite *MaintenanceRepositoryTestSuite) TestIntegration() {
	ctx := context.Background()

	suite.Run("should handle full CRUD workflow", func() {
		maintenance := &Maintenance{
			MachineSerialNumber: "INTEGRATIONM",
			WorkOrderNumber:     "INTEGRATIONW",
			ActionTaken:         "Integration action",
			ReportedBy:          "TechInt",
		}
		err := suite.repo.Create(ctx, maintenance)
		suite.Require().NoError(err)
		suite.Assert().NotEmpty(maintenance.CreatedAt)
		suite.Assert().NotEmpty(maintenance.UpdatedAt)

		retrieved, err := suite.repo.GetByWorkOrder(ctx, "INTEGRATIONM", "INTEGRATIONW")
		suite.Require().NoError(err)
		suite.Assert().Equal("Integration action", retrieved.ActionTaken)
		suite.Assert().Equal("TechInt", retrieved.ReportedBy)

		retrieved.ActionTaken = "Updated Integration Action"
		retrieved.ReportedBy = "TechInt2"
		err = suite.repo.Update(ctx, retrieved)
		suite.Require().NoError(err)

		updated, err := suite.repo.GetByWorkOrder(ctx, "INTEGRATIONM", "INTEGRATIONW")
		suite.Require().NoError(err)
		suite.Assert().Equal("Updated Integration Action", updated.ActionTaken)
		suite.Assert().Equal("TechInt2", updated.ReportedBy)

		err = suite.repo.Delete(ctx, "INTEGRATIONM", "INTEGRATIONW")
		suite.Require().NoError(err)

		_, err = suite.repo.GetByWorkOrder(ctx, "INTEGRATIONM", "INTEGRATIONW")
		suite.Require().Error(err)
		suite.Assert().Equal("maintenance not found", err.Error())
	})
}
