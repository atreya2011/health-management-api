package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrBodyRecordNotFound is returned when a body record is not found
var ErrBodyRecordNotFound = errors.New("body record not found")

// BodyRecordRepository defines the interface for body record data access
type BodyRecordRepository interface {
	// Save creates a new body record or updates an existing one based on UserID and Date
	Save(ctx context.Context, record *BodyRecord) (*BodyRecord, error)
	
	// FindByUser retrieves paginated body records for a user
	FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*BodyRecord, error)
	
	// FindByUserAndDateRange retrieves body records for a user within a specific date range
	FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*BodyRecord, error)
	
	// CountByUser returns the total number of body records for a user
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
