package postgres

import (
	"context"
	"fmt"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// pgExerciseRecordRepository implements the domain.ExerciseRecordRepository interface
type pgExerciseRecordRepository struct {
	q *db.Queries
}

// NewPgExerciseRecordRepository creates a new PostgreSQL exercise record repository
func NewPgExerciseRecordRepository(pool *pgxpool.Pool) domain.ExerciseRecordRepository {
	return &pgExerciseRecordRepository{
		q: db.New(pool),
	}
}

// Create creates a new exercise record
func (r *pgExerciseRecordRepository) Create(ctx context.Context, record *domain.ExerciseRecord) (*domain.ExerciseRecord, error) {
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

	return toDomainExerciseRecord(dbRecord), nil
}

// FindByUser retrieves paginated exercise records for a user
func (r *pgExerciseRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.ExerciseRecord, error) {
	params := db.ListExerciseRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListExerciseRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercise records: %w", err)
	}

	records := make([]*domain.ExerciseRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = toDomainExerciseRecord(dbRecord)
	}

	return records, nil
}

// Delete deletes an exercise record by ID and user ID
func (r *pgExerciseRecordRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteExerciseRecordParams{
		ID:     id,
		UserID: userID,
	}

	err := r.q.DeleteExerciseRecord(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return errors.Wrap(err, "failed to delete exercise record")
	}

	return nil
}

// CountByUser returns the total number of exercise records for a user
func (r *pgExerciseRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountExerciseRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count exercise records: %w", err)
	}

	return count, nil
}

// toDomainExerciseRecord converts a db.ExerciseRecord (pgx-based) to a domain.ExerciseRecord
func toDomainExerciseRecord(dbRecord db.ExerciseRecord) *domain.ExerciseRecord {
	var durationMinutes *int32
	if dbRecord.DurationMinutes.Valid {
		durationMinutes = &dbRecord.DurationMinutes.Int32
	}

	var caloriesBurned *int32
	if dbRecord.CaloriesBurned.Valid {
		caloriesBurned = &dbRecord.CaloriesBurned.Int32
	}

	return &domain.ExerciseRecord{
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
