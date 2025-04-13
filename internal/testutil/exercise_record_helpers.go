package testutil

import (
	"context"
	"fmt"
	"time"

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	postgres "github.com/atreya2011/health-management-api/internal/db"
	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestExerciseRecord creates a test exercise record in the database using sqlc
// Takes Queries directly.
func CreateTestExerciseRecord(ctx context.Context, queries *db.Queries, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time) (*postgres.ExerciseRecord, error) { // Return postgres.ExerciseRecord
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
	}

	dbRecord, err := queries.CreateExerciseRecord(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("could not create test exercise record: %w", err)
	}

	var durationMinutesPtr *int32
	var caloriesBurnedPtr *int32

	if dbRecord.DurationMinutes.Valid {
		durationMinutesPtr = &dbRecord.DurationMinutes.Int32
	}

	if dbRecord.CaloriesBurned.Valid {
		caloriesBurnedPtr = &dbRecord.CaloriesBurned.Int32
	}

	recordedAtVal := dbRecord.RecordedAt // No conversion needed

	return &postgres.ExerciseRecord{ // Return postgres.ExerciseRecord
		ID:              dbRecord.ID,
		UserID:          dbRecord.UserID,
		ExerciseName:    dbRecord.ExerciseName,
		DurationMinutes: durationMinutesPtr,
		CaloriesBurned:  caloriesBurnedPtr,
		RecordedAt:      recordedAtVal,      // Use the time.Time value
		CreatedAt:       dbRecord.CreatedAt, // Already time.Time
		UpdatedAt:       dbRecord.UpdatedAt, // Already time.Time
	}, nil
}

// NewExerciseRecordRepository creates a new exercise record repository for testing
func NewExerciseRecordRepository(pool *pgxpool.Pool) *postgres.PgExerciseRecordRepository { // Return concrete type
	return postgres.NewPgExerciseRecordRepository(pool)
}
