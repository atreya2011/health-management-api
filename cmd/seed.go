package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	"github.com/atreya2011/health-management-api/internal/config"
	applog "github.com/atreya2011/health-management-api/internal/log"
	"github.com/atreya2011/health-management-api/internal/persistence/postgres"
	"github.com/atreya2011/health-management-api/internal/testutil" // Added import
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
	userRepo := postgres.NewPgUserRepository(dbPool)
	bodyRecordRepo := postgres.NewPgBodyRecordRepository(dbPool)
	// Initialize other repositories as needed

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a test user if it doesn't exist
	testUser, err := userRepo.FindBySubjectID(ctx, "test-subject-id")
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) { // Use postgres error
			// Create a new test user
			newUser := &postgres.User{ // Use postgres.User
				SubjectID: "test-subject-id",
			}
			if err := userRepo.Create(ctx, newUser); err != nil {
				logger.Error("Failed to create test user", "error", err)
				return
			}
			logger.Info("Created test user", "subjectID", "test-subject-id")

			// Get the created user
			testUser, err = userRepo.FindBySubjectID(ctx, "test-subject-id")
			if err != nil {
				logger.Error("Failed to retrieve newly created test user", "error", err)
				return
			}
		} else {
			logger.Error("Failed to check for test user", "error", err)
			return
		}
	}

	logger.Info("Using test user for mock data", "userID", testUser.ID)

	// Create mock body records for the specified number of days
	now := time.Now().UTC().Truncate(24 * time.Hour)

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)

		// Generate some realistic but varying data
		weight := 70.0 + float64(i%5)
		bodyFat := 15.0 + float64(i%3)

		record := &postgres.BodyRecord{ // Use postgres.BodyRecord
			UserID:            testUser.ID,
			Date:              date,
			WeightKg:          &weight,
			BodyFatPercentage: &bodyFat,
		}

		_, err := bodyRecordRepo.Save(ctx, record)
		if err != nil {
			logger.Warn("Failed to create mock body record", "date", date, "error", err)
			continue
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
