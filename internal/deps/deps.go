package deps

import (
	"context"
	"log"
	"log/slog"
	"os"

	"ralts-cms/internal/attachments"
	"ralts-cms/internal/audit"
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
	Logger *slog.Logger

	PostgresClient *pgxpool.Pool
	S3Client       pkgs3.Client

	// Repositories
	MachinesRepository    machines.Repository
	MaintenanceRepository maintenance.Repository
	UsersRepository       users.Repository
	AuditRepository       audit.Repository

	// Services
	AttachmentService attachments.AttachmentService
	AuditService      audit.AuditService
}

// Initialise creates and returns a new Dependencies instance with all required
// services, repositories, and database connections initialized
func Initialise() *Dependencies {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize structured logger
	logger := newLogger(cfg.Env)

	// Set as default logger for the application
	slog.SetDefault(logger)

	// Initialize PostgreSQL client
	pgClient, err := NewPostgresClient(context.Background(), cfg)
	if err != nil {
		logger.Error("Failed to initialize PostgreSQL pool", "error", err)
		os.Exit(1)
	}

	// Initialize S3 client
	s3Client, err := NewS3Client(context.Background(), cfg)
	if err != nil {
		logger.Error("Failed to initialize S3 client", "error", err)
		os.Exit(1)
	}

	// Initialize repositories (update as needed to use pgPool)
	machinesRepo := machines.NewRepository(pgClient)
	maintenanceRepo := maintenance.NewRepository(pgClient)
	usersRepo := users.NewRepository(pgClient)
	auditRepo := audit.NewRepository(pgClient)

	// Initialize services
	attachmentService := attachments.NewService(s3Client, cfg.S3BucketName)

	// Initialize audit service with configured retention and cleanup interval
	auditConfig := audit.ServiceConfig{
		BufferSize:      audit.DefaultBufferSize,
		RetentionDays:   cfg.AuditRetentionDays,
		CleanupInterval: cfg.AuditCleanupInterval,
	}
	auditService := audit.NewServiceWithConfig(auditRepo, logger, auditConfig)
	auditService.Start()

	return &Dependencies{
		Config:                cfg,
		Logger:                logger,
		PostgresClient:        pgClient,
		S3Client:              s3Client,
		MachinesRepository:    machinesRepo,
		MaintenanceRepository: maintenanceRepo,
		UsersRepository:       usersRepo,
		AuditRepository:       auditRepo,
		AttachmentService:     attachmentService,
		AuditService:          auditService,
	}
}

// Shutdown gracefully shuts down all services and connections
func (d *Dependencies) Shutdown(ctx context.Context) {
	d.Logger.Info("Shutting down dependencies...")

	// Stop the audit service first (drains remaining events)
	if d.AuditService != nil {
		d.AuditService.Stop(ctx)
	}

	// Close database connection
	if d.PostgresClient != nil {
		d.PostgresClient.Close()
	}

	d.Logger.Info("Dependencies shutdown complete")
}

// newLogger creates a structured logger based on the environment
func newLogger(env string) *slog.Logger {
	var handler slog.Handler

	// Configure logger based on environment
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Default to INFO level
	}

	// In development, use more verbose logging and text format for readability
	if env == appEnvDevelopment {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// In production, use JSON format for structured logging
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
