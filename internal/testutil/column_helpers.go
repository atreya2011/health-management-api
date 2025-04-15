package testutil

import (
	"context"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestColumn creates a test column in the database using the provided pool and returns the created record.
func CreateTestColumn(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, title, content string,
	category pgtype.Text, tags []string, publishedAt pgtype.Timestamptz) (db.Column, error) { // Return db.Column and error

	// Create the column in the database
	insertQuery := `
		INSERT INTO columns (id, title, content, category, tags, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	nowUTC := time.Now().UTC()
	publishedAtVal := publishedAt

	_, err := pool.Exec(ctx, insertQuery,
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
		return db.Column{}, fmt.Errorf("could not insert test column: %w", err)
	}

	// Fetch the created column to return it
	selectQuery := `
		SELECT id, title, content, category, tags, published_at, created_at, updated_at
		FROM columns
		WHERE id = $1
	`
	row := pool.QueryRow(ctx, selectQuery, id)
	var createdColumn db.Column
	err = row.Scan(
		&createdColumn.ID,
		&createdColumn.Title,
		&createdColumn.Content,
		&createdColumn.Category,
		&createdColumn.Tags,
		&createdColumn.PublishedAt,
		&createdColumn.CreatedAt,
		&createdColumn.UpdatedAt,
	)
	if err != nil {
		return db.Column{}, fmt.Errorf("could not fetch created test column: %w", err)
	}

	return createdColumn, nil
}

// SeedMockColumns deletes existing columns and inserts a predefined set of mock columns.
// This is useful for setting up a consistent state for tests or seeding development environments.
func SeedMockColumns(ctx context.Context, pool *pgxpool.Pool) error {
	// Delete existing columns to avoid conflicts and ensure a clean state
	_, err := pool.Exec(ctx, "TRUNCATE TABLE columns RESTART IDENTITY CASCADE") // Use TRUNCATE for efficiency
	if err != nil {
		return fmt.Errorf("failed to delete existing columns: %w", err)
	}

	// Define mock data directly
	mockColumnsData := []struct {
		ID          uuid.UUID
		Title       string
		Content     string
		Category    string // Use string for simplicity, convert to pgtype.Text later
		Tags        []string
		PublishedAt time.Time // Use time.Time, convert to pgtype.Timestamptz later
	}{
		{
			ID:          uuid.New(),
			Title:       "Health Tips for Daily Life",
			Content:     "Content about health tips...",
			Category:    "health",
			Tags:        []string{"health", "wellness"},
			PublishedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Title:       "Diet Strategies for Weight Loss",
			Content:     "Content about diet strategies...",
			Category:    "nutrition",
			Tags:        []string{"diet", "nutrition", "health"},
			PublishedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Title:       "Exercise Routines for Beginners",
			Content:     "Content about exercise routines...",
			Category:    "fitness",
			Tags:        []string{"exercise", "fitness"},
			PublishedAt: time.Now().Add(-72 * time.Hour),
		},
		// Add an unpublished column for testing
		{
			ID:          uuid.New(),
			Title:       "Future Health Trends",
			Content:     "Content about future trends...",
			Category:    "trends",
			Tags:        []string{"future", "health"},
			PublishedAt: time.Now().Add(24 * time.Hour), // Published in the future
		},
	}

	// Insert mock columns using CreateTestColumn
	for _, data := range mockColumnsData {
		categoryVal := pgtype.Text{String: data.Category, Valid: data.Category != ""}
		publishedAtVal := pgtype.Timestamptz{Time: data.PublishedAt.UTC(), Valid: !data.PublishedAt.IsZero()} // Ensure UTC

		_, err := CreateTestColumn(ctx, pool, data.ID, data.Title, data.Content, categoryVal, data.Tags, publishedAtVal) // Capture error only
		if err != nil {
			// Log or handle the error for the specific column creation failure
			fmt.Printf("Warning: Failed to create mock column '%s' using CreateTestColumn: %v\n", data.Title, err)
		}
	}

	return nil // Indicate success even if some individual insertions failed
}
