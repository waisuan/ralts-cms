package deps

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDynamoDBConfig(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{
			name:     "development environment",
			env:      "development",
			expected: true,
		},
		{
			name:     "local environment",
			env:      "local",
			expected: true,
		},
		{
			name:     "production environment",
			env:      "production",
			expected: false,
		},
		{
			name:     "staging environment",
			env:      "staging",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Env:                tt.env,
				DynamoDBEndpoint:   "http://localhost:8000",
				DynamoDBRegion:     "us-east-1",
				DynamoDBTable:      "test-table",
				AwsAccessKeyID:     "test-key",
				AwsSecretAccessKey: "test-secret",
			}

			dbConfig := NewDynamoDBConfig(cfg)

			assert.Equal(t, cfg.DynamoDBEndpoint, dbConfig.Endpoint)
			assert.Equal(t, cfg.DynamoDBRegion, dbConfig.Region)
			assert.Equal(t, cfg.DynamoDBTable, dbConfig.Table)
			assert.Equal(t, tt.expected, dbConfig.UseLocalEndpoint)
			assert.Equal(t, cfg.AwsAccessKeyID, dbConfig.AccessKeyID)
			assert.Equal(t, cfg.AwsSecretAccessKey, dbConfig.SecretAccessKey)
		})
	}
}

func TestNewDynamoDBClient_LocalDevelopment(t *testing.T) {
	dbConfig := &DynamoDBConfig{
		Endpoint:         "http://localhost:8000",
		Region:           "us-east-1",
		Table:            "test-table",
		UseLocalEndpoint: true,
		AccessKeyID:      "local",
		SecretAccessKey:  "local",
	}

	client, err := NewDynamoDBClient(context.Background(), dbConfig)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for local development
	assert.NotNil(t, client)
}

func TestNewDynamoDBClient_Production(t *testing.T) {
	dbConfig := &DynamoDBConfig{
		Endpoint:         "https://dynamodb.us-east-1.amazonaws.com",
		Region:           "us-east-1",
		Table:            "test-table",
		UseLocalEndpoint: false,
		AccessKeyID:      "",
		SecretAccessKey:  "",
	}

	client, err := NewDynamoDBClient(context.Background(), dbConfig)
	require.NoError(t, err)
	require.NotNil(t, client)

	// The client should be created successfully for production
	assert.NotNil(t, client)
}

func TestDynamoDBConfig_Complete(t *testing.T) {
	cfg := &Config{
		Env:                "development",
		DynamoDBEndpoint:   "http://localhost:8000",
		DynamoDBRegion:     "us-east-1",
		DynamoDBTable:      "test-table",
		AwsAccessKeyID:     "test-key",
		AwsSecretAccessKey: "test-secret",
	}

	dbConfig := NewDynamoDBConfig(cfg)

	// Test all fields are set correctly
	assert.Equal(t, "http://localhost:8000", dbConfig.Endpoint)
	assert.Equal(t, "us-east-1", dbConfig.Region)
	assert.Equal(t, "test-table", dbConfig.Table)
	assert.True(t, dbConfig.UseLocalEndpoint)
	assert.Equal(t, "test-key", dbConfig.AccessKeyID)
	assert.Equal(t, "test-secret", dbConfig.SecretAccessKey)
}

func TestDynamoDBConfig_ProductionSettings(t *testing.T) {
	cfg := &Config{
		Env:                "production",
		DynamoDBEndpoint:   "https://dynamodb.us-east-1.amazonaws.com",
		DynamoDBRegion:     "us-east-1",
		DynamoDBTable:      "prod-table",
		AwsAccessKeyID:     "",
		AwsSecretAccessKey: "",
	}

	dbConfig := NewDynamoDBConfig(cfg)

	// Test production settings
	assert.Equal(t, "https://dynamodb.us-east-1.amazonaws.com", dbConfig.Endpoint)
	assert.Equal(t, "us-east-1", dbConfig.Region)
	assert.Equal(t, "prod-table", dbConfig.Table)
	assert.False(t, dbConfig.UseLocalEndpoint)
	assert.Equal(t, "", dbConfig.AccessKeyID)
	assert.Equal(t, "", dbConfig.SecretAccessKey)
}
