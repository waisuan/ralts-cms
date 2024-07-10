package deps

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

type Database struct {
	Hostname string `env:"DB_HOSTNAME" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	Username string `env:"DB_USERNAME" envDefault:"postgres"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres"`
	DbName   string `env:"DB_NAME" envDefault:"postgres"`
}

type Config struct {
	AppName string `env:"APP_NAME" envDefault:"ralts-cms"`
	Env     string `env:"APP_ENV" envDefault:"development"`
	DB      Database
}

func LoadConfig() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	cfg := Config{}

	if "" != appEnv {
		log.Printf("Loading %s config\n", appEnv)
		err := godotenv.Load(dir(".env." + appEnv))
		if err != nil {
			return nil, fmt.Errorf("error loading app config: %w", err)
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
