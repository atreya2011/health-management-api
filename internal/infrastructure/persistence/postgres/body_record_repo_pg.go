package postgres

import (
	"context"
	"errors" // Use standard errors package
	"fmt"
	"time"

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pkg/errors" // Removed, use fmt.Errorf with %w
)

// BodyRecord represents a body composition record for a user (Moved from domain)
type BodyRecord struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Date              time.Time // Store as time.Time (YYYY-MM-DD 00:00:00 UTC)
	WeightKg          *float64
	BodyFatPercentage *float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Validate performs validation on the body record (Moved from domain)
func (br *BodyRecord) Validate() error {
	// Validate weight (if provided)
	if br.WeightKg != nil {
		weight := *br.WeightKg
		if weight <= 0 {
			return errors.New("weight must be positive")
		}
		if weight > 500 {
			return errors.New("weight exceeds maximum allowed value")
		}
	}

	// Validate body fat percentage (if provided)
	if br.BodyFatPercentage != nil {
		bodyFat := *br.BodyFatPercentage
		if bodyFat < 0 {
			return errors.New("body fat percentage cannot be negative")
		}
		if bodyFat > 100 {
			return errors.New("body fat percentage cannot exceed 100%")
		}
	}

	return nil
}

// PgBodyRecordRepository provides database operations for BodyRecord (Exported)
type PgBodyRecordRepository struct { // Renamed to export
	q *db.Queries
}

// NewPgBodyRecordRepository creates a new PostgreSQL body record repository
func NewPgBodyRecordRepository(pool *pgxpool.Pool) *PgBodyRecordRepository { // Return exported type
	return &PgBodyRecordRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Save creates a new body record or updates an existing one based on UserID and Date
func (r *PgBodyRecordRepository) Save(ctx context.Context, record *BodyRecord) (*BodyRecord, error) { // Use local BodyRecord
	var weightVal, bodyFatVal pgtype.Numeric

	// Convert *float64 to pgtype.Numeric by scanning from string
	// Note: Validation should happen before calling Save, moved to handler
	if record.WeightKg != nil {
		weightStr := fmt.Sprintf("%f", *record.WeightKg)
		if err := weightVal.Scan(weightStr); err != nil {
			// Use fmt.Errorf with %w
			return nil, fmt.Errorf("failed to scan weight string '%s' into pgtype.Numeric: %w", weightStr, err)
		}
	} else {
		weightVal = pgtype.Numeric{Valid: false}
	}

	if record.BodyFatPercentage != nil {
		bodyFatStr := fmt.Sprintf("%f", *record.BodyFatPercentage)
		if err := bodyFatVal.Scan(bodyFatStr); err != nil {
			// Use fmt.Errorf with %w
			return nil, fmt.Errorf("failed to scan bodyFat string '%s' into pgtype.Numeric: %w", bodyFatStr, err)
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

	return toLocalBodyRecord(dbRecord), nil // Use local conversion func
}

// FindByUser retrieves paginated body records for a user
func (r *PgBodyRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*BodyRecord, error) { // Use local BodyRecord slice
	params := db.ListBodyRecordsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbRecords, err := r.q.ListBodyRecordsByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list body records: %w", err)
	}

	records := make([]*BodyRecord, len(dbRecords)) // Use local BodyRecord slice
	for i, dbRecord := range dbRecords {
		records[i] = toLocalBodyRecord(dbRecord) // Use local conversion func
	}

	return records, nil
}

// FindByUserAndDateRange retrieves body records for a user within a specific date range
func (r *PgBodyRecordRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*BodyRecord, error) { // Use local BodyRecord slice
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

	records := make([]*BodyRecord, len(dbRecords)) // Use local BodyRecord slice
	for i, dbRecord := range dbRecords {
		records[i] = toLocalBodyRecord(dbRecord) // Use local conversion func
	}

	return records, nil
}

// CountByUser returns the total number of body records for a user
func (r *PgBodyRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountBodyRecordsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count body records: %w", err)
	}

	return count, nil
}

// toLocalBodyRecord converts a db.BodyRecord (sqlc-generated) to a local BodyRecord
func toLocalBodyRecord(dbRecord db.BodyRecord) *BodyRecord { // Return local BodyRecord
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

	return &BodyRecord{ // Use local BodyRecord
		ID:                dbRecord.ID,
		UserID:            dbRecord.UserID,
		Date:              dateVal,
		WeightKg:          weightKg,
		BodyFatPercentage: bodyFatPercentage,
		CreatedAt:         dbRecord.CreatedAt,
		UpdatedAt:         dbRecord.UpdatedAt,
	}
}
