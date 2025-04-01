package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrDiaryEntryNotFound is returned when a diary entry is not found
var ErrDiaryEntryNotFound = errors.New("diary entry not found")

// DiaryEntryRepository defines the interface for diary entry data access
type DiaryEntryRepository interface {
	// Create creates a new diary entry
	Create(ctx context.Context, entry *DiaryEntry) (*DiaryEntry, error)
	
	// Update updates an existing diary entry
	Update(ctx context.Context, entry *DiaryEntry) (*DiaryEntry, error)
	
	// FindByID retrieves a diary entry by ID and user ID
	FindByID(ctx context.Context, id, userID uuid.UUID) (*DiaryEntry, error)
	
	// FindByUser retrieves paginated diary entries for a user
	FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*DiaryEntry, error)
	
	// Delete deletes a diary entry by ID and user ID
	Delete(ctx context.Context, id, userID uuid.UUID) error
	
	// CountByUser returns the total number of diary entries for a user
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
