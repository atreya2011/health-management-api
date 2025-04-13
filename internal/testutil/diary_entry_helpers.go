package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/repo"
	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestDiaryEntry creates a test diary entry in the database using sqlc
// Takes Queries directly.
func CreateTestDiaryEntry(ctx context.Context, queries *db.Queries, userID uuid.UUID, title string, content string, entryDate time.Time) (uuid.UUID, error) {
	var titleVal pgtype.Text // Use pgtype.Text

	if title != "" {
		titleVal = pgtype.Text{String: title, Valid: true}
	}

	// Convert time.Time to pgtype.Date
	pgDate := pgtype.Date{Time: entryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    userID,
		Title:     titleVal,
		Content:   content,
		EntryDate: pgDate, // Use pgtype.Date
	}

	dbEntry, err := queries.CreateDiaryEntry(ctx, params)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not create test diary entry: %w", err)
	}

	return dbEntry.ID, nil
}

// NewDiaryEntryRepository creates a new diary entry repository for testing
func NewDiaryEntryRepository(pool *pgxpool.Pool) *repo.DiaryEntryRepository { // Return concrete type
	return repo.NewDiaryEntryRepository(pool)
}
