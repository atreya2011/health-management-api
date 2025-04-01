package handlers

import (
	"context"
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
