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

	// Test with negative weight
	negativeWeightRecord := &BodyRecord{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Date:      time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:  floatPtr(-10.0), // Negative weight should be invalid
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// This should now fail with our validation
	if err := negativeWeightRecord.Validate(); err == nil {
		t.Error("Expected validation error for negative weight, got nil")
	} else {
		t.Logf("Validation correctly failed for negative weight: %v", err)
	}

	// Test with extremely high weight
	highWeightRecord := &BodyRecord{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Date:      time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:  floatPtr(600.0), // Over 500 should be invalid
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// This should now fail with our validation
	if err := highWeightRecord.Validate(); err == nil {
		t.Error("Expected validation error for high weight, got nil")
	} else {
		t.Logf("Validation correctly failed for high weight: %v", err)
	}

	// Test with extremely high body fat percentage
	highBodyFatRecord := &BodyRecord{
		ID:                uuid.New(),
		UserID:            uuid.New(),
		Date:              time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:          floatPtr(75.5),
		BodyFatPercentage: floatPtr(101.0), // Over 100% should be invalid
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// This should now fail with our validation
	if err := highBodyFatRecord.Validate(); err == nil {
		t.Error("Expected validation error for high body fat percentage, got nil")
	} else {
		t.Logf("Validation correctly failed for high body fat percentage: %v", err)
	}

	// Test with negative body fat percentage
	negativeBodyFatRecord := &BodyRecord{
		ID:                uuid.New(),
		UserID:            uuid.New(),
		Date:              time.Now().UTC().Truncate(24 * time.Hour),
		WeightKg:          floatPtr(75.5),
		BodyFatPercentage: floatPtr(-5.0), // Negative should be invalid
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// This should now fail with our validation
	if err := negativeBodyFatRecord.Validate(); err == nil {
		t.Error("Expected validation error for negative body fat percentage, got nil")
	} else {
		t.Logf("Validation correctly failed for negative body fat percentage: %v", err)
	}
}

// Helper function to create a pointer to a float64
func floatPtr(v float64) *float64 {
	return &v
}
