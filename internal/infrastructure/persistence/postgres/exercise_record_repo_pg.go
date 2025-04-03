package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgExerciseRecordRepository implements the domain.ExerciseRecordRepository interface
type pgExerciseRecordRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPgExerciseRecordRepository creates a new PostgreSQL exercise record repository
func NewPgExerciseRecordRepository(pool *pgxpool.Pool) domain.ExerciseRecordRepository {
	adapter := NewPgxAdapter(pool)
	return &pgExerciseRecordRepository{
		pool: pool,
		q:    db.New(adapter),
	}
}

// Create creates a new exercise record
func (r *pgExerciseRecordRepository) Create(ctx context.Context, record *domain.ExerciseRecord) (*domain.ExerciseRecord, error) {
	// Convert *int32 to sql.NullInt32
	var durationMinutes sql.NullInt32
	if record.DurationMinutes != nil {
		durationMinutes = sql.NullInt32{
			Int32: *record.DurationMinutes,
			Valid: true,
		}
	}

	var caloriesBurned sql.NullInt32
	if record.CaloriesBurned != nil {
		caloriesBurned = sql.NullInt32{
			Int32: *record.CaloriesBurned,
			Valid: true,
		}
	}

	params := db.CreateExerciseRecordParams{
		UserID:          record.UserID,
		ExerciseName:    record.ExerciseName,
		DurationMinutes: durationMinutes,
		CaloriesBurned:  caloriesBurned,
		RecordedAt:      record.RecordedAt,
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
		if err == sql.ErrNoRows {
			return domain.ErrExerciseRecordNotFound
		}
		return fmt.Errorf("failed to delete exercise record: %w", err)
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

// toDomainExerciseRecord converts a db.ExerciseRecord to a domain.ExerciseRecord
func toDomainExerciseRecord(dbRecord db.ExerciseRecord) *domain.ExerciseRecord {
	// Convert sql.NullInt32 to *int32
	var durationMinutes *int32
	if dbRecord.DurationMinutes.Valid {
		d := dbRecord.DurationMinutes.Int32
		durationMinutes = &d
	}

	var caloriesBurned *int32
	if dbRecord.CaloriesBurned.Valid {
		c := dbRecord.CaloriesBurned.Int32
		caloriesBurned = &c
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
