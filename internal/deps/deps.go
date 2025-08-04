package deps

import (
	"context"
	"log"

	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/users"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies holds all application dependencies including configuration,
// database connections, and repository instances
type Dependencies struct {
	Config *Config

	PostgresClient *pgxpool.Pool

	// Repositories
	MachinesRepository    machines.Repository
	MaintenanceRepository maintenance.Repository
	UsersRepository       users.Repository
}

// Initialise creates and returns a new Dependencies instance with all required
// services, repositories, and database connections initialized
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
	machinesRepo := machines.NewRepository(pgClient)
	maintenanceRepo := maintenance.NewRepository(pgClient)
	usersRepo := users.NewRepository(pgClient)

	return &Dependencies{
		Config:                cfg,
		PostgresClient:        pgClient,
		MachinesRepository:    machinesRepo,
		MaintenanceRepository: maintenanceRepo,
		UsersRepository:       usersRepo,
	}
}
