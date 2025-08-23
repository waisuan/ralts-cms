package testutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //nolint:blank-imports
	_ "github.com/golang-migrate/migrate/v4/source/file"       //nolint:blank-imports
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabase represents a test database instance with testcontainers
type TestDatabase struct {
	Container      *postgres.PostgresContainer
	PostgresClient *pgxpool.Pool
	ConnStr        string
}

// DatabaseConfig holds configuration for test database setup
type DatabaseConfig struct {
	DatabaseName string
	Username     string
	Password     string
	Image        string
}

// SetupTestDatabase creates a new PostgreSQL testcontainer, runs migrations,
// and returns a TestDatabase instance ready for testing
func SetupTestDatabase(t *testing.T) *TestDatabase {
	ctx := context.Background()
	config := DatabaseConfig{
		DatabaseName: "test_db",
		Username:     "test_user",
		Password:     "test_password",
		Image:        "postgres:16-alpine",
	}

	// Start PostgreSQL container with specified configuration
	container, err := postgres.Run(
		ctx,
		config.Image,
		postgres.WithDatabase(config.DatabaseName),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		postgres.WithSQLDriver("pgx"), // Use pgx driver for better performance
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				// First, wait for the container to log readiness twice
				// This is because PostgreSQL restarts itself after the first startup
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
				// Then, wait for docker to actually serve the port on localhost
				// This is important for non-Linux OSes like Mac and Windows
				wait.ForListeningPort("5432/tcp"),
			),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Run database migrations
	if err := runMigrations(connStr); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping database: %v", err)
	}

	testDB := &TestDatabase{
		Container:      container,
		PostgresClient: pool,
		ConnStr:        connStr,
	}

	// Setup cleanup - this will run when the test finishes
	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// Close closes the database connection and terminates the container
func (db *TestDatabase) Close() {
	if db.PostgresClient != nil {
		db.PostgresClient.Close()
	}

	if db.Container != nil {
		if err := testcontainers.TerminateContainer(db.Container); err != nil {
			// Log error but don't fail the test during cleanup
			fmt.Printf("Warning: failed to terminate container: %v\n", err)
		}
	}
}

// CleanupAllTables truncates all tables in the database while preserving schema
// This is useful for cleaning up between tests without recreating the container
func (db *TestDatabase) CleanupAllTables(ctx context.Context) error {
	// Get all table names except system tables and migration schema
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'schema_migrations'
		ORDER BY tablename;
	`

	rows, err := db.PostgresClient.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating table rows: %w", err)
	}

	// Truncate all tables in reverse order to handle foreign key constraints
	for i := len(tables) - 1; i >= 0; i-- {
		truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", tables[i])
		if _, err := db.PostgresClient.Exec(ctx, truncateQuery); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", tables[i], err)
		}
	}

	return nil
}

// runMigrations executes all pending database migrations
func runMigrations(connStr string) error {
	// Find the project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Construct migrations path
	migrationsPath := "file://" + filepath.Join(projectRoot, "db", "migrations")

	// Create migrate instance
	m, err := migrate.New(migrationsPath, connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			fmt.Printf("Warning: failed to close migrate instance: source=%v, db=%v\n", sourceErr, dbErr)
		}
	}()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// findProjectRoot searches for the project root directory by looking for go.mod
// This is similar to the dir() function in internal/deps/config.go
func findProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return "", fmt.Errorf("go.mod not found - not in a Go module")
		}
		currentDir = parent
	}
}
