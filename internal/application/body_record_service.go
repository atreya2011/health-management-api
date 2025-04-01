package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
)

// BodyRecordService defines the interface for body record application service
type BodyRecordService interface {
	CreateOrUpdateBodyRecord(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error)
	GetBodyRecordsForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error)
	GetBodyRecordsForUserDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error)
}

// bodyRecordServiceImpl implements the BodyRecordService interface
type bodyRecordServiceImpl struct {
	repo domain.BodyRecordRepository
	log  *slog.Logger
}

// NewBodyRecordService creates a new body record service
func NewBodyRecordService(repo domain.BodyRecordRepository, log *slog.Logger) BodyRecordService {
	return &bodyRecordServiceImpl{
		repo: repo,
		log:  log,
	}
}

// CreateOrUpdateBodyRecord creates or updates a body record for a specific date
func (s *bodyRecordServiceImpl) CreateOrUpdateBodyRecord(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error) {
	record := &domain.BodyRecord{
		UserID:            userID,
		Date:              date,
		WeightKg:          weight,
		BodyFatPercentage: fatPercent,
	}

	// Validate the record
	if err := record.Validate(); err != nil {
		s.log.WarnContext(ctx, "Validation failed for body record", "userID", userID, "error", err)
		return nil, fmt.Errorf("invalid body record data: %w", err)
	}

	s.log.InfoContext(ctx, "Saving body record", "userID", userID, "date", date)
	savedRecord, err := s.repo.Save(ctx, record)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to save body record", "userID", userID, "error", err)
		return nil, fmt.Errorf("could not save body record: %w", err)
	}

	return savedRecord, nil
}

// GetBodyRecordsForUser retrieves paginated body records for a user
func (s *bodyRecordServiceImpl) GetBodyRecordsForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error) {
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

	s.log.InfoContext(ctx, "Fetching body records for user", "userID", userID, "page", page, "pageSize", pageSize)
	
	// Get records
	records, err := s.repo.FindByUser(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch body records", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not fetch body records: %w", err)
	}

	// Get total count
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count body records", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not count body records: %w", err)
	}

	return records, total, nil
}

// GetBodyRecordsForUserDateRange retrieves body records for a user within a specific date range
func (s *bodyRecordServiceImpl) GetBodyRecordsForUserDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
	s.log.InfoContext(ctx, "Fetching body records for user by date range", "userID", userID, "startDate", start, "endDate", end)
	
	records, err := s.repo.FindByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch body records by date range", "userID", userID, "error", err)
		return nil, fmt.Errorf("could not fetch body records by date range: %w", err)
	}

	return records, nil
}
