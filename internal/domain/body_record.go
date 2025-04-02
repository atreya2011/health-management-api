package domain

import (
	"errors"
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
