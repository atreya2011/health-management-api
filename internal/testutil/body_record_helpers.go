package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTestBodyRecord creates a test body record in the database using sqlc
// Takes Queries directly.
func CreateTestBodyRecord(ctx context.Context, queries *db.Queries, userID uuid.UUID, date time.Time, weight *float64, bodyFat *float64) (*domain.BodyRecord, error) {
	var weightVal, bodyFatVal pgtype.Numeric // Use pgtype.Numeric

	if weight != nil {
		// Scan from a string representation of the float
		weightStr := fmt.Sprintf("%f", *weight)
		if err := weightVal.Scan(weightStr); err != nil {
			return nil, fmt.Errorf("failed to scan weight string '%s' into pgtype.Numeric: %w", weightStr, err)
		}
	}

	if bodyFat != nil {
		// Scan from a string representation of the float
		bodyFatStr := fmt.Sprintf("%f", *bodyFat)
		if err := bodyFatVal.Scan(bodyFatStr); err != nil {
			return nil, fmt.Errorf("failed to scan bodyFat string '%s' into pgtype.Numeric: %w", bodyFatStr, err)
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
		return nil, fmt.Errorf("could not create test body record: %w", err)
	}

	var weightKg *float64
	var bodyFatPercentage *float64

	if dbRecord.WeightKg.Valid {
		var w float64
		if err := dbRecord.WeightKg.Scan(&w); err == nil {
			weightKg = &w
		} else {
			fmt.Printf("Warning: could not scan WeightKg back to float64: %v\n", err)
		}
	}

	if dbRecord.BodyFatPercentage.Valid {
		var bf float64
		if err := dbRecord.BodyFatPercentage.Scan(&bf); err == nil {
			bodyFatPercentage = &bf
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
	}, nil
}

// NewBodyRecordRepository creates a new body record repository for testing
func NewBodyRecordRepository(pool *pgxpool.Pool) domain.BodyRecordRepository {
	return postgres.NewPgBodyRecordRepository(pool)
}
