package deps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDynamoDBClient_LocalDevelopment(t *testing.T) {
	cfg := &Config{
		Env:              "development",
		DynamoDBEndpoint: "http://localhost:8000",
		DynamoDBRegion:   "us-east-1",
		DynamoDBTable:    "test-table",
	}

	client, err := NewDynamoDBClient(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for local development
	assert.NotNil(t, client)
}

func TestNewDynamoDBClient_Production(t *testing.T) {
	cfg := &Config{
		Env:              "production",
		DynamoDBEndpoint: "https://dynamodb.us-east-1.amazonaws.com",
		DynamoDBRegion:   "us-east-1",
		DynamoDBTable:    "test-table",
	}

	client, err := NewDynamoDBClient(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for production
	assert.NotNil(t, client)
}

func TestNewDynamoDBClient_LocalEnvironment(t *testing.T) {
	cfg := &Config{
		Env:              "local",
		DynamoDBEndpoint: "http://localhost:8000",
		DynamoDBRegion:   "us-east-1",
		DynamoDBTable:    "test-table",
	}

	client, err := NewDynamoDBClient(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for local environment
	assert.NotNil(t, client)
}

func TestNewDynamoDBClient_StagingEnvironment(t *testing.T) {
	cfg := &Config{
		Env:              "staging",
		DynamoDBEndpoint: "https://dynamodb.us-east-1.amazonaws.com",
		DynamoDBRegion:   "us-east-1",
		DynamoDBTable:    "staging-table",
	}

	client, err := NewDynamoDBClient(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for staging environment
	assert.NotNil(t, client)
}
