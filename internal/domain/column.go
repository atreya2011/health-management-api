package domain

import (
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
	// Add validation rules here, e.g., non-empty title and content
	return nil
}

// IsPublished checks if the column is published and visible
func (c *Column) IsPublished() bool {
	return c.PublishedAt != nil && c.PublishedAt.Before(time.Now())
}
