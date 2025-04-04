package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// pgBodyRecordRepository implements the domain.BodyRecordRepository interface
type pgBodyRecordRepository struct {
	q *db.Queries
}

// NewPgBodyRecordRepository creates a new PostgreSQL body record repository
func NewPgBodyRecordRepository(pool *pgxpool.Pool) domain.BodyRecordRepository {
	return &pgBodyRecordRepository{
		q: db.New(pool),
	}
}

// Save creates a new body record or updates an existing one based on UserID and Date
func (r *pgBodyRecordRepository) Save(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
	var weightVal, bodyFatVal pgtype.Numeric

	// Convert *float64 to pgtype.Numeric by scanning from string
	if record.WeightKg != nil {
		weightStr := fmt.Sprintf("%f", *record.WeightKg)
		if err := weightVal.Scan(weightStr); err != nil {
			return nil, errors.Wrapf(err, "failed to scan weight string '%s' into pgtype.Numeric", weightStr)
		}
	} else {
		weightVal = pgtype.Numeric{Valid: false}
	}

	if record.BodyFatPercentage != nil {
		bodyFatStr := fmt.Sprintf("%f", *record.BodyFatPercentage)
		if err := bodyFatVal.Scan(bodyFatStr); err != nil {
			return nil, errors.Wrapf(err, "failed to scan bodyFat string '%s' into pgtype.Numeric", bodyFatStr)
		}
	} else {
		bodyFatVal = pgtype.Numeric{Valid: false}
	}

	pgDate := pgtype.Date{Time: record.Date, Valid: true}

	params := db.CreateBodyRecordParams{
		UserID:            record.UserID,
		Date:              pgDate,
		WeightKg:          weightVal,
		BodyFatPercentage: bodyFatVal,
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
	pgStartDate := pgtype.Date{Time: startDate, Valid: true}
	pgEndDate := pgtype.Date{Time: endDate, Valid: true}

	params := db.ListBodyRecordsByUserDateRangeParams{
		UserID: userID,
		Date:   pgStartDate,
		Date_2: pgEndDate,
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

// toDomainBodyRecord converts a db.BodyRecord (pgx-based) to a domain.BodyRecord
func toDomainBodyRecord(dbRecord db.BodyRecord) *domain.BodyRecord {
	var weightKg *float64
	var bodyFatPercentage *float64

	if dbRecord.WeightKg.Valid {
		w, err := dbRecord.WeightKg.Float64Value()
		if err == nil {
			weightKg = &w.Float64
		} else {
			fmt.Printf("Warning: could not scan WeightKg back to float64: %v\n", err)
		}
	}

	if dbRecord.BodyFatPercentage.Valid {
		bf, err := dbRecord.BodyFatPercentage.Float64Value()
		if err == nil {
			bodyFatPercentage = &bf.Float64
		} else {
			fmt.Printf("Warning: could not scan BodyFatPercentage back to float64: %v\n", err)
		}
	}

	var dateVal time.Time
	if dbRecord.Date.Valid {
		dateVal = dbRecord.Date.Time
	}

	return &domain.BodyRecord{
		ID:                dbRecord.ID,
		UserID:            dbRecord.UserID,
		Date:              dateVal,
		WeightKg:          weightKg,
		BodyFatPercentage: bodyFatPercentage,
		CreatedAt:         dbRecord.CreatedAt,
		UpdatedAt:         dbRecord.UpdatedAt,
	}
}
