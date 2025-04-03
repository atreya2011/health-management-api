package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ExerciseRecord represents an exercise record for a user
type ExerciseRecord struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ExerciseName    string
	DurationMinutes *int32
	CaloriesBurned  *int32
	RecordedAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Validate performs validation on the exercise record
func (er *ExerciseRecord) Validate() error {
	// Validate exercise name (required)
	if er.ExerciseName == "" {
		return errors.New("exercise name cannot be empty")
	}

	// Validate exercise name length
	if len(er.ExerciseName) > 100 {
		return errors.New("exercise name exceeds maximum allowed length (100 characters)")
	}

	// Validate duration minutes (if provided)
	if er.DurationMinutes != nil {
		duration := *er.DurationMinutes
		if duration <= 0 {
			return errors.New("duration must be positive")
		}
		if duration > 1440 { // 24 hours in minutes
			return errors.New("duration exceeds maximum allowed value (24 hours)")
		}
	}

	// Validate calories burned (if provided)
	if er.CaloriesBurned != nil {
		calories := *er.CaloriesBurned
		if calories < 0 {
			return errors.New("calories burned cannot be negative")
		}
		if calories > 10000 {
			return errors.New("calories burned exceeds maximum allowed value")
		}
	}

	// Validate recorded_at is not in the future
	now := time.Now()
	if er.RecordedAt.After(now) {
		return errors.New("recorded date cannot be in the future")
	}

	return nil
}
