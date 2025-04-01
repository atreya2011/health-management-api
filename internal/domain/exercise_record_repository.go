package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrExerciseRecordNotFound is returned when an exercise record is not found
var ErrExerciseRecordNotFound = errors.New("exercise record not found")

// ExerciseRecordRepository defines the interface for exercise record data access
type ExerciseRecordRepository interface {
	// Create creates a new exercise record
	Create(ctx context.Context, record *ExerciseRecord) (*ExerciseRecord, error)
	
	// FindByUser retrieves paginated exercise records for a user
	FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ExerciseRecord, error)
	
	// Delete deletes an exercise record by ID and user ID
	Delete(ctx context.Context, id, userID uuid.UUID) error
	
	// CountByUser returns the total number of exercise records for a user
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
