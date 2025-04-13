package postgres

import (
	"context"
	"errors"
	"fmt"

	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrColumnNotFound is returned when a column is not found
var ErrColumnNotFound = errors.New("column not found")

// ColumnRepository provides database operations for Column
type ColumnRepository struct {
	q *db.Queries
}

// NewColumnRepository creates a new PostgreSQL column repository
func NewColumnRepository(pool *pgxpool.Pool) *ColumnRepository { // Return exported type
	return &ColumnRepository{ // Use exported type
		q: db.New(pool),
	}
}

// FindPublished retrieves paginated published columns
func (r *ColumnRepository) FindPublished(ctx context.Context, limit, offset int) ([]db.Column, error) {
	params := db.ListPublishedColumnsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbColumns, err := r.q.ListPublishedColumns(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list published columns: %w", err)
	}

	// Return generated structs directly
	return dbColumns, nil
}

// FindByID retrieves a column by ID
func (r *ColumnRepository) FindByID(ctx context.Context, id uuid.UUID) (db.Column, error) {
	dbColumn, err := r.q.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Column{}, ErrColumnNotFound // Return zero value and local error
		}
		return db.Column{}, fmt.Errorf("failed to find column: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbColumn, nil
}

// FindByCategory retrieves paginated columns by category
func (r *ColumnRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]db.Column, error) {
	params := db.ListColumnsByCategoryParams{
		Category: pgtype.Text{String: category, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	}

	dbColumns, err := r.q.ListColumnsByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by category: %w", err)
	}

	// Return generated structs directly
	return dbColumns, nil
}

// FindByTag retrieves paginated columns by tag
func (r *ColumnRepository) FindByTag(ctx context.Context, tag string, limit, offset int) ([]db.Column, error) {
	params := db.ListColumnsByTagParams{
		Column1: tag,
		Limit:   int32(limit),
		Offset:  int32(offset),
	}
	dbColumns, err := r.q.ListColumnsByTag(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by tag: %w", err) // Use fmt.Errorf
	}

	// Return generated structs directly
	return dbColumns, nil
}

// CountPublished returns the total number of published columns
func (r *ColumnRepository) CountPublished(ctx context.Context) (int64, error) {
	count, err := r.q.CountPublishedColumns(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count published columns: %w", err)
	}

	return count, nil
}

// CountByCategory returns the total number of published columns in a category
func (r *ColumnRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	count, err := r.q.CountColumnsByCategory(ctx, pgtype.Text{String: category, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by category: %w", err) // Use fmt.Errorf
	}

	return count, nil
}

// CountByTag returns the total number of published columns with a tag
func (r *ColumnRepository) CountByTag(ctx context.Context, tag string) (int64, error) {
	count, err := r.q.CountColumnsByTag(ctx, tag)
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by tag: %w", err) // Use fmt.Errorf
	}

	return count, nil
}

// Removed toLocalColumn function as it's no longer needed
