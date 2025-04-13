package postgres

import (
	"context"
	// "errors" // Removed unused import
	"fmt"
	"time" // Keep for input parameters

	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pkg/errors" // Removed, use fmt.Errorf with %w
)

// PgBodyRecordRepository provides database operations for BodyRecord
type PgBodyRecordRepository struct {
	q *db.Queries
}

// NewPgBodyRecordRepository creates a new PostgreSQL body record repository
func NewPgBodyRecordRepository(pool *pgxpool.Pool) *PgBodyRecordRepository { // Return exported type
	return &PgBodyRecordRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Save creates a new body record or updates an existing one based on UserID and Date
func (r *PgBodyRecordRepository) Save(ctx context.Context, userID uuid.UUID, date time.Time, weightKg *float64, bodyFatPercentage *float64) (db.BodyRecord, error) {
	var weightVal, bodyFatVal pgtype.Numeric

	// Convert *float64 to pgtype.Numeric by scanning from string
	if weightKg != nil {
		weightStr := fmt.Sprintf("%f", *weightKg)
		if err := weightVal.Scan(weightStr); err != nil {
			return db.BodyRecord{}, fmt.Errorf("failed to scan weight string '%s' into pgtype.Numeric: %w", weightStr, err)
		}
	} else {
		weightVal = pgtype.Numeric{Valid: false}
	}

	if bodyFatPercentage != nil {
		bodyFatStr := fmt.Sprintf("%f", *bodyFatPercentage)
		if err := bodyFatVal.Scan(bodyFatStr); err != nil {
			return db.BodyRecord{}, fmt.Errorf("failed to scan bodyFat string '%s' into pgtype.Numeric: %w", bodyFatStr, err)
		}
	} else {
		bodyFatVal = pgtype.Numeric{Valid: false}
	}

	pgDate := pgtype.Date{Time: date, Valid: true}

	params := db.CreateBodyRecordParams{
		UserID:            userID,
		Date:              pgDate,
		WeightKg:          weightVal,
		BodyFatPercentage: bodyFatVal,
	}

	dbRecord, err := r.q.CreateBodyRecord(ctx, params)
	if err != nil {
		// Return zero value of db.BodyRecord on error
		return db.BodyRecord{}, fmt.Errorf("failed to save body record: %w", err)
	}

	// Return generated struct directly
	return dbRecord, nil
}

// FindByUser retrieves paginated body records for a user
func (r *PgBodyRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.BodyRecord, error) {
	params := db.ListBodyRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListBodyRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list body records: %w", err)
	}

	// Return generated structs directly
	return dbRecords, nil
}

// FindByUserAndDateRange retrieves body records for a user within a specific date range
func (r *PgBodyRecordRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]db.BodyRecord, error) {
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

	// Return generated structs directly
	return dbRecords, nil
}

// CountByUser returns the total number of body records for a user
func (r *PgBodyRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountBodyRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count body records: %w", err)
	}

	return count, nil
}

// Removed toLocalBodyRecord function as it's no longer needed
