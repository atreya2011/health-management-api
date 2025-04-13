package postgres

import (
	"context"
	"errors" // Use standard errors
	"fmt"
	"time" // Added for validation

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pkg/errors" // Removed, use fmt.Errorf with %w
)

// ErrExerciseRecordNotFound is returned when an exercise record is not found (Moved from domain)
var ErrExerciseRecordNotFound = errors.New("exercise record not found")

// ExerciseRecord represents an exercise record for a user (Moved from domain)
type ExerciseRecord struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ExerciseName    string
	DurationMinutes *int32
	CaloriesBurned  *int32
	RecordedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate performs validation on the exercise record (Moved from domain)
func (er *ExerciseRecord) Validate() error {
	// Validate exercise name (required)
	if er.ExerciseName == "" {
		return errors.New("exercise name cannot be empty")
	}

	// Validate exercise name length
	if len(er.ExerciseName) > 100 {
		return errors.New("exercise name exceeds maximum allowed length (100 characters)")
	}

	// Validate duration minutes (if provided)
	if er.DurationMinutes != nil {
		duration := *er.DurationMinutes
		if duration <= 0 {
			return errors.New("duration must be positive")
		}
		if duration > 1440 { // 24 hours in minutes
			return errors.New("duration exceeds maximum allowed value (24 hours)")
		}
	}

	// Validate calories burned (if provided)
	if er.CaloriesBurned != nil {
		calories := *er.CaloriesBurned
		if calories < 0 {
			return errors.New("calories burned cannot be negative")
		}
		if calories > 10000 {
			return errors.New("calories burned exceeds maximum allowed value")
		}
	}

	// Validate recorded_at is not in the future
	now := time.Now()
	if er.RecordedAt.After(now) {
		return errors.New("recorded date cannot be in the future")
	}

	return nil
}

// PgExerciseRecordRepository provides database operations for ExerciseRecord (Exported)
type PgExerciseRecordRepository struct { // Renamed to export
	q *db.Queries
}

// NewPgExerciseRecordRepository creates a new PostgreSQL exercise record repository
func NewPgExerciseRecordRepository(pool *pgxpool.Pool) *PgExerciseRecordRepository { // Return exported type
	return &PgExerciseRecordRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new exercise record
func (r *PgExerciseRecordRepository) Create(ctx context.Context, record *ExerciseRecord) (*ExerciseRecord, error) { // Use local ExerciseRecord
	var durationMinutesVal, caloriesBurnedVal pgtype.Int4

	if record.DurationMinutes != nil {
		durationMinutesVal = pgtype.Int4{Int32: *record.DurationMinutes, Valid: true}
	}

	if record.CaloriesBurned != nil {
		caloriesBurnedVal = pgtype.Int4{Int32: *record.CaloriesBurned, Valid: true}
	}

	recordedAtUTC := record.RecordedAt.UTC()

	params := db.CreateExerciseRecordParams{
		UserID:          record.UserID,
		ExerciseName:    record.ExerciseName,
		DurationMinutes: durationMinutesVal,
		CaloriesBurned:  caloriesBurnedVal,
		RecordedAt:      recordedAtUTC,
	}

	dbRecord, err := r.q.CreateExerciseRecord(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create exercise record: %w", err)
	}

	return toLocalExerciseRecord(dbRecord), nil // Use local conversion func
}

// FindByUser retrieves paginated exercise records for a user
func (r *PgExerciseRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ExerciseRecord, error) { // Use local ExerciseRecord slice
	params := db.ListExerciseRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListExerciseRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercise records: %w", err)
	}

	records := make([]*ExerciseRecord, len(dbRecords)) // Use local ExerciseRecord slice
	for i, dbRecord := range dbRecords {
		records[i] = toLocalExerciseRecord(dbRecord) // Use local conversion func
	}

	return records, nil
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

// toLocalExerciseRecord converts a db.ExerciseRecord (sqlc-generated) to a local ExerciseRecord
func toLocalExerciseRecord(dbRecord db.ExerciseRecord) *ExerciseRecord { // Return local ExerciseRecord
	var durationMinutes *int32
	if dbRecord.DurationMinutes.Valid {
		durationMinutes = &dbRecord.DurationMinutes.Int32
	}

	var caloriesBurned *int32
	if dbRecord.CaloriesBurned.Valid {
		caloriesBurned = &dbRecord.CaloriesBurned.Int32
	}

	return &ExerciseRecord{ // Use local ExerciseRecord
		ID:              dbRecord.ID,
		UserID:          dbRecord.UserID,
		ExerciseName:    dbRecord.ExerciseName,
		DurationMinutes: durationMinutes,
		CaloriesBurned:  caloriesBurned,
		RecordedAt:      dbRecord.RecordedAt,
		CreatedAt:       dbRecord.CreatedAt,
		UpdatedAt:       dbRecord.UpdatedAt,
	}
}
