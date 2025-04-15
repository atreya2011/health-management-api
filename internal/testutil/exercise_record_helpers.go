package testutil

import (
	"context"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen" // Import proto types
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateTestExerciseRecord creates a test exercise record in the database using sqlc
// Takes Queries directly and returns the generated db.ExerciseRecord.
// Accepts the current time for timestamps.
func CreateTestExerciseRecord(ctx context.Context, queries *db.Queries, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time, now time.Time) (db.ExerciseRecord, error) {
	var durationMinutesVal, caloriesBurnedVal pgtype.Int4 // Use pgtype.Int4 for INTEGER

	if durationMinutes != nil {
		durationMinutesVal = pgtype.Int4{Int32: *durationMinutes, Valid: true}
	}

	if caloriesBurned != nil {
		caloriesBurnedVal = pgtype.Int4{Int32: *caloriesBurned, Valid: true}
	}

	// Ensure the time is UTC before passing.
	recordedAtUTC := recordedAt.UTC()

	params := db.CreateExerciseRecordParams{
		UserID:          userID,
		ExerciseName:    exerciseName,
		DurationMinutes: durationMinutesVal,
		CaloriesBurned:  caloriesBurnedVal,
		RecordedAt:      recordedAtUTC, // Pass time.Time directly (ensure it's UTC)
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	dbRecord, err := queries.CreateExerciseRecord(ctx, params)
	if err != nil {
		return db.ExerciseRecord{}, fmt.Errorf("could not create test exercise record: %w", err) // Return zero value on error
	}

	// Return the generated struct directly
	return dbRecord, nil
}
