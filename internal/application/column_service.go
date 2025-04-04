package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
)

// ColumnService defines the interface for column application service
type ColumnService interface {
	ListPublishedColumns(ctx context.Context, page, pageSize int) ([]*domain.Column, int64, error)
	GetColumn(ctx context.Context, id uuid.UUID) (*domain.Column, error)
	ListColumnsByCategory(ctx context.Context, category string, page, pageSize int) ([]*domain.Column, int64, error)
	ListColumnsByTag(ctx context.Context, tag string, page, pageSize int) ([]*domain.Column, int64, error)
}

// columnServiceImpl implements the ColumnService interface
type columnServiceImpl struct {
	repo domain.ColumnRepository
	log  *slog.Logger
}

// NewColumnService creates a new column service
func NewColumnService(repo domain.ColumnRepository, log *slog.Logger) ColumnService {
	return &columnServiceImpl{
		repo: repo,
		log:  log,
	}
}

// ListPublishedColumns retrieves paginated published columns
func (s *columnServiceImpl) ListPublishedColumns(ctx context.Context, page, pageSize int) ([]*domain.Column, int64, error) {
	// Apply default/max page size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	// Ensure page is at least 1
	if page <= 0 {
		page = 1
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize

	s.log.InfoContext(ctx, "Fetching published columns", "page", page, "pageSize", pageSize)
	
	// Get columns
	columns, err := s.repo.FindPublished(ctx, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch published columns", "error", err)
		return nil, 0, fmt.Errorf("could not fetch published columns: %w", err)
	}

	// Get total count
	total, err := s.repo.CountPublished(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count published columns", "error", err)
		return nil, 0, fmt.Errorf("could not count published columns: %w", err)
	}

	return columns, total, nil
}

// GetColumn retrieves a specific column by ID
func (s *columnServiceImpl) GetColumn(ctx context.Context, id uuid.UUID) (*domain.Column, error) {
	s.log.InfoContext(ctx, "Fetching column", "id", id)
	
	column, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == domain.ErrColumnNotFound {
			s.log.WarnContext(ctx, "Column not found", "id", id)
			return nil, domain.ErrColumnNotFound
		}
		s.log.ErrorContext(ctx, "Failed to fetch column", "id", id, "error", err)
		return nil, fmt.Errorf("could not fetch column: %w", err)
	}

	// Check if the column is published
	if !column.IsPublished() {
		s.log.WarnContext(ctx, "Attempted to access unpublished column", "id", id)
		return nil, domain.ErrColumnNotFound
	}

	return column, nil
}

// ListColumnsByCategory retrieves paginated columns by category
func (s *columnServiceImpl) ListColumnsByCategory(ctx context.Context, category string, page, pageSize int) ([]*domain.Column, int64, error) {
	// Apply default/max page size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	// Ensure page is at least 1
	if page <= 0 {
		page = 1
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize

	s.log.InfoContext(ctx, "Fetching columns by category", "category", category, "page", page, "pageSize", pageSize)
	
	// Get columns
	columns, err := s.repo.FindByCategory(ctx, category, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch columns by category", "category", category, "error", err)
		return nil, 0, fmt.Errorf("could not fetch columns by category: %w", err)
	}

	// Get total count
	total, err := s.repo.CountByCategory(ctx, category)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count columns by category", "category", category, "error", err)
		return nil, 0, fmt.Errorf("could not count columns by category: %w", err)
	}

	return columns, total, nil
}

// ListColumnsByTag retrieves paginated columns by tag
func (s *columnServiceImpl) ListColumnsByTag(ctx context.Context, tag string, page, pageSize int) ([]*domain.Column, int64, error) {
	// Apply default/max page size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	
	// Ensure page is at least 1
	if page <= 0 {
		page = 1
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize

	s.log.InfoContext(ctx, "Fetching columns by tag", "tag", tag, "page", page, "pageSize", pageSize)
	
	// Get columns
	columns, err := s.repo.FindByTag(ctx, tag, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch columns by tag", "tag", tag, "error", err)
		return nil, 0, fmt.Errorf("could not fetch columns by tag: %w", err)
	}

	// Get total count
	total, err := s.repo.CountByTag(ctx, tag)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count columns by tag", "tag", tag, "error", err)
		return nil, 0, fmt.Errorf("could not count columns by tag: %w", err)
	}

	return columns, total, nil
}
