package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Column represents a health-related article or column
type Column struct {
	ID          uuid.UUID
	Title       string
	Content     string
	Category    *string
	Tags        []string
	PublishedAt *time.Time // Nullable, only show if not null and in the past
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate performs validation on the column
func (c *Column) Validate() error {
	if c.Title == "" {
		return errors.New("title cannot be empty")
	}
	if c.Content == "" {
		return errors.New("content cannot be empty")
	}
	return nil
}

// IsPublished checks if the column is published and visible
func (c *Column) IsPublished() bool {
	return c.PublishedAt != nil && c.PublishedAt.Before(time.Now())
}
