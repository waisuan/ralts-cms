package deps

import (
	"context"
	"log"

	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Config *Config

	PostgresClient *pgxpool.Pool

	// Repositories
	MachineRepository     machines.Repository
	MaintenanceRepository maintenance.Repository
}

func Initialise() *Dependencies {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %e", err)
	}

	// Initialize PostgreSQL client
	pgClient, err := NewPostgresClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize PostgreSQL pool: %v", err)
	}

	// Initialize repositories (update as needed to use pgPool)
	machineRepo := machines.NewRepository(pgClient)
	maintenanceRepo := maintenance.NewRepository(pgClient)

	return &Dependencies{
		Config:                cfg,
		PostgresClient:        pgClient,
		MachineRepository:     machineRepo,
		MaintenanceRepository: maintenanceRepo,
	}
}
