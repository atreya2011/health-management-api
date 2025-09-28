package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
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

// FindPublished retrieves paginated published columns, accepting the current time.
func (r *ColumnRepository) FindPublished(ctx context.Context, limit, offset int, now time.Time) ([]db.Column, error) {
	params := db.ListPublishedColumnsParams{
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	}

	dbColumns, err := r.q.ListPublishedColumns(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list published columns: %w", err)
	}

	// Return generated structs directly
	return dbColumns, nil
}

// FindByID retrieves a column by ID, checking if it's published based on the current time.
func (r *ColumnRepository) FindByID(ctx context.Context, id uuid.UUID, now time.Time) (db.Column, error) {
	params := db.GetColumnByIDParams{
		ID:          id,
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}
	dbColumn, err := r.q.GetColumnByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Column{}, ErrColumnNotFound // Return zero value and local error
		}
		return db.Column{}, fmt.Errorf("failed to find column: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbColumn, nil
}

// FindByCategory retrieves paginated columns by category, accepting the current time.
func (r *ColumnRepository) FindByCategory(ctx context.Context, category string, limit, offset int, now time.Time) ([]db.Column, error) {
	params := db.ListColumnsByCategoryParams{
		Category:    pgtype.Text{String: category, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	}

	dbColumns, err := r.q.ListColumnsByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by category: %w", err)
	}

	// Return generated structs directly
	return dbColumns, nil
}

// FindByTag retrieves paginated columns by tag, accepting the current time.
func (r *ColumnRepository) FindByTag(ctx context.Context, tag string, limit, offset int, now time.Time) ([]db.Column, error) {
	params := db.ListColumnsByTagParams{
		Column1:     tag,
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	}
	dbColumns, err := r.q.ListColumnsByTag(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by tag: %w", err) // Use fmt.Errorf
	}

	// Return generated structs directly
	return dbColumns, nil
}

// CountPublished returns the total number of published columns, accepting the current time.
func (r *ColumnRepository) CountPublished(ctx context.Context, now time.Time) (int64, error) {
	count, err := r.q.CountPublishedColumns(ctx, pgtype.Timestamptz{Time: now, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("failed to count published columns: %w", err)
	}

	return count, nil
}

// CountByCategory returns the total number of published columns in a category, accepting the current time.
func (r *ColumnRepository) CountByCategory(ctx context.Context, category string, now time.Time) (int64, error) {
	params := db.CountColumnsByCategoryParams{
		Category:    pgtype.Text{String: category, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}
	count, err := r.q.CountColumnsByCategory(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by category: %w", err) // Use fmt.Errorf
	}

	return count, nil
}

// CountByTag returns the total number of published columns with a tag, accepting the current time.
func (r *ColumnRepository) CountByTag(ctx context.Context, tag string, now time.Time) (int64, error) {
	params := db.CountColumnsByTagParams{
		Column1:     tag,
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}
	count, err := r.q.CountColumnsByTag(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to count columns by tag: %w", err) // Use fmt.Errorf
	}

	return count, nil
}
