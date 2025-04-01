package application

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
)

// mockBodyRecordRepository is a simple mock implementation of domain.BodyRecordRepository
// We're using a simple mock here instead of testify/mock to avoid dependencies
type mockBodyRecordRepository struct {
	saveFunc                    func(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error)
	findByUserFunc              func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.BodyRecord, error)
	findByUserAndDateRangeFunc  func(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.BodyRecord, error)
	countByUserFunc             func(ctx context.Context, userID uuid.UUID) (int64, error)
}

func (m *mockBodyRecordRepository) Save(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
	return m.saveFunc(ctx, record)
}

func (m *mockBodyRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.BodyRecord, error) {
	return m.findByUserFunc(ctx, userID, limit, offset)
}

func (m *mockBodyRecordRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.BodyRecord, error) {
	return m.findByUserAndDateRangeFunc(ctx, userID, startDate, endDate)
}

func (m *mockBodyRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return m.countByUserFunc(ctx, userID)
}

func TestBodyRecordService_CreateOrUpdateBodyRecord(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5
	
	// Create a mock repository
	mockRepo := &mockBodyRecordRepository{
		saveFunc: func(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
			// Verify input
			if record.UserID != userID {
				t.Errorf("Expected UserID %v, got %v", userID, record.UserID)
			}
			if !record.Date.Equal(date) {
				t.Errorf("Expected Date %v, got %v", date, record.Date)
			}
			if *record.WeightKg != weight {
				t.Errorf("Expected WeightKg %v, got %v", weight, *record.WeightKg)
			}
			
			// Return a successful result
			return &domain.BodyRecord{
				ID:        uuid.New(),
				UserID:    userID,
				Date:      date,
				WeightKg:  &weight,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	result, err := service.CreateOrUpdateBodyRecord(context.Background(), userID, date, &weight, nil)
	
	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the result
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, result.UserID)
	}
	if !result.Date.Equal(date) {
		t.Errorf("Expected Date %v, got %v", date, result.Date)
	}
	if *result.WeightKg != weight {
		t.Errorf("Expected WeightKg %v, got %v", weight, *result.WeightKg)
	}
}

func TestBodyRecordService_CreateOrUpdateBodyRecord_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5
	
	// Create a mock repository that returns an error
	mockRepo := &mockBodyRecordRepository{
		saveFunc: func(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
			return nil, errors.New("database error")
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	result, err := service.CreateOrUpdateBodyRecord(context.Background(), userID, date, &weight, nil)
	
	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	// Verify the result
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}
