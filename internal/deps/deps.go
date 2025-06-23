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

	// Repositories
	MachineRepository     machine.Repository
	MaintenanceRepository maintenance.Repository
}

func Initialise() *Dependencies {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %e", err)
	}

	// Initialize DynamoDB client
	dynamoClient, err := NewDynamoDBClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize DynamoDB client: %e", err)
	}

	// Validate DynamoDB connection
	if err := ValidateDynamoDBConnection(context.Background(), dynamoClient, cfg.DynamoDBTable); err != nil {
		log.Fatalf("failed to validate DynamoDB connection: %e", err)
	}

	// Initialize repositories
	machineRepo := machine.NewRepository(dynamoClient, cfg.DynamoDBTable)
	maintenanceRepo := maintenance.NewRepository(dynamoClient, cfg.DynamoDBTable)

	return &Dependencies{
		Config:                cfg,
		DynamoDBClient:        dynamoClient,
		MachineRepository:     machineRepo,
		MaintenanceRepository: maintenanceRepo,
	}
}
