package deps

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	AppName string `env:"APP_NAME"`
	Env     string `env:"APP_ENV" envDefault:"development"`

	// PostgreSQL Configuration
	DATABASE_URL string `env:"DATABASE_URL" envDefault:"postgres://ralts_user:ralts_password@localhost:5432/ralts_cms"`

	// Server Configuration
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`

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
}

func LoadConfig() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	cfg := Config{}

	if appEnv != "" {
		log.Printf("Loading %s config\n", appEnv)
		err := godotenv.Load(dir(".env." + appEnv))
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading app config: %w", err)
		}
		if err != nil && os.IsNotExist(err) {
			log.Printf("Warning: .env.%s not found, continuing without it", appEnv)
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error loading app config: %w", err)
	}

	return &cfg, nil
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
