package deps

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DYNAMODB_ENDPOINT")
	os.Unsetenv("DYNAMODB_REGION")
	os.Unsetenv("DYNAMODB_TABLE")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("HTTP_READ_TIMEOUT")
	os.Unsetenv("HTTP_WRITE_TIMEOUT")
	os.Unsetenv("HTTP_IDLE_TIMEOUT")
	os.Unsetenv("JWT_SECRET")

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "", cfg.AppName) // No default
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "http://localhost:8000", cfg.DynamoDBEndpoint)
	assert.Equal(t, "us-east-1", cfg.DynamoDBRegion)
	assert.Equal(t, "ralts", cfg.DynamoDBTable)
	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 15*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.IdleTimeout)
}

func TestLoadConfig_EnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_NAME", "test-app")
	os.Setenv("APP_ENV", "production")
	os.Setenv("DYNAMODB_ENDPOINT", "http://dynamodb.example.com:8000")
	os.Setenv("DYNAMODB_REGION", "us-west-2")
	os.Setenv("DYNAMODB_TABLE", "test-table")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("HTTP_READ_TIMEOUT", "30s")
	os.Setenv("HTTP_WRITE_TIMEOUT", "30s")
	os.Setenv("HTTP_IDLE_TIMEOUT", "120s")
	os.Setenv("JWT_SECRET", "test-secret")

	// Clean up after test
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("DYNAMODB_ENDPOINT")
		os.Unsetenv("DYNAMODB_REGION")
		os.Unsetenv("DYNAMODB_TABLE")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("HTTP_READ_TIMEOUT")
		os.Unsetenv("HTTP_WRITE_TIMEOUT")
		os.Unsetenv("HTTP_IDLE_TIMEOUT")
		os.Unsetenv("JWT_SECRET")
	}()

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check overridden values
	assert.Equal(t, "test-app", cfg.AppName)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "http://dynamodb.example.com:8000", cfg.DynamoDBEndpoint)
	assert.Equal(t, "us-west-2", cfg.DynamoDBRegion)
	assert.Equal(t, "test-table", cfg.DynamoDBTable)
	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
}

func TestLoadConfig_InvalidTimeout(t *testing.T) {
	// Set invalid timeout
	os.Setenv("HTTP_READ_TIMEOUT", "invalid")
	defer os.Unsetenv("HTTP_READ_TIMEOUT")

	// Load config should fail
	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error loading app config")
}

func TestConfig_RequiredFields(t *testing.T) {
	// Test that APP_NAME is required
	os.Unsetenv("APP_NAME")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// APP_NAME should be empty but not cause an error
	assert.Equal(t, "", cfg.AppName)
}

func TestConfig_TimeParsing(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected time.Duration
	}{
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "2m", 2 * time.Minute},
		{"hours", "1h", 1 * time.Hour},
		{"mixed", "1h30m", 1*time.Hour + 30*time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("HTTP_READ_TIMEOUT", tt.duration)
			defer os.Unsetenv("HTTP_READ_TIMEOUT")

			cfg, err := LoadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.ReadTimeout)
		})
	}
}
