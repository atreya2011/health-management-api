package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	// App imports
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	"github.com/atreya2011/health-management-api/internal/infrastructure/config"
	applog "github.com/atreya2011/health-management-api/internal/infrastructure/log"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres"
	"github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1/healthappv1connect"
	"github.com/atreya2011/health-management-api/internal/infrastructure/rpc/handlers"
)

// seedMockData adds mock data to the database for testing purposes
func seedMockData(ctx context.Context, userRepo domain.UserRepository, bodyRecordRepo domain.BodyRecordRepository, logger *slog.Logger) {
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
	
	// Create mock body records for the past 30 days
	now := time.Now().UTC().Truncate(24 * time.Hour)
	
	for i := 0; i < 30; i++ {
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
	}
	
	logger.Info("Mock data seeded successfully")
}

func main() {
	// Initialize logger
	logger := applog.NewLogger()
	logger.Info("Starting server...")

	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database connection
	dbPool, err := postgres.NewDBPool(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("Database connection pool established")

	// Initialize repositories
	userRepo := postgres.NewPgUserRepository(dbPool)
	bodyRecordRepo := postgres.NewPgBodyRecordRepository(dbPool)
	// Initialize other repositories as needed
	
	// Seed mock data
	logger.Info("Seeding mock data...")
	seedMockData(context.Background(), userRepo, bodyRecordRepo, logger)

	// Initialize application services
	bodyRecordService := application.NewBodyRecordService(bodyRecordRepo, logger)
	// Initialize other services as needed

	// Initialize auth interceptor
	jwtConfig := &auth.JWTConfig{
		SecretKey: cfg.JWT.SecretKey,
	}
	authInterceptor := auth.AuthInterceptor(jwtConfig, userRepo, logger)

	// Create interceptors
	interceptors := connect.WithInterceptors(
		authInterceptor,
		// Add more interceptors here (logging, metrics, recovery)
	)

	// Initialize handlers
	bodyRecordHandler := handlers.NewBodyRecordHandler(bodyRecordService, logger)
	// Initialize other handlers as needed

	// Create router
	mux := http.NewServeMux()

	// Register Connect handlers
	mux.Handle(healthappv1connect.NewBodyRecordServiceHandler(bodyRecordHandler, interceptors))
	// Register other handlers as needed

	// Serve OpenAPI spec
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", http.FileServer(http.Dir("./third_party/openapi"))))

	// Configure server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server listening", "address", addr)

	// Create server with h2c for HTTP/2 without TLS
	server := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	logger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server shutdown gracefully")
}
