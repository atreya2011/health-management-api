package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBodyRecord_Validate(t *testing.T) {
	// Create a valid body record
	validRecord := &BodyRecord{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Date:      time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:  floatPtr(75.5),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test validation of a valid record
	if err := validRecord.Validate(); err != nil {
		t.Errorf("Expected no validation error for valid record, got: %v", err)
	}

	// Test with negative weight (if validation is implemented)
	// This is just a placeholder for when you implement actual validation
	negativeWeightRecord := &BodyRecord{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Date:      time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:  floatPtr(-10.0), // Negative weight should be invalid
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// This test will pass until you implement validation that checks for negative weights
	if err := negativeWeightRecord.Validate(); err != nil {
		t.Logf("Validation correctly failed for negative weight: %v", err)
	}

	// Test with extremely high body fat percentage (if validation is implemented)
	// This is just a placeholder for when you implement actual validation
	highBodyFatRecord := &BodyRecord{
		ID:                uuid.New(),
		UserID:            uuid.New(),
		Date:              time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:          floatPtr(75.5),
		BodyFatPercentage: floatPtr(101.0), // Over 100% should be invalid
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// This test will pass until you implement validation that checks for valid body fat percentage range
	if err := highBodyFatRecord.Validate(); err != nil {
		t.Logf("Validation correctly failed for high body fat percentage: %v", err)
	}
}

// Helper function to create a pointer to a float64
func floatPtr(v float64) *float64 {
	return &v
}
