package deps

import (
	"context"
	"log"

	"ralts-cms/internal/machine"
	"ralts-cms/internal/maintenance"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Dependencies struct {
	Config *Config

	// DynamoDB
	DynamoDBClient *dynamodb.Client
	DynamoDBConfig *DynamoDBConfig

	// Repositories
	MachineRepository     machine.Repository
	MaintenanceRepository maintenance.Repository
}

func Initialise() *Dependencies {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %e", err)
	}

	// Initialize DynamoDB configuration
	dbConfig := NewDynamoDBConfig(cfg)

	// Initialize DynamoDB client
	dynamoClient, err := NewDynamoDBClient(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("failed to initialize DynamoDB client: %e", err)
	}

	// Validate DynamoDB connection
	if err := ValidateDynamoDBConnection(context.Background(), dynamoClient, dbConfig.Table); err != nil {
		log.Fatalf("failed to validate DynamoDB connection: %e", err)
	}

	// Initialize repositories
	machineRepo := machine.NewRepository(dynamoClient, dbConfig.Table)
	maintenanceRepo := maintenance.NewRepository(dynamoClient, dbConfig.Table)

	return &Dependencies{
		Config:                cfg,
		DynamoDBClient:        dynamoClient,
		DynamoDBConfig:        dbConfig,
		MachineRepository:     machineRepo,
		MaintenanceRepository: maintenanceRepo,
	}
}
