package deps

import (
	"gorm.io/gorm"
	"log"
)

type Dependencies struct {
	Config *Config
	DB     *gorm.DB
}

func Initialise() *Dependencies {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %e", err)
	}

	db, err := InitPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to initialise db: %e", err)
	}

	return &Dependencies{
		Config: cfg,
		DB:     db,
	}
}
