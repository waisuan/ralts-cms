package deps

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewDynamoDBClient creates a new DynamoDB client based on the configuration
func NewDynamoDBClient(ctx context.Context, cfg *Config) (*dynamodb.Client, error) {
	var awsConfig aws.Config
	var err error

	useLocalEndpoint := cfg.Env == "development" || cfg.Env == "local"

	if useLocalEndpoint {
		// Local development configuration
		log.Printf("Initializing DynamoDB client for local development")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.DynamoDBRegion),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           cfg.DynamoDBEndpoint,
						SigningRegion: cfg.DynamoDBRegion,
					}, nil
				},
			)),
		)
	} else {
		// Production configuration - uses AWS credentials from environment or IAM roles
		log.Printf("Initializing DynamoDB client for production environment")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.DynamoDBRegion),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsConfig)
	return client, nil
}

// ValidateDynamoDBConnection tests the connection to DynamoDB
func ValidateDynamoDBConnection(ctx context.Context, client *dynamodb.Client, tableName string) error {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to DynamoDB table %s: %w", tableName, err)
	}

	log.Printf("Successfully connected to DynamoDB table: %s", tableName)
	return nil
}
