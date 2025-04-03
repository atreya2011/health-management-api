package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/config"
	applog "github.com/atreya2011/health-management-api/internal/infrastructure/log"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres"
)

var (
	days int
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
		if errors.Is(err, domain.ErrUserNotFound) {
			// Create a new test user
			newUser := &domain.User{
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
		
		record := &domain.BodyRecord{
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
	
	logger.Info("Mock data seeding completed successfully", "days", days)
}
