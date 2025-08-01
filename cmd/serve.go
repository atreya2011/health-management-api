package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	// App imports
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	"github.com/atreya2011/health-management-api/internal/infrastructure/config"
	applog "github.com/atreya2011/health-management-api/internal/infrastructure/log"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres"
	"github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1/healthappv1connect"
	"github.com/atreya2011/health-management-api/internal/infrastructure/rpc/handlers"
)

var (
	port string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the Health Management API server with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Local flags
	serveCmd.Flags().StringVarP(&port, "port", "p", "", "port to run the server on (overrides config)")
}

func runServer() {
	// Initialize logger
	logger := applog.NewLogger()
	if verboseMode {
		// Set more verbose logging when verbose flag is enabled
		logger.Info("Verbose mode enabled")
	}
	logger.Info("Starting server...")

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Override port if specified via flag
	if port != "" {
		cfg.Server.Port = port
		logger.Info("Using port from command line flag", "port", port)
	}

	// Initialize database connection
	logger.Info("Connecting to database...", "url", cfg.Database.URL)
	dbPool, err := postgres.NewDBPool(&cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("Database connection pool established")

	// Check if the database schema is correct
	schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer schemaCancel()

	var columnName string
	err = dbPool.QueryRow(schemaCtx, "SELECT column_name FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'subject_id'").Scan(&columnName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("Database schema is incorrect. Make sure migrations have been applied.", "error", "subject_id column not found in users table")
		} else {
			logger.Error("Failed to check database schema", "error", err)
		}
		logger.Info("Run 'make migrate-up' to apply migrations")
		os.Exit(1)
	}
	logger.Info("Database schema verified", "column_name", columnName)

	// Initialize repositories
	userRepo := postgres.NewPgUserRepository(dbPool)
	bodyRecordRepo := postgres.NewPgBodyRecordRepository(dbPool)
	diaryEntryRepo := postgres.NewPgDiaryEntryRepository(dbPool)
	exerciseRecordRepo := postgres.NewPgExerciseRecordRepository(dbPool)
	columnRepo := postgres.NewPgColumnRepository(dbPool)

	// Initialize application services
	bodyRecordService := application.NewBodyRecordService(bodyRecordRepo, logger)
	diaryService := application.NewDiaryService(diaryEntryRepo, logger)
	exerciseRecordService := application.NewExerciseRecordService(exerciseRecordRepo, logger)
	columnService := application.NewColumnService(columnRepo, logger)

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
	diaryHandler := handlers.NewDiaryHandler(diaryService, logger)
	exerciseRecordHandler := handlers.NewExerciseRecordHandler(exerciseRecordService, logger)
	columnHandler := handlers.NewColumnHandler(columnService, logger)

	// Create router
	mux := http.NewServeMux()

	// Register Connect handlers
	bodyRecordHandlerPath, bodyRecordServiceHandler := healthappv1connect.NewBodyRecordServiceHandler(bodyRecordHandler, interceptors)
	mux.Handle(bodyRecordHandlerPath, bodyRecordServiceHandler)
	diaryHandlerPath, diaryServiceHandler := healthappv1connect.NewDiaryServiceHandler(diaryHandler, interceptors)
	mux.Handle(diaryHandlerPath, diaryServiceHandler)
	exerciseRecordHandlerPath, exerciseRecordServiceHandler := healthappv1connect.NewExerciseRecordServiceHandler(exerciseRecordHandler, interceptors)
	mux.Handle(exerciseRecordHandlerPath, exerciseRecordServiceHandler)
	// Column service doesn't require authentication
	columnHandlerPath, columnServiceHandler := healthappv1connect.NewColumnServiceHandler(columnHandler)
	mux.Handle(columnHandlerPath, columnServiceHandler)

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
