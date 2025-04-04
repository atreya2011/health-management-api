package testutil

import (
	"context"
	// "database/sql" // Removed unused import
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewColumnRepository creates a new column repository for testing
func NewColumnRepository(pool *pgxpool.Pool) domain.ColumnRepository {
	return postgres.NewPgColumnRepository(pool)
}

// CreateTestColumn creates a test column in the database using the provided pool
func CreateTestColumn(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, title, content string,
	category pgtype.Text, tags []string, publishedAt pgtype.Timestamptz) error { // Use pgtype

	// Create the column in the database
	query := `
		INSERT INTO columns (id, title, content, category, tags, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	// TIMESTAMPTZ columns usually map directly to time.Time with pgx/v5
	// Ensure times are UTC before insertion
	nowUTC := time.Now().UTC()
	publishedAtVal := publishedAt // Use the passed pgtype.Timestamptz directly

	_, err := pool.Exec(ctx, query,
		id,
		title,
		content,
		category,       // Pass pgtype.Text directly
		tags,           // Pass []string directly (pgx handles text arrays)
		publishedAtVal, // Pass pgtype.Timestamptz directly
		nowUTC,         // Pass time.Time (UTC) for created_at
		nowUTC,         // Pass time.Time (UTC) for updated_at
	)
	
	if err != nil {
		return fmt.Errorf("could not create test column: %w", err)
	}
	
	return nil
}

// SeedMockColumns deletes existing columns and inserts a predefined set of mock columns.
// This is useful for setting up a consistent state for tests or seeding development environments.
func SeedMockColumns(ctx context.Context, pool *pgxpool.Pool) error {
	// Delete existing columns to avoid conflicts and ensure a clean state
	_, err := pool.Exec(ctx, "DELETE FROM columns")
	if err != nil {
		return fmt.Errorf("failed to delete existing columns: %w", err)
	}

	// Define mock columns
	columns := []*domain.Column{
		{
			ID:      uuid.New(),
			Title:   "Health Tips for Daily Life",
			Content: "Content about health tips...",
			Category: func() *string {
				s := "health"
				return &s
			}(),
			Tags:        []string{"health", "wellness"},
			PublishedAt: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:      uuid.New(),
			Title:   "Diet Strategies for Weight Loss",
			Content: "Content about diet strategies...",
			Category: func() *string {
				s := "nutrition"
				return &s
			}(),
			Tags:        []string{"diet", "nutrition", "health"},
			PublishedAt: func() *time.Time { t := time.Now().Add(-48 * time.Hour); return &t }(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:      uuid.New(),
			Title:   "Exercise Routines for Beginners",
			Content: "Content about exercise routines...",
			Category: func() *string {
				s := "fitness"
				return &s
			}(),
			Tags:        []string{"exercise", "fitness"},
			PublishedAt: func() *time.Time { t := time.Now().Add(-72 * time.Hour); return &t }(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Insert mock columns using CreateTestColumn
	for _, column := range columns {
		// Convert domain types to pgtype types needed by CreateTestColumn
		var categoryVal pgtype.Text
		if column.Category != nil {
			categoryVal = pgtype.Text{String: *column.Category, Valid: true}
		}

		var publishedAtVal pgtype.Timestamptz
		if column.PublishedAt != nil {
			publishedAtVal = pgtype.Timestamptz{Time: column.PublishedAt.UTC(), Valid: true} // Ensure UTC
		}

		// Call CreateTestColumn for each mock column
		err := CreateTestColumn(ctx, pool, column.ID, column.Title, column.Content, categoryVal, column.Tags, publishedAtVal)
		if err != nil {
			// Log or handle the error for the specific column creation failure
			fmt.Printf("Warning: Failed to create mock column '%s' using CreateTestColumn: %v\n", column.Title, err)
		}
	}

	return nil // Indicate success even if some individual insertions failed
}
