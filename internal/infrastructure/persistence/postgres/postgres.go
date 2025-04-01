package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDBPool creates a new PostgreSQL connection pool
func NewDBPool(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Create a connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing database URL: %w", err)
	}

	// Set pool configuration options
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create a connection context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection pool: %w", err)
	}

	// Ping the database to verify the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return pool, nil
}
