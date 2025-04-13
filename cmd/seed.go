package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/atreya2011/health-management-api/internal/config"
	postgres "github.com/atreya2011/health-management-api/internal/db"
	applog "github.com/atreya2011/health-management-api/internal/log"
	"github.com/atreya2011/health-management-api/internal/testutil"
)

var (
	days int
	mock bool
)

// seedCmd represents the seed command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with mock data",
	Long:  `Seed the database with mock data for testing and development purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		runSeed()
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)

	// Local flags
	seedCmd.Flags().IntVarP(&days, "days", "d", 30, "number of days to generate mock data for")
	seedCmd.Flags().BoolVarP(&mock, "mock", "m", false, "seed mock data for testing")
}

func runSeed() {
	// Initialize logger
	logger := applog.NewLogger()
	if verboseMode {
		logger.Info("Verbose mode enabled")
	}
	logger.Info("Starting database seeding...")

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return
	}

	// Initialize database connection
	dbPool, err := postgres.NewDBPool(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer dbPool.Close()
	logger.Info("Database connection pool established")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(dbPool)
	bodyRecordRepo := postgres.NewBodyRecordRepository(dbPool)
	// Initialize other repositories as needed

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create or get a test user
	testUser, err := userRepo.Create(ctx, "test-subject-id")
	if err != nil {
		// Check if it's the specific "not found" error which Create now handles internally by fetching
		// If it's any other error during creation or fetching, log and return
		if !errors.Is(err, postgres.ErrUserNotFound) {
			logger.Error("Failed to create or retrieve test user", "error", err)
			return
		}
		// If ErrUserNotFound somehow bubbles up despite the logic in Create, log it.
		// This case shouldn't happen based on the current Create implementation.
		logger.Error("Unexpected ErrUserNotFound after Create call", "error", err)
		return
	}
	// If Create returns successfully, testUser contains either the newly created or existing user
	logger.Info("Ensured test user exists", "subjectID", "test-subject-id", "userID", testUser.ID)

	// Note: The original code fetched the user again after creation.
	// The modified Create method now returns the user directly (either new or existing),
	// so the second fetch is no longer necessary.

	// Check if context timed out after user creation/retrieval
	if ctx.Err() != nil {
		logger.Error("Context deadline exceeded after user creation/retrieval", "error", ctx.Err())
		return
	}

	// Original error handling for FindBySubjectID (kept for reference, but Create handles this now)
	// testUser, err := userRepo.FindBySubjectID(ctx, "test-subject-id")
	// if err != nil {
	// 	if errors.Is(err, postgres.ErrUserNotFound) { // Use postgres error
	// 		// Create logic was here...
	// 	} else {
	// 		logger.Error("Failed to check for test user", "error", err)
	// 		return
	// 	}
	// }

	// logger.Info("Using test user for mock data", "userID", testUser.ID) // Removed duplicate log and extra braces

	// Create mock body records for the specified number of days
	now := time.Now().UTC().Truncate(24 * time.Hour)

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)

		// Generate some realistic but varying data
		weight := 70.0 + float64(i%5)
		bodyFat := 15.0 + float64(i%3)

		// Call Save with individual arguments
		_, err := bodyRecordRepo.Save(ctx, testUser.ID, date, &weight, &bodyFat)
		if err != nil {
			logger.Warn("Failed to create mock body record", "date", date, "error", err)
			continue // Continue to next day even if one fails
		}

		if verboseMode {
			logger.Info("Created mock body record", "date", date, "weight", weight, "bodyFat", bodyFat)
		}
	}

	// Seed mock columns if requested
	if mock {
		logger.Info("Seeding mock columns using testutil...")
		err := testutil.SeedMockColumns(ctx, dbPool)
		if err != nil {
			logger.Error("Failed to seed mock columns", "error", err)
			// Decide if we should return or just log the error
			// For seeding, often logging is sufficient, but depends on requirements.
		} else {
			logger.Info("Mock columns seeded successfully")
		}
	}

	logger.Info("Mock data seeding completed successfully", "days", days, "mock", mock)
}
