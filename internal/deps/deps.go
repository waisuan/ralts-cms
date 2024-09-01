package deps

import (
	"gorm.io/gorm"
	"log"
	"ralts-cms/internal/machines"
)

type Dependencies struct {
	Config            *Config
	DB                *gorm.DB
	MachineRepository machines.Repository
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

	mr := machines.NewRepository(db)

	return &Dependencies{
		Config:            cfg,
		DB:                db,
		MachineRepository: mr,
	}
}
