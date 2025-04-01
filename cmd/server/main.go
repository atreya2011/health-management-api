package main

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
