package domain

import (
	"time"

	"github.com/google/uuid"
)

// DiaryEntry represents a diary entry for a user
type DiaryEntry struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     *string
	Content   string
	EntryDate time.Time // Store as time.Time (YYYY-MM-DD 00:00:00 UTC)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate performs validation on the diary entry
func (de *DiaryEntry) Validate() error {
	// Add validation rules here, e.g., non-empty content
	return nil
}
