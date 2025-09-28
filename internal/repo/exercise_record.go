package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrExerciseRecordNotFound is returned when an exercise record is not found
var ErrExerciseRecordNotFound = errors.New("exercise record not found")

// ExerciseRecordRepository provides database operations for ExerciseRecord
type ExerciseRecordRepository struct {
	q *db.Queries
}

// NewExerciseRecordRepository creates a new PostgreSQL exercise record repository
func NewExerciseRecordRepository(pool *pgxpool.Pool) *ExerciseRecordRepository { // Return exported type
	return &ExerciseRecordRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new exercise record, accepting the current time.
func (r *ExerciseRecordRepository) Create(ctx context.Context, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time, now time.Time) (db.ExerciseRecord, error) {
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
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	dbRecord, err := r.q.CreateExerciseRecord(ctx, params)
	if err != nil {
		return db.ExerciseRecord{}, fmt.Errorf("failed to create exercise record: %w", err)
	}

	// Return generated struct directly
	return dbRecord, nil
}

// FindByUser retrieves paginated exercise records for a user
func (r *ExerciseRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.ExerciseRecord, error) {
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
func (r *ExerciseRecordRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteExerciseRecordParams{
		ID:     id,
		UserID: userID,
	}

	err := r.q.DeleteExerciseRecord(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no rows were deleted (record not found or doesn't belong to user), return specific error.
			return ErrExerciseRecordNotFound
		}
		return fmt.Errorf("failed to delete exercise record: %w", err)
	}

	return nil
}

// CountByUser returns the total number of exercise records for a user
func (r *ExerciseRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountExerciseRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count exercise records: %w", err)
	}

	return count, nil
}
