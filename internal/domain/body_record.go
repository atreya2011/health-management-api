package domain

import (
	"time"

	"github.com/google/uuid"
)

// BodyRecord represents a body composition record for a user
type BodyRecord struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Date              time.Time // Store as time.Time (YYYY-MM-DD 00:00:00 UTC)
	WeightKg          *float64
	BodyFatPercentage *float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Validate performs validation on the body record
func (br *BodyRecord) Validate() error {
	// Add validation rules here, e.g., range checks for weight/percentage
	return nil
}
