package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgBodyRecordRepository implements the domain.BodyRecordRepository interface
type pgBodyRecordRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewPgBodyRecordRepository creates a new PostgreSQL body record repository
func NewPgBodyRecordRepository(pool *pgxpool.Pool) domain.BodyRecordRepository {
	adapter := NewPgxAdapter(pool)
	return &pgBodyRecordRepository{
		pool: pool,
		q:    db.New(adapter),
	}
}

// Save creates a new body record or updates an existing one based on UserID and Date
func (r *pgBodyRecordRepository) Save(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
	// Convert *float64 to sql.NullString
	var weightKg sql.NullString
	var bodyFatPercentage sql.NullString
	
	if record.WeightKg != nil {
		weightKg = sql.NullString{
			String: fmt.Sprintf("%.2f", *record.WeightKg),
			Valid:  true,
		}
	}
	
	if record.BodyFatPercentage != nil {
		bodyFatPercentage = sql.NullString{
			String: fmt.Sprintf("%.2f", *record.BodyFatPercentage),
			Valid:  true,
		}
	}
	
	params := db.CreateBodyRecordParams{
		UserID:            record.UserID,
		Date:              record.Date,
		WeightKg:          weightKg,
		BodyFatPercentage: bodyFatPercentage,
	}

	dbRecord, err := r.q.CreateBodyRecord(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to save body record: %w", err)
	}

	return toDomainBodyRecord(dbRecord), nil
}

// FindByUser retrieves paginated body records for a user
func (r *pgBodyRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.BodyRecord, error) {
	params := db.ListBodyRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListBodyRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list body records: %w", err)
	}

	records := make([]*domain.BodyRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = toDomainBodyRecord(dbRecord)
	}

	return records, nil
}

// FindByUserAndDateRange retrieves body records for a user within a specific date range
func (r *pgBodyRecordRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.BodyRecord, error) {
	params := db.ListBodyRecordsByUserDateRangeParams{
		UserID: userID,
		Date:   startDate,
		Date_2: endDate,
	}

	dbRecords, err := r.q.ListBodyRecordsByUserDateRange(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list body records by date range: %w", err)
	}

	records := make([]*domain.BodyRecord, len(dbRecords))
	for i, dbRecord := range dbRecords {
		records[i] = toDomainBodyRecord(dbRecord)
	}

	return records, nil
}

// CountByUser returns the total number of body records for a user
func (r *pgBodyRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountBodyRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count body records: %w", err)
	}

	return count, nil
}

// toDomainBodyRecord converts a db.BodyRecord to a domain.BodyRecord
func toDomainBodyRecord(dbRecord db.BodyRecord) *domain.BodyRecord {
	// Convert sql.NullString to *float64
	var weightKg *float64
	var bodyFatPercentage *float64
	
	if dbRecord.WeightKg.Valid {
		val, err := strconv.ParseFloat(dbRecord.WeightKg.String, 64)
		if err == nil {
			weightKg = &val
		}
	}
	
	if dbRecord.BodyFatPercentage.Valid {
		val, err := strconv.ParseFloat(dbRecord.BodyFatPercentage.String, 64)
		if err == nil {
			bodyFatPercentage = &val
		}
	}
	
	return &domain.BodyRecord{
		ID:                dbRecord.ID,
		UserID:            dbRecord.UserID,
		Date:              dbRecord.Date,
		WeightKg:          weightKg,
		BodyFatPercentage: bodyFatPercentage,
		CreatedAt:         dbRecord.CreatedAt,
		UpdatedAt:         dbRecord.UpdatedAt,
	}
}
