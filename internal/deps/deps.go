package deps

import (
	"context"
	"log"

	"ralts-cms/internal/attachments"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
	"ralts-cms/internal/users"

	pkgs3 "ralts-cms/pkg/s3"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies holds all application dependencies including configuration,
// database connections, and repository instances
type Dependencies struct {
	Config *Config

	PostgresClient *pgxpool.Pool
	S3Client       pkgs3.Client

	// Repositories
	MachinesRepository    machines.Repository
	MaintenanceRepository maintenance.Repository
	UsersRepository       users.Repository

	// Services
	AttachmentService attachments.AttachmentService
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

	// Initialize S3 client
	s3Client, err := NewS3Client(context.Background(), cfg)
	if err != nil {
		log.Fatalf("failed to initialize S3 client: %v", err)
	}

	// Initialize repositories (update as needed to use pgPool)
	machinesRepo := machines.NewRepository(pgClient)
	maintenanceRepo := maintenance.NewRepository(pgClient)
	usersRepo := users.NewRepository(pgClient)

	// Initialize services
	attachmentService := attachments.NewService(s3Client, cfg.S3BucketName)

	return &Dependencies{
		Config:                cfg,
		PostgresClient:        pgClient,
		S3Client:              s3Client,
		MachinesRepository:    machinesRepo,
		MaintenanceRepository: maintenanceRepo,
		UsersRepository:       usersRepo,
		AttachmentService:     attachmentService,
	}
}
