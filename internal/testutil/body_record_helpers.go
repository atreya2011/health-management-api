package testutil

import (
	"context"
	"fmt"
	"time"

	postgres "github.com/atreya2011/health-management-api/internal/db"
	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestBodyRecord creates a test body record in the database using sqlc
// Takes Queries directly and returns the generated db.BodyRecord.
func CreateTestBodyRecord(ctx context.Context, queries *db.Queries, userID uuid.UUID, date time.Time, weight *float64, bodyFat *float64) (db.BodyRecord, error) { // Return db.BodyRecord
	var weightVal, bodyFatVal pgtype.Numeric // Use pgtype.Numeric

	if weight != nil {
		// Scan from a string representation of the float
		weightStr := fmt.Sprintf("%f", *weight)
		if err := weightVal.Scan(weightStr); err != nil {
			return db.BodyRecord{}, fmt.Errorf("failed to scan weight string '%s' into pgtype.Numeric: %w", weightStr, err) // Return zero value
		}
	}

	if bodyFat != nil {
		// Scan from a string representation of the float
		bodyFatStr := fmt.Sprintf("%f", *bodyFat)
		if err := bodyFatVal.Scan(bodyFatStr); err != nil {
			return db.BodyRecord{}, fmt.Errorf("failed to scan bodyFat string '%s' into pgtype.Numeric: %w", bodyFatStr, err) // Return zero value
		}
	}

	pgDate := pgtype.Date{Time: date, Valid: true}

	params := db.CreateBodyRecordParams{
		UserID:            userID,
		Date:              pgDate,
		WeightKg:          weightVal,
		BodyFatPercentage: bodyFatVal,
	}

	dbRecord, err := queries.CreateBodyRecord(ctx, params)
	if err != nil {
		return db.BodyRecord{}, fmt.Errorf("could not create test body record: %w", err) // Return zero value on error
	}

	// Return the generated struct directly
	return dbRecord, nil
}

// NewBodyRecordRepository creates a new body record repository for testing
func NewBodyRecordRepository(pool *pgxpool.Pool) *postgres.BodyRecordRepository { // Return concrete type
	return postgres.NewBodyRecordRepository(pool)
}
