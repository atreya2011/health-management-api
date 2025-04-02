package handlers

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/domain"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// mockBodyRecordService is a simple mock implementation of application.BodyRecordService
type mockBodyRecordService struct {
	createOrUpdateBodyRecordFunc       func(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error)
	getBodyRecordsForUserFunc          func(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error)
	getBodyRecordsForUserDateRangeFunc func(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error)
}

func (m *mockBodyRecordService) CreateOrUpdateBodyRecord(ctx context.Context, userID uuid.UUID, date time.Time, weight *float64, fatPercent *float64) (*domain.BodyRecord, error) {
	return m.createOrUpdateBodyRecordFunc(ctx, userID, date, weight, fatPercent)
}

func (m *mockBodyRecordService) GetBodyRecordsForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error) {
	return m.getBodyRecordsForUserFunc(ctx, userID, page, pageSize)
}

func (m *mockBodyRecordService) GetBodyRecordsForUserDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
	return m.getBodyRecordsForUserDateRangeFunc(ctx, userID, start, end)
}

// mockContext is a simple context that contains a user ID
type mockContext struct {
	context.Context
	userID uuid.UUID
}

// Value implements context.Context
func (c mockContext) Value(key interface{}) interface{} {
	// This is a simplified version that only handles the user ID key
	// In a real implementation, you would check the key and return the appropriate value
	return c.userID
}

func TestBodyRecordHandler_CreateBodyRecord(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create test data
	userID := uuid.New()
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5

	// Create a mock service
	mockService := &mockBodyRecordService{
		createOrUpdateBodyRecordFunc: func(ctx context.Context, uid uuid.UUID, d time.Time, w *float64, f *float64) (*domain.BodyRecord, error) {
			// Verify input
			if uid != userID {
				t.Errorf("Expected UserID %v, got %v", userID, uid)
			}
			if !d.Equal(date) {
				t.Errorf("Expected Date %v, got %v", date, d)
			}
			if *w != weight {
				t.Errorf("Expected WeightKg %v, got %v", weight, *w)
			}

			// Return a successful result
			return &domain.BodyRecord{
				ID:        uuid.New(),
				UserID:    userID,
				Date:      date,
				WeightKg:  w,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.CreateBodyRecordRequest{
		Date:     date.Format("2006-01-02"),
		WeightKg: &wrapperspb.DoubleValue{Value: weight},
	})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateBodyRecord(ctx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil || resp.Msg.BodyRecord == nil {
		t.Fatal("Expected response with body record, got nil")
	}

	if resp.Msg.BodyRecord.UserId != userID.String() {
		t.Errorf("Expected UserID %v, got %v", userID.String(), resp.Msg.BodyRecord.UserId)
	}

	if resp.Msg.BodyRecord.Date != date.Format("2006-01-02") {
		t.Errorf("Expected Date %v, got %v", date.Format("2006-01-02"), resp.Msg.BodyRecord.Date)
	}

	if resp.Msg.BodyRecord.WeightKg == nil || resp.Msg.BodyRecord.WeightKg.Value != weight {
		t.Errorf("Expected WeightKg %v, got %v", weight, resp.Msg.BodyRecord.WeightKg)
	}
}

func TestBodyRecordHandler_CreateBodyRecord_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create test data
	userID := uuid.New()
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5

	// Create a mock service that returns an error
	mockService := &mockBodyRecordService{
		createOrUpdateBodyRecordFunc: func(ctx context.Context, uid uuid.UUID, d time.Time, w *float64, f *float64) (*domain.BodyRecord, error) {
			return nil, errors.New("service error")
		},
	}

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.CreateBodyRecordRequest{
		Date:     date.Format("2006-01-02"),
		WeightKg: &wrapperspb.DoubleValue{Value: weight},
	})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateBodyRecord(ctx, req)

	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify the response
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
}

func TestBodyRecordHandler_ListBodyRecords(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create test data
	userID := uuid.New()
	pageSize := int32(10)
	pageNumber := int32(1)
	
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
	
	totalRecords := int64(len(mockRecords))

	// Create a mock service
	mockService := &mockBodyRecordService{
		getBodyRecordsForUserFunc: func(ctx context.Context, uid uuid.UUID, page, pSize int) ([]*domain.BodyRecord, int64, error) {
			// Verify input
			if uid != userID {
				t.Errorf("Expected UserID %v, got %v", userID, uid)
			}
			if page != int(pageNumber) {
				t.Errorf("Expected page %v, got %v", pageNumber, page)
			}
			if pSize != int(pageSize) {
				t.Errorf("Expected pageSize %v, got %v", pageSize, pSize)
			}

			return mockRecords, totalRecords, nil
		},
	}

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.ListBodyRecordsRequest{
		Pagination: &v1.PageRequest{
			PageSize:   pageSize,
			PageNumber: pageNumber,
		},
	})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.ListBodyRecords(ctx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(resp.Msg.BodyRecords) != len(mockRecords) {
		t.Errorf("Expected %d records, got %d", len(mockRecords), len(resp.Msg.BodyRecords))
	}

	if resp.Msg.Pagination == nil {
		t.Fatal("Expected pagination in response, got nil")
	}

	if resp.Msg.Pagination.TotalItems != int32(totalRecords) {
		t.Errorf("Expected total items %d, got %d", totalRecords, resp.Msg.Pagination.TotalItems)
	}

	if resp.Msg.Pagination.CurrentPage != pageNumber {
		t.Errorf("Expected current page %d, got %d", pageNumber, resp.Msg.Pagination.CurrentPage)
	}
}

func TestBodyRecordHandler_ListBodyRecords_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create test data
	userID := uuid.New()

	// Create a mock service that returns an error
	mockService := &mockBodyRecordService{
		getBodyRecordsForUserFunc: func(ctx context.Context, uid uuid.UUID, page, pageSize int) ([]*domain.BodyRecord, int64, error) {
			return nil, 0, errors.New("service error")
		},
	}

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.ListBodyRecordsRequest{})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.ListBodyRecords(ctx, req)

	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify the response
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
}

func TestBodyRecordHandler_GetBodyRecordsByDateRange(t *testing.T) {
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

	// Create a mock service
	mockService := &mockBodyRecordService{
		getBodyRecordsForUserDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
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

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.GetBodyRecordsByDateRange(ctx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(resp.Msg.BodyRecords) != len(mockRecords) {
		t.Errorf("Expected %d records, got %d", len(mockRecords), len(resp.Msg.BodyRecords))
	}
}

func TestBodyRecordHandler_GetBodyRecordsByDateRange_Error(t *testing.T) {
	// Create a logger that writes to nowhere
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create test data
	userID := uuid.New()
	startDate := time.Now().UTC().Truncate(24 * time.Hour).Add(-7 * 24 * time.Hour)
	endDate := time.Now().UTC().Truncate(24 * time.Hour)

	// Create a mock service that returns an error
	mockService := &mockBodyRecordService{
		getBodyRecordsForUserDateRangeFunc: func(ctx context.Context, uid uuid.UUID, start, end time.Time) ([]*domain.BodyRecord, error) {
			return nil, errors.New("service error")
		},
	}

	// Create the handler with the mock service
	handler := NewBodyRecordHandler(mockService, logger)

	// Create a request
	req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	})

	// Create a context with a user ID
	ctx := mockContext{
		Context: context.Background(),
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.GetBodyRecordsByDateRange(ctx, req)

	// Check for errors
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify the response
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
}

// Helper function to create a pointer to a float64
func floatPtr(v float64) *float64 {
	return &v
}
