// Package deps provides dependency injection and configuration management
// for the Ralts-CMS application, including database and server configuration.
package deps

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

const (
	appEnvDevelopment = "development"
	appEnvTest        = "test"
	// jwtSecretDevDefault is the placeholder allowed only in development and automated test runs.
	jwtSecretDevDefault = "your-jwt-secret-key"
)

// Config holds all configuration settings for the Ralts-CMS application
type Config struct {
	AppName string `env:"APP_NAME"`
	Env     string `env:"APP_ENV" envDefault:"development"`

	// PostgreSQL Configuration
	DatabaseURL string `env:"DATABASE_URL"`

	// Server Configuration
	ServerPort string `env:"PORT" envDefault:"8080"`

	// HTTP Timeouts
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`

	// Authentication
	JWTSecret string `env:"JWT_SECRET" envDefault:"your-jwt-secret-key"`

	// API Configuration
	DefaultMachinesLimit    int32 `env:"DEFAULT_MACHINE_LIMIT" envDefault:"50"`
	MaxMachinesLimit        int64 `env:"MAX_MACHINE_LIMIT" envDefault:"100"`
	DefaultMaintenanceLimit int32 `env:"DEFAULT_MAINTENANCE_LIMIT" envDefault:"50"`
	MaxMaintenanceLimit     int64 `env:"MAX_MAINTENANCE_LIMIT" envDefault:"100"`

	// AWS / S3 Configuration
	AWSS3BucketName     string `env:"AWS_S3_BUCKET_NAME" envDefault:"ralts-cms-attachments"`
	AWSAccessKeyID      string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey  string `env:"AWS_SECRET_ACCESS_KEY"`
	AWSDefaultRegion    string `env:"AWS_DEFAULT_REGION"`
	AWSEndpointURL      string `env:"AWS_ENDPOINT_URL"`
	AWSS3ForcePathStyle bool   `env:"AWS_S3_FORCE_PATH_STYLE" envDefault:"false"`

	// Audit Configuration
	AuditRetentionDays   int           `env:"AUDIT_RETENTION_DAYS" envDefault:"7"`
	AuditCleanupInterval time.Duration `env:"AUDIT_CLEANUP_INTERVAL" envDefault:"1h"`
}

// LoadConfig loads and parses configuration from environment variables and .env files
func LoadConfig() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	cfg := Config{}

	if appEnv != "" {
		slog.Info("Loading configuration", "environment", appEnv)
		err := godotenv.Load(dir(".env." + appEnv))
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading app config: %w", err)
		}
		if err != nil && os.IsNotExist(err) {
			slog.Warn("Environment file not found, continuing without it", "file", ".env."+appEnv)
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error loading app config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Env == appEnvDevelopment || cfg.Env == appEnvTest {
		return nil
	}
	if cfg.JWTSecret == "" || cfg.JWTSecret == jwtSecretDevDefault {
		return fmt.Errorf("JWT_SECRET must be set to a strong secret when APP_ENV is not %q or %q (do not use the default placeholder)", appEnvDevelopment, appEnvTest)
	}
	return nil
}

// dir returns the absolute path of the given environment file (envFile) in the Go module's
// root directory. It searches for the 'go.mod' file from the current working directory upwards
// and appends the envFile to the directory containing 'go.mod'.
// It panics if it fails to find the 'go.mod' file.
func dir(envFile string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return filepath.Join(currentDir, envFile)
}
