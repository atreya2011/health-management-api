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

func TestBodyRecordService_GetBodyRecordsForUser(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	page := 1
	pageSize := 10
	offset := (page - 1) * pageSize
	
	// Create mock records
	mockRecords := []*domain.BodyRecord{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Date:      time.Now().UTC().Truncate(24 * time.Hour),
			WeightKg:  floatPtr(75.5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:                uuid.New(),
			UserID:            userID,
			Date:              time.Now().UTC().Truncate(24 * time.Hour).Add(-24 * time.Hour),
			WeightKg:          floatPtr(76.0),
			BodyFatPercentage: floatPtr(15.5),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}
	
	// Create a mock repository
	mockRepo := &mockBodyRecordRepository{
		findByUserFunc: func(ctx context.Context, uid uuid.UUID, limit, off int) ([]*domain.BodyRecord, error) {
			// Verify input
			if uid != userID {
				t.Errorf("Expected UserID %v, got %v", userID, uid)
			}
			if limit != pageSize {
				t.Errorf("Expected limit %v, got %v", pageSize, limit)
			}
			if off != offset {
				t.Errorf("Expected offset %v, got %v", offset, off)
			}
			
			return mockRecords, nil
		},
		countByUserFunc: func(ctx context.Context, uid uuid.UUID) (int64, error) {
			// Verify input
			if uid != userID {
				t.Errorf("Expected UserID %v, got %v", userID, uid)
			}
			
			return int64(len(mockRecords)), nil
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	records, total, err := service.GetBodyRecordsForUser(context.Background(), userID, page, pageSize)
	
	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the results
	if len(records) != len(mockRecords) {
		t.Errorf("Expected %d records, got %d", len(mockRecords), len(records))
	}
	
	if total != int64(len(mockRecords)) {
		t.Errorf("Expected total %d, got %d", len(mockRecords), total)
	}
}

func TestBodyRecordService_GetBodyRecordsForUser_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	page := 1
	pageSize := 10
	
	// Create a mock repository that returns an error
	mockRepo := &mockBodyRecordRepository{
		findByUserFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*domain.BodyRecord, error) {
			return nil, errors.New("database error")
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	records, total, err := service.GetBodyRecordsForUser(context.Background(), userID, page, pageSize)
	
	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	// Verify the results
	if records != nil {
		t.Errorf("Expected nil records, got %v", records)
	}
	
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
}

func TestBodyRecordService_GetBodyRecordsForUserDateRange(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	startDate := time.Now().UTC().Truncate(24 * time.Hour).Add(-7 * 24 * time.Hour)
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	
	// Create mock records
	mockRecords := []*domain.BodyRecord{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Date:      endDate,
			WeightKg:  floatPtr(75.5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:                uuid.New(),
			UserID:            userID,
			Date:              startDate,
			WeightKg:          floatPtr(76.0),
			BodyFatPercentage: floatPtr(15.5),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}
	
	// Create a mock repository
	mockRepo := &mockBodyRecordRepository{
		findByUserAndDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
			// Verify input
			if uid != userID {
				t.Errorf("Expected UserID %v, got %v", userID, uid)
			}
			if !start.Equal(startDate) {
				t.Errorf("Expected startDate %v, got %v", startDate, start)
			}
			if !end.Equal(endDate) {
				t.Errorf("Expected endDate %v, got %v", endDate, end)
			}
			
			return mockRecords, nil
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	records, err := service.GetBodyRecordsForUserDateRange(context.Background(), userID, startDate, endDate)
	
	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the results
	if len(records) != len(mockRecords) {
		t.Errorf("Expected %d records, got %d", len(mockRecords), len(records))
	}
}

func TestBodyRecordService_GetBodyRecordsForUserDateRange_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	
	// Create test data
	userID := uuid.New()
	startDate := time.Now().UTC().Truncate(24 * time.Hour).Add(-7 * 24 * time.Hour)
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	
	// Create a mock repository that returns an error
	mockRepo := &mockBodyRecordRepository{
		findByUserAndDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
			return nil, errors.New("database error")
		},
	}
	
	// Create the service with the mock repository
	service := NewBodyRecordService(mockRepo, logger)
	
	// Call the method being tested
	records, err := service.GetBodyRecordsForUserDateRange(context.Background(), userID, startDate, endDate)
	
	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	// Verify the results
	if records != nil {
		t.Errorf("Expected nil records, got %v", records)
	}
}

// Helper function to create a pointer to a float64
func floatPtr(v float64) *float64 {
	return &v
}
