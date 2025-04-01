package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrColumnNotFound is returned when a column is not found
var ErrColumnNotFound = errors.New("column not found")

// ColumnRepository defines the interface for column data access
type ColumnRepository interface {
	// FindPublished retrieves paginated published columns
	FindPublished(ctx context.Context, limit, offset int) ([]*Column, error)
	
	// FindByID retrieves a column by ID
	FindByID(ctx context.Context, id uuid.UUID) (*Column, error)
	
	// FindByCategory retrieves paginated columns by category
	FindByCategory(ctx context.Context, category string, limit, offset int) ([]*Column, error)
	
	// FindByTag retrieves paginated columns by tag
	FindByTag(ctx context.Context, tag string, limit, offset int) ([]*Column, error)
	
	// CountPublished returns the total number of published columns
	CountPublished(ctx context.Context) (int64, error)
	
	// CountByCategory returns the total number of published columns in a category
	CountByCategory(ctx context.Context, category string) (int64, error)
	
	// CountByTag returns the total number of published columns with a tag
	CountByTag(ctx context.Context, tag string) (int64, error)
}
