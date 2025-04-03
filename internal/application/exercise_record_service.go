package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
)

// ExerciseRecordService defines the interface for exercise record application service
type ExerciseRecordService interface {
	CreateExerciseRecord(ctx context.Context, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time) (*domain.ExerciseRecord, error)
	ListExerciseRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.ExerciseRecord, int64, error)
	DeleteExerciseRecord(ctx context.Context, id, userID uuid.UUID) error
}

// exerciseRecordServiceImpl implements the ExerciseRecordService interface
type exerciseRecordServiceImpl struct {
	repo domain.ExerciseRecordRepository
	log  *slog.Logger
}

// NewExerciseRecordService creates a new exercise record service
func NewExerciseRecordService(repo domain.ExerciseRecordRepository, log *slog.Logger) ExerciseRecordService {
	return &exerciseRecordServiceImpl{
		repo: repo,
		log:  log,
	}
}

// CreateExerciseRecord creates a new exercise record
func (s *exerciseRecordServiceImpl) CreateExerciseRecord(ctx context.Context, userID uuid.UUID, exerciseName string, durationMinutes *int32, caloriesBurned *int32, recordedAt time.Time) (*domain.ExerciseRecord, error) {
	record := &domain.ExerciseRecord{
		ID:              uuid.New(),
		UserID:          userID,
		ExerciseName:    exerciseName,
		DurationMinutes: durationMinutes,
		CaloriesBurned:  caloriesBurned,
		RecordedAt:      recordedAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Validate the record
	if err := record.Validate(); err != nil {
		s.log.WarnContext(ctx, "Validation failed for exercise record", "userID", userID, "error", err)
		return nil, fmt.Errorf("invalid exercise record data: %w", err)
	}

	s.log.InfoContext(ctx, "Creating exercise record", "userID", userID, "exerciseName", exerciseName)
	savedRecord, err := s.repo.Create(ctx, record)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to create exercise record", "userID", userID, "error", err)
		return nil, fmt.Errorf("could not create exercise record: %w", err)
	}

	return savedRecord, nil
}

// ListExerciseRecords retrieves paginated exercise records for a user
func (s *exerciseRecordServiceImpl) ListExerciseRecords(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.ExerciseRecord, int64, error) {
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

	s.log.InfoContext(ctx, "Fetching exercise records for user", "userID", userID, "page", page, "pageSize", pageSize)
	
	// Get records
	records, err := s.repo.FindByUser(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch exercise records", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not fetch exercise records: %w", err)
	}

	// Get total count
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count exercise records", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not count exercise records: %w", err)
	}

	return records, total, nil
}

// DeleteExerciseRecord deletes an exercise record
func (s *exerciseRecordServiceImpl) DeleteExerciseRecord(ctx context.Context, id, userID uuid.UUID) error {
	s.log.InfoContext(ctx, "Deleting exercise record", "id", id, "userID", userID)
	
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if err == domain.ErrExerciseRecordNotFound {
			s.log.WarnContext(ctx, "Exercise record not found for deletion", "id", id, "userID", userID)
			return domain.ErrExerciseRecordNotFound
		}
		s.log.ErrorContext(ctx, "Failed to delete exercise record", "id", id, "userID", userID, "error", err)
		return fmt.Errorf("could not delete exercise record: %w", err)
	}

	return nil
}
