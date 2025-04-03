package domain

import (
	"errors"
	"strings"
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
	// Validate content (required)
	if strings.TrimSpace(de.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// Validate content length
	if len(de.Content) > 10000 {
		return errors.New("content exceeds maximum allowed length (10000 characters)")
	}

	// Validate title length (if provided)
	if de.Title != nil && len(*de.Title) > 200 {
		return errors.New("title exceeds maximum allowed length (200 characters)")
	}

	// Validate entry date is not in the future
	now := time.Now()
	if de.EntryDate.After(now) {
		return errors.New("entry date cannot be in the future")
	}

	return nil
}
