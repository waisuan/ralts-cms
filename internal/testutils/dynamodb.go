package testutils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBTestSuite provides common setup and teardown for DynamoDB integration tests
type DynamoDBTestSuite struct {
	Client    *dynamodb.Client
	TableName string
}

// NewDynamoDBTestSuite creates a new test suite with DynamoDB client and table setup
func NewDynamoDBTestSuite(tableName string) *DynamoDBTestSuite {
	return &DynamoDBTestSuite{
		TableName: tableName,
	}
}

// SetupSuite initializes the DynamoDB client and creates the test table
func (suite *DynamoDBTestSuite) SetupSuite() error {
	// Setup test DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			},
		)),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return err
	}

	suite.Client = dynamodb.NewFromConfig(cfg)

	// Create test table
	return suite.createTestTable(context.Background())
}

// TearDownSuite cleans up the test table
func (suite *DynamoDBTestSuite) TearDownSuite() error {
	return suite.deleteTestTable(context.Background())
}

// createTestTable creates the DynamoDB table with the required schema
func (suite *DynamoDBTestSuite) createTestTable(ctx context.Context) error {
	_, err := suite.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(suite.TableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1_PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1_SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		BillingMode: types.BillingModePayPerRequest,
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("gsi_1"),
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1_PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1_SK"), KeyType: types.KeyTypeRange},
				},
			},
		},
	})
	return err
}

// deleteTestTable removes the test table
func (suite *DynamoDBTestSuite) deleteTestTable(ctx context.Context) error {
	_, err := suite.Client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(suite.TableName),
	})
	return err
}

// SetupTest cleans up the table before each test to ensure isolation
func (suite *DynamoDBTestSuite) SetupTest() error {
	// Clear all items from the table
	return suite.clearTable(context.Background())
}

// clearTable removes all items from the test table
func (suite *DynamoDBTestSuite) clearTable(ctx context.Context) error {
	// Scan the table to get all items
	scanOutput, err := suite.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(suite.TableName),
	})
	if err != nil {
		return err
	}

	// Delete all items
	for _, item := range scanOutput.Items {
		pkAttr, ok := item["PK"].(*types.AttributeValueMemberS)
		if !ok {
			continue
		}
		skAttr, ok := item["SK"].(*types.AttributeValueMemberS)
		if !ok {
			continue
		}

		_, err := suite.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(suite.TableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: pkAttr.Value},
				"SK": &types.AttributeValueMemberS{Value: skAttr.Value},
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// RunWithDynamoDB runs tests with DynamoDB setup and teardown
func RunWithDynamoDB(t *testing.T, tableName string, testFunc func(*DynamoDBTestSuite)) {
	suite := NewDynamoDBTestSuite(tableName)

	// Setup
	err := suite.SetupSuite()
	if err != nil {
		t.Fatalf("Failed to setup DynamoDB test suite: %v", err)
	}
	defer func() {
		if err := suite.TearDownSuite(); err != nil {
			t.Logf("Failed to teardown DynamoDB test suite: %v", err)
		}
	}()

	// Run the test
	testFunc(suite)
}

// CheckDynamoDBConnection checks if DynamoDB is available for testing
func CheckDynamoDBConnection() bool {
	// Try to connect to DynamoDB
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			},
		)),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return false
	}

	client := dynamodb.NewFromConfig(cfg)
	_, err = client.ListTables(context.Background(), &dynamodb.ListTablesInput{})
	return err == nil
}

// SkipIfDynamoDBUnavailable skips the test if DynamoDB is not available
func SkipIfDynamoDBUnavailable(t *testing.T) {
	if !CheckDynamoDBConnection() {
		t.Skip("DynamoDB not available, skipping integration test")
	}
}
