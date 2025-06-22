package deps

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBConfig holds DynamoDB-specific configuration
type DynamoDBConfig struct {
	Endpoint string
	Region   string
	Table    string
	// Production settings
	UseLocalEndpoint bool
	// AWS credentials for local development
	AccessKeyID     string
	SecretAccessKey string
}

// NewDynamoDBConfig creates a new DynamoDB configuration from the main config
func NewDynamoDBConfig(cfg *Config) *DynamoDBConfig {
	return &DynamoDBConfig{
		Endpoint:         cfg.DynamoDBEndpoint,
		Region:           cfg.DynamoDBRegion,
		Table:            cfg.DynamoDBTable,
		UseLocalEndpoint: cfg.Env == "development" || cfg.Env == "local",
		AccessKeyID:      cfg.AwsAccessKeyID,
		SecretAccessKey:  cfg.AwsSecretAccessKey,
	}
}

// NewDynamoDBClient creates a new DynamoDB client based on the configuration
func NewDynamoDBClient(ctx context.Context, dbConfig *DynamoDBConfig) (*dynamodb.Client, error) {
	var awsConfig aws.Config
	var err error

	if dbConfig.UseLocalEndpoint {
		// Local development configuration
		log.Printf("Initializing DynamoDB client for local development")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(dbConfig.Region),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           dbConfig.Endpoint,
						SigningRegion: dbConfig.Region,
					}, nil
				},
			)),
		)
	} else {
		// Production configuration - uses AWS credentials from environment or IAM roles
		log.Printf("Initializing DynamoDB client for production environment")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(dbConfig.Region),
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
