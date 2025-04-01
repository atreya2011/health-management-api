package domain

import (
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
	// Add validation rules here, e.g., non-empty exercise name
	return nil
}
