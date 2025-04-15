package testutil

import (
	"context"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateTestDiaryEntry creates a test diary entry in the database using sqlc
// Takes Queries directly and returns the created entry.
func CreateTestDiaryEntry(ctx context.Context, queries *db.Queries, userID uuid.UUID, title string, content string, entryDate time.Time) (db.DiaryEntry, error) { // Return db.DiaryEntry
	var titleVal pgtype.Text // Use pgtype.Text

	if title != "" {
		titleVal = pgtype.Text{String: title, Valid: true}
	}

	pgDate := pgtype.Date{Time: entryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    userID,
		Title:     titleVal,
		Content:   content,
		EntryDate: pgDate, // Use pgtype.Date
	}

	dbEntry, err := queries.CreateDiaryEntry(ctx, params)
	if err != nil {
		return db.DiaryEntry{}, fmt.Errorf("could not create test diary entry: %w", err)
	}

	// No need to fetch again, CreateDiaryEntry returns the full struct
	return dbEntry, nil
}
