package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// pgColumnRepository implements the domain.ColumnRepository interface
type pgColumnRepository struct {
	q *db.Queries
}

// NewPgColumnRepository creates a new PostgreSQL column repository
func NewPgColumnRepository(pool *pgxpool.Pool) domain.ColumnRepository {
	return &pgColumnRepository{
		q: db.New(pool),
	}
}

// FindPublished retrieves paginated published columns
func (r *pgColumnRepository) FindPublished(ctx context.Context, limit, offset int) ([]*domain.Column, error) {
	params := db.ListPublishedColumnsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbColumns, err := r.q.ListPublishedColumns(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list published columns: %w", err)
	}

	columns := make([]*domain.Column, len(dbColumns))
	for i, dbColumn := range dbColumns {
		columns[i] = toDomainColumn(dbColumn)
	}

	return columns, nil
}

// FindByID retrieves a column by ID
func (r *pgColumnRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Column, error) {
	dbColumn, err := r.q.GetColumnByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, errors.Wrap(err, "failed to find column")
	}

	return toDomainColumn(dbColumn), nil
}

// FindByCategory retrieves paginated columns by category
func (r *pgColumnRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Column, error) {
	params := db.ListColumnsByCategoryParams{
		Category: pgtype.Text{String: category, Valid: true},
		Limit:    int32(limit),
		Offset:   int32(offset),
	}

	dbColumns, err := r.q.ListColumnsByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns by category: %w", err)
	}

	columns := make([]*domain.Column, len(dbColumns))
	for i, dbColumn := range dbColumns {
		columns[i] = toDomainColumn(dbColumn)
	}

	return columns, nil
}

// FindByTag retrieves paginated columns by tag
func (r *pgColumnRepository) FindByTag(ctx context.Context, tag string, limit, offset int) ([]*domain.Column, error) {
	params := db.ListColumnsByTagParams{
		Column1: tag,
		Limit:   int32(limit),
		Offset:  int32(offset),
	}
	dbColumns, err := r.q.ListColumnsByTag(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list columns by tag")
	}

	columns := make([]*domain.Column, len(dbColumns))
	for i, dbCol := range dbColumns {
		columns[i] = toDomainColumn(dbCol)
	}
	return columns, nil
}

// CountPublished returns the total number of published columns
func (r *pgColumnRepository) CountPublished(ctx context.Context) (int64, error) {
	count, err := r.q.CountPublishedColumns(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count published columns: %w", err)
	}

	return count, nil
}

// CountByCategory returns the total number of published columns in a category
func (r *pgColumnRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	count, err := r.q.CountColumnsByCategory(ctx, pgtype.Text{String: category, Valid: true})
	if err != nil {
		return 0, errors.Wrap(err, "failed to count columns by category")
	}

	return count, nil
}

// CountByTag returns the total number of published columns with a tag
func (r *pgColumnRepository) CountByTag(ctx context.Context, tag string) (int64, error) {
	count, err := r.q.CountColumnsByTag(ctx, tag)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count columns by tag")
	}

	return count, nil
}

// toDomainColumn converts a db.Column (now pgx-based) to a domain.Column
func toDomainColumn(dbColumn db.Column) *domain.Column {
	var category *string
	if dbColumn.Category.Valid {
		category = &dbColumn.Category.String
	}

	var publishedAt *time.Time
	if dbColumn.PublishedAt.Valid {
		publishedAt = &dbColumn.PublishedAt.Time
	}

	return &domain.Column{
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
