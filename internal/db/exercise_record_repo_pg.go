package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrExerciseRecordNotFound is returned when an exercise record is not found
var ErrExerciseRecordNotFound = errors.New("exercise record not found")

// PgExerciseRecordRepository provides database operations for ExerciseRecord
type PgExerciseRecordRepository struct {
	q *db.Queries
}

// NewPgExerciseRecordRepository creates a new PostgreSQL exercise record repository
func NewPgExerciseRecordRepository(pool *pgxpool.Pool) *PgExerciseRecordRepository { // Return exported type
	return &PgExerciseRecordRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new exercise record
func (r *PgExerciseRecordRepository) Create(ctx context.Context, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time) (db.ExerciseRecord, error) {
	var durationMinutesVal, caloriesBurnedVal pgtype.Int4

	if durationMinutes != nil {
		durationMinutesVal = pgtype.Int4{Int32: *durationMinutes, Valid: true}
	}

	if caloriesBurned != nil {
		caloriesBurnedVal = pgtype.Int4{Int32: *caloriesBurned, Valid: true}
	}

	recordedAtUTC := recordedAt.UTC() // Ensure UTC

	params := db.CreateExerciseRecordParams{
		UserID:          userID,
		ExerciseName:    exerciseName,
		DurationMinutes: durationMinutesVal,
		CaloriesBurned:  caloriesBurnedVal,
		RecordedAt:      recordedAtUTC,
	}

	dbRecord, err := r.q.CreateExerciseRecord(ctx, params)
	if err != nil {
		return db.ExerciseRecord{}, fmt.Errorf("failed to create exercise record: %w", err)
	}

	// Return generated struct directly
	return dbRecord, nil
}

// FindByUser retrieves paginated exercise records for a user
func (r *PgExerciseRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.ExerciseRecord, error) {
	params := db.ListExerciseRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListExerciseRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercise records: %w", err)
	}

	// Return generated structs directly
	return dbRecords, nil
}

// Delete deletes an exercise record by ID and user ID
func (r *PgExerciseRecordRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteExerciseRecordParams{
		ID:     id,
		UserID: userID,
	}

	err := r.q.DeleteExerciseRecord(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Consider returning ErrExerciseRecordNotFound here too for consistency?
			// For now, matching original behavior which returned nil on no rows for delete.
			return nil
		}
		return fmt.Errorf("failed to delete exercise record: %w", err) // Use fmt.Errorf
	}

	return nil
}

// CountByUser returns the total number of exercise records for a user
func (r *PgExerciseRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountExerciseRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count exercise records: %w", err)
	}

	return count, nil
}

// Removed toLocalExerciseRecord function as it's no longer needed
