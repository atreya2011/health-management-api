package postgres

import (
	"context"
	"errors" // Use standard errors
	"fmt"
	"time"

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pkg/errors" // Removed, use fmt.Errorf with %w
)

// ErrColumnNotFound is returned when a column is not found (Moved from domain)
var ErrColumnNotFound = errors.New("column not found")

// Column represents a health-related article or column (Moved from domain)
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

// Validate performs validation on the column (Moved from domain)
func (c *Column) Validate() error {
	if c.Title == "" {
		return errors.New("title cannot be empty")
	}
	if c.Content == "" {
		return errors.New("content cannot be empty")
	}
	return nil
}

// IsPublished checks if the column is published and visible (Moved from domain)
func (c *Column) IsPublished() bool {
	return c.PublishedAt != nil && c.PublishedAt.Before(time.Now())
}

// PgColumnRepository provides database operations for Column (Exported)
type PgColumnRepository struct { // Renamed to export
	q *db.Queries
}

// NewPgColumnRepository creates a new PostgreSQL column repository
func NewPgColumnRepository(pool *pgxpool.Pool) *PgColumnRepository { // Return exported type
	return &PgColumnRepository{ // Use exported type
		q: db.New(pool),
	}
}

// FindPublished retrieves paginated published columns
func (r *PgColumnRepository) FindPublished(ctx context.Context, limit, offset int) ([]*Column, error) { // Use local Column slice
	params := db.ListPublishedColumnsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbColumns, err := r.q.ListPublishedColumns(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list published columns: %w", err)
	}

	columns := make([]*Column, len(dbColumns)) // Use local Column slice
	for i, dbColumn := range dbColumns {
		columns[i] = toLocalColumn(dbColumn) // Use local conversion func
	}

	return columns, nil
}

// FindByID retrieves a column by ID
func (r *PgColumnRepository) FindByID(ctx context.Context, id uuid.UUID) (*Column, error) { // Use local Column
	dbColumn, err := r.q.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrColumnNotFound // Use local error
		}
		return nil, fmt.Errorf("failed to find column: %w", err) // Use fmt.Errorf
	}

	return toLocalColumn(dbColumn), nil // Use local conversion func
}

// FindByCategory retrieves paginated columns by category
func (r *PgColumnRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*Column, error) { // Use local Column slice
	params := db.ListColumnsByCategoryParams{
		Category: pgtype.Text{String: category, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	}

	dbColumns, err := r.q.ListColumnsByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by category: %w", err)
	}

	columns := make([]*Column, len(dbColumns)) // Use local Column slice
	for i, dbColumn := range dbColumns {
		columns[i] = toLocalColumn(dbColumn) // Use local conversion func
	}

	return columns, nil
}

// FindByTag retrieves paginated columns by tag
func (r *PgColumnRepository) FindByTag(ctx context.Context, tag string, limit, offset int) ([]*Column, error) { // Use local Column slice
	params := db.ListColumnsByTagParams{
		Column1: tag,
		Limit:   int32(limit),
		Offset:  int32(offset),
	}
	dbColumns, err := r.q.ListColumnsByTag(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by tag: %w", err) // Use fmt.Errorf
	}

	columns := make([]*Column, len(dbColumns)) // Use local Column slice
	for i, dbCol := range dbColumns {
		columns[i] = toLocalColumn(dbCol) // Use local conversion func
	}
	return columns, nil
}

// CountPublished returns the total number of published columns
func (r *PgColumnRepository) CountPublished(ctx context.Context) (int64, error) {
	count, err := r.q.CountPublishedColumns(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count published columns: %w", err)
	}

	return count, nil
}

// CountByCategory returns the total number of published columns in a category
func (r *PgColumnRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	count, err := r.q.CountColumnsByCategory(ctx, pgtype.Text{String: category, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by category: %w", err) // Use fmt.Errorf
	}

	return count, nil
}

// CountByTag returns the total number of published columns with a tag
func (r *PgColumnRepository) CountByTag(ctx context.Context, tag string) (int64, error) {
	count, err := r.q.CountColumnsByTag(ctx, tag)
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by tag: %w", err) // Use fmt.Errorf
	}

	return count, nil
}

// toLocalColumn converts a db.Column (sqlc-generated) to a local Column
func toLocalColumn(dbColumn db.Column) *Column { // Return local Column
	var category *string
	if dbColumn.Category.Valid {
		category = &dbColumn.Category.String
	}

	var publishedAt *time.Time
	if dbColumn.PublishedAt.Valid {
		publishedAt = &dbColumn.PublishedAt.Time
	}

	return &Column{ // Use local Column
		ID:          dbColumn.ID,
		Title:       dbColumn.Title,
		Content:     dbColumn.Content,
		Category:    category,
		Tags:        dbColumn.Tags,
		PublishedAt: publishedAt,
		CreatedAt:   dbColumn.CreatedAt,
		UpdatedAt:   dbColumn.UpdatedAt,
	}
}
