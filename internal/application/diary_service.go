package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
)

// DiaryService defines the interface for diary entry application service
type DiaryService interface {
	CreateDiaryEntry(ctx context.Context, userID uuid.UUID, title *string, content string, entryDate time.Time) (*domain.DiaryEntry, error)
	UpdateDiaryEntry(ctx context.Context, id uuid.UUID, userID uuid.UUID, title *string, content string) (*domain.DiaryEntry, error)
	ListDiaryEntries(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.DiaryEntry, int64, error)
	GetDiaryEntry(ctx context.Context, id, userID uuid.UUID) (*domain.DiaryEntry, error)
	DeleteDiaryEntry(ctx context.Context, id, userID uuid.UUID) error
}

// diaryServiceImpl implements the DiaryService interface
type diaryServiceImpl struct {
	repo domain.DiaryEntryRepository
	log  *slog.Logger
}

// NewDiaryService creates a new diary entry service
func NewDiaryService(repo domain.DiaryEntryRepository, log *slog.Logger) DiaryService {
	return &diaryServiceImpl{
		repo: repo,
		log:  log,
	}
}

// CreateDiaryEntry creates a new diary entry
func (s *diaryServiceImpl) CreateDiaryEntry(ctx context.Context, userID uuid.UUID, title *string, content string, entryDate time.Time) (*domain.DiaryEntry, error) {
	entry := &domain.DiaryEntry{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Content:   content,
		EntryDate: entryDate,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate the entry
	if err := entry.Validate(); err != nil {
		s.log.WarnContext(ctx, "Validation failed for diary entry", "userID", userID, "error", err)
		return nil, fmt.Errorf("invalid diary entry data: %w", err)
	}

	s.log.InfoContext(ctx, "Creating diary entry", "userID", userID, "entryDate", entryDate)
	savedEntry, err := s.repo.Create(ctx, entry)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to create diary entry", "userID", userID, "error", err)
		return nil, fmt.Errorf("could not create diary entry: %w", err)
	}

	return savedEntry, nil
}

// UpdateDiaryEntry updates an existing diary entry
func (s *diaryServiceImpl) UpdateDiaryEntry(ctx context.Context, id uuid.UUID, userID uuid.UUID, title *string, content string) (*domain.DiaryEntry, error) {
	// First, check if the entry exists and belongs to the user
	existingEntry, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if err == domain.ErrDiaryEntryNotFound {
			s.log.WarnContext(ctx, "Diary entry not found for update", "id", id, "userID", userID)
			return nil, domain.ErrDiaryEntryNotFound
		}
		s.log.ErrorContext(ctx, "Failed to fetch diary entry for update", "id", id, "userID", userID, "error", err)
		return nil, fmt.Errorf("could not fetch diary entry for update: %w", err)
	}

	// Update the entry fields
	existingEntry.Title = title
	existingEntry.Content = content
	existingEntry.UpdatedAt = time.Now()

	// Validate the updated entry
	if err := existingEntry.Validate(); err != nil {
		s.log.WarnContext(ctx, "Validation failed for updated diary entry", "id", id, "userID", userID, "error", err)
		return nil, fmt.Errorf("invalid diary entry data: %w", err)
	}

	s.log.InfoContext(ctx, "Updating diary entry", "id", id, "userID", userID)
	updatedEntry, err := s.repo.Update(ctx, existingEntry)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to update diary entry", "id", id, "userID", userID, "error", err)
		return nil, fmt.Errorf("could not update diary entry: %w", err)
	}

	return updatedEntry, nil
}

// ListDiaryEntries retrieves paginated diary entries for a user
func (s *diaryServiceImpl) ListDiaryEntries(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.DiaryEntry, int64, error) {
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

	s.log.InfoContext(ctx, "Fetching diary entries for user", "userID", userID, "page", page, "pageSize", pageSize)
	
	// Get entries
	entries, err := s.repo.FindByUser(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to fetch diary entries", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not fetch diary entries: %w", err)
	}

	// Get total count
	total, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		s.log.ErrorContext(ctx, "Failed to count diary entries", "userID", userID, "error", err)
		return nil, 0, fmt.Errorf("could not count diary entries: %w", err)
	}

	return entries, total, nil
}

// GetDiaryEntry retrieves a specific diary entry by ID
func (s *diaryServiceImpl) GetDiaryEntry(ctx context.Context, id, userID uuid.UUID) (*domain.DiaryEntry, error) {
	s.log.InfoContext(ctx, "Fetching diary entry", "id", id, "userID", userID)
	
	entry, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		if err == domain.ErrDiaryEntryNotFound {
			s.log.WarnContext(ctx, "Diary entry not found", "id", id, "userID", userID)
			return nil, domain.ErrDiaryEntryNotFound
		}
		s.log.ErrorContext(ctx, "Failed to fetch diary entry", "id", id, "userID", userID, "error", err)
		return nil, fmt.Errorf("could not fetch diary entry: %w", err)
	}

	return entry, nil
}

// DeleteDiaryEntry deletes a diary entry
func (s *diaryServiceImpl) DeleteDiaryEntry(ctx context.Context, id, userID uuid.UUID) error {
	s.log.InfoContext(ctx, "Deleting diary entry", "id", id, "userID", userID)
	
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if err == domain.ErrDiaryEntryNotFound {
			s.log.WarnContext(ctx, "Diary entry not found for deletion", "id", id, "userID", userID)
			return domain.ErrDiaryEntryNotFound
		}
		s.log.ErrorContext(ctx, "Failed to delete diary entry", "id", id, "userID", userID, "error", err)
		return fmt.Errorf("could not delete diary entry: %w", err)
	}

	return nil
}
