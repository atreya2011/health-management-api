package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/atreya2011/health-management-api/internal/auth" // Added for UserContextKey
	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/atreya2011/health-management-api/internal/testutil" // Keep for CreateTestUser
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Import the postgres driver
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	testQueries *db.Queries
	testPool    *pgxpool.Pool
	testLogger  *slog.Logger
	testUserID  uuid.UUID

	// Keep track of these for teardown
	dockerPool *dockertest.Pool
	resource   *dockertest.Resource
	sqlDB      *sql.DB // For migrations
)

func runMigrations(db *sql.DB) error {
	migrationFiles, err := findMigrationFiles()
	if err != nil {
		return fmt.Errorf("could not find migration files: %w", err)
	}

	// Sort migration files to ensure they run in the correct order (Glob usually sorts, but explicit sort is safer if needed)
	// sort.Strings(migrationFiles) // If needed

	for _, file := range migrationFiles {
		if strings.HasSuffix(file, ".up.sql") {
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("could not read migration file %s: %w", file, err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				// Provide more context on migration error
				return fmt.Errorf("could not execute migration file %s: %w", file, err)
			}
			log.Printf("Executed migration: %s", file) // Use standard log in TestMain
		}
	}
	return nil
}

func findMigrationFiles() ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// Walk up to find the project root containing 'db/migrations'
	for {
		migrationsDir := filepath.Join(dir, "db", "migrations")
		stat, err := os.Stat(migrationsDir)
		if err == nil && stat.IsDir() {
			// Found the migrations directory
			files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
			if err != nil {
				return nil, fmt.Errorf("error listing files in %s: %w", migrationsDir, err)
			}
			if len(files) == 0 {
				return nil, fmt.Errorf("no *.sql files found in %s", migrationsDir)
			}
			log.Printf("Found %d migration files in %s", len(files), migrationsDir)
			return files, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory without finding migrations
			break
		}
		dir = parent
	}
	// If loop finishes without finding, try a relative path from CWD as fallback
	cwd, _ := os.Getwd()
	// Assuming tests run from internal/rpc/handlers, the relative path is ../../db/migrations
	relMigrationsDir := filepath.Join(cwd, "..", "..", "db", "migrations")
	if _, err := os.Stat(relMigrationsDir); err == nil {
		log.Printf("Falling back to relative path: %s", relMigrationsDir)
		files, err := filepath.Glob(filepath.Join(relMigrationsDir, "*.sql"))
		if err != nil {
			return nil, fmt.Errorf("error listing files in relative path %s: %w", relMigrationsDir, err)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no *.sql files found in relative path %s", relMigrationsDir)
		}
		return files, nil
	}

	return nil, fmt.Errorf("could not find migrations directory 'db/migrations' by walking up from %s or using relative path %s", cwd, relMigrationsDir)
}

func TestMain(m *testing.M) {
	testLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})) // More verbose logging for setup
	testLogger.Info("Setting up test database for handlers package...")

	var err error

	// --- Start Database Setup ---
	dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Increase startup timeout slightly?
	dockerPool.MaxWait = 120 * time.Second // Example: 2 minutes

	resource, err = dockerPool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=testdb_handlers",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// Set expiry only after successful start
	if err = resource.Expire(120); err != nil { // 120 seconds = 2 minutes
		log.Printf("Warning: Could not set container expiry: %s", err) // Log warning, don't fail
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s/testdb_handlers?sslmode=disable", hostAndPort)
	testLogger.Info("Database container started", "url", fmt.Sprintf("postgres://postgres:postgres@%s/testdb_handlers", hostAndPort))

	// Exponential backoff/retry for connecting to the database.
	if err = dockerPool.Retry(func() error {
		var retryErr error
		sqlDB, retryErr = sql.Open("postgres", databaseURL)
		if retryErr != nil {
			testLogger.Warn("Retrying DB connection (sql.Open)...", "error", retryErr)
			return retryErr
		}
		pingErr := sqlDB.Ping()
		if pingErr != nil {
			testLogger.Warn("Retrying DB connection (Ping)...", "error", pingErr)
			sqlDB.Close()  // Close connection if ping fails
			return pingErr // Return ping error
		}
		testLogger.Info("Database connection successful.")
		return nil // Connection successful
	}); err != nil {
		log.Printf("Attempting to purge container after connection failure...")
		_ = dockerPool.Purge(resource) // Attempt cleanup
		log.Fatalf("Could not connect to database container after retries: %s", err)
	}

	// Run migrations using the copied functions
	testLogger.Info("Running database migrations...")
	if err := runMigrations(sqlDB); err != nil {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = dockerPool.Purge(resource)
		log.Fatalf("Could not run migrations: %s", err)
	}
	testLogger.Info("Migrations completed successfully.")

	// Create pgx pool *after* migrations
	testPool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = dockerPool.Purge(resource)
		log.Fatalf("Could not create pgx connection pool: %s", err)
	}

	testQueries = db.New(testPool)
	// --- End Database Setup ---

	// --- Start Seeding Data ---
	testLogger.Info("Seeding initial test data...")
	ctx := context.Background()                                     // Use a background context for setup
	createdUserID, err := testutil.CreateTestUser(ctx, testQueries) // Use the exported helper
	if err != nil {
		// Teardown before fatal logging
		testPool.Close()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = dockerPool.Purge(resource)
		log.Fatalf("Failed to create test user for TestMain: %v", err)
	}
	testUserID = createdUserID
	testLogger.Info("Test user created", "userID", testUserID)
	// --- End Seeding Data ---

	testLogger.Info("Running tests...")
	exitCode := m.Run()

	// --- Start Teardown ---
	testLogger.Info("Tearing down test database...")
	// Use defer to ensure cleanup happens even on panic during tests
	// defer func() { ... }() // Alternative approach, but explicit call after m.Run() is standard

	if testPool != nil {
		testPool.Close()
		testLogger.Debug("Closed pgx pool.")
	}
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Could not close migration DB connection during teardown: %s", err)
		} else {
			testLogger.Debug("Closed migration DB connection.")
		}
	}
	// Purge the container using the stored dockerPool and resource
	if resource != nil && dockerPool != nil {
		testLogger.Debug("Purging docker container...")
		if err := dockerPool.Purge(resource); err != nil {
			log.Printf("Could not purge resource during teardown: %s", err)
		} else {
			testLogger.Debug("Docker container purged.")
		}
	}
	testLogger.Info("Teardown complete.")
	// --- End Teardown ---

	os.Exit(exitCode)
}

// Define testContext here if it's needed across multiple files in the package
// It simplifies passing the global testUserID to handlers via context.
type testContext struct {
	context.Context
}

func (c testContext) Value(key interface{}) interface{} {
	if key == auth.UserContextKey {
		return testUserID // Return the global testUserID
	}
	return c.Context.Value(key)
}

// Helper to create a new test context easily
func newTestContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background() // Ensure we don't wrap a nil context
	}
	return testContext{Context: ctx}
}

// resetDB truncates data tables to ensure test isolation.
// It keeps the user created in TestMain.
func resetDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	// List tables to truncate. Order might matter due to foreign keys if not using CASCADE,
	// but CASCADE should handle it. Exclude the 'users' table for now.
	tables := []string{
		"body_records",
		"diary_entries",
		"exercise_records",
		"columns",
		// Add other data tables here if necessary
	}
	for _, table := range tables {
		// Use CASCADE to handle foreign key constraints if any exist between these tables.
		// If foreign keys point *to* the users table, truncating these should be fine.
		sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := pool.Exec(ctx, sql)
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
	testLogger.Debug("Truncated data tables", "tables", tables)
}
