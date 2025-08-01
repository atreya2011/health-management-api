package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Import the postgres driver
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// TestDatabase represents a test database instance
type TestDatabase struct {
	Pool     *pgxpool.Pool
	Resource *dockertest.Resource
	Pool2    *dockertest.Pool
	DB       *sql.DB
	Queries  *db.Queries
}

// SetupTestDatabase creates a new PostgreSQL container and sets up the database
func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	// Uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	// Pull the PostgreSQL image
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		// Set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	// Set a reasonable timeout for container startup
	err = resource.Expire(120) // Tell docker to hard kill the container in 120 seconds
	if err != nil {
		t.Fatalf("Could not set container expiry: %s", err)
	}

	// Get the container's host and port
	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s/testdb?sslmode=disable", hostAndPort)

	// Wait for the database to be ready
	var sqlDB *sql.DB
	if err = pool.Retry(func() error {
		var err error
		sqlDB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	}); err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		t.Fatalf("Could not run migrations: %s", err)
	}

	// Create a pgx connection pool
	pgxPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Could not create connection pool: %s", err)
	}

	// Initialize sqlc queries directly with the pool (pgx/v5 mode)
	sqlcQueries := db.New(pgxPool) // Pass pool directly

	return &TestDatabase{
		Pool:     pgxPool,
		Resource: resource,
		Pool2:    pool,
		DB:       sqlDB,
		Queries:  sqlcQueries,
	}
}

// TeardownTestDatabase cleans up the test database
func (td *TestDatabase) TeardownTestDatabase(t *testing.T) {
	t.Helper()

	// Close the pgx connection pool
	if td.Pool != nil {
		td.Pool.Close()
	}

	// Close the sql.DB connection
	if td.DB != nil {
		if err := td.DB.Close(); err != nil {
			t.Logf("Could not close database connection: %s", err)
		}
	}

	// Kill and remove the container
	if td.Resource != nil {
		if err := td.Pool2.Purge(td.Resource); err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	}
}

// runMigrations runs the database migrations
func runMigrations(db *sql.DB) error {
	// Find the migrations directory
	migrationFiles, err := findMigrationFiles()
	if err != nil {
		return fmt.Errorf("could not find migration files: %w", err)
	}

	// Sort migration files to ensure they run in the correct order
	for _, file := range migrationFiles {
		if strings.HasSuffix(file, ".up.sql") {
			// Read the migration file
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("could not read migration file %s: %w", file, err)
			}

			// Execute the migration
			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("could not execute migration file %s: %w", file, err)
			}

			log.Printf("Executed migration: %s", file)
		}
	}

	return nil
}

// findMigrationFiles finds all migration files in the db/migrations directory
func findMigrationFiles() ([]string, error) {
	// Start from the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Try to find the migrations directory by walking up the directory tree
	for {
		migrationsDir := filepath.Join(dir, "db", "migrations")
		if _, err := os.Stat(migrationsDir); err == nil {
			// Found the migrations directory
			files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
			if err != nil {
				return nil, err
			}
			return files, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("could not find migrations directory")
}
