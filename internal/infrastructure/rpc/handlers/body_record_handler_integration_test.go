package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// mockContext is a simple context that contains a user ID for testing
// This is the same implementation as in the original test file
type mockContext struct {
	context.Context
	userID uuid.UUID
}

// Value implements context.Context
func (c mockContext) Value(key interface{}) interface{} {
	// This is a simplified version that only handles the user ID key
	if key == auth.UserContextKey {
		return c.userID
	}
	return c.Context.Value(key)
}

// TestBodyRecordHandler_Integration_CreateBodyRecord tests the CreateBodyRecord handler with a real database
func TestBodyRecordHandler_Integration_CreateBodyRecord(t *testing.T) {
	// Skip integration tests if not explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository and service
	repo := testutil.NewBodyRecordRepository(testDB.Pool)
	service := application.NewBodyRecordService(repo, logger)

	// Create the handler with the real service
	handler := NewBodyRecordHandler(service, logger)

	// Test data
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5

	// Create a request
	req := connect.NewRequest(&v1.CreateBodyRecordRequest{
		Date:     date.Format("2006-01-02"),
		WeightKg: &wrapperspb.DoubleValue{Value: weight},
	})

	// Create a context with a user ID
	testCtx := mockContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateBodyRecord(testCtx, req)

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

// TestBodyRecordHandler_Integration_ListBodyRecords tests the ListBodyRecords handler with a real database
func TestBodyRecordHandler_Integration_ListBodyRecords(t *testing.T) {
	// Skip integration tests if not explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	
	weight1 := 75.5
	weight2 := 76.0
	bodyFat := 15.5
	
	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}
	
	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Create a real repository and service
	repo := testutil.NewBodyRecordRepository(testDB.Pool)
	service := application.NewBodyRecordService(repo, logger)

	// Create the handler with the real service
	handler := NewBodyRecordHandler(service, logger)

	// Test parameters
	pageSize := int32(10)
	pageNumber := int32(1)

	// Create a request
	req := connect.NewRequest(&v1.ListBodyRecordsRequest{
		Pagination: &v1.PageRequest{
			PageSize:   pageSize,
			PageNumber: pageNumber,
		},
	})

	// Create a context with a user ID
	testCtx := mockContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.ListBodyRecords(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	expectedCount := 2
	if len(resp.Msg.BodyRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(resp.Msg.BodyRecords))
	}

	if resp.Msg.Pagination == nil {
		t.Fatal("Expected pagination in response, got nil")
	}

	if resp.Msg.Pagination.TotalItems != int32(expectedCount) {
		t.Errorf("Expected total items %d, got %d", expectedCount, resp.Msg.Pagination.TotalItems)
	}

	if resp.Msg.Pagination.CurrentPage != pageNumber {
		t.Errorf("Expected current page %d, got %d", pageNumber, resp.Msg.Pagination.CurrentPage)
	}
}

// TestBodyRecordHandler_Integration_GetBodyRecordsByDateRange tests the GetBodyRecordsByDateRange handler with a real database
func TestBodyRecordHandler_Integration_GetBodyRecordsByDateRange(t *testing.T) {
	// Skip integration tests if not explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.DB)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)
	
	weight1 := 75.5
	weight2 := 76.0
	weight3 := 77.0
	bodyFat := 15.5
	
	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}
	
	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}
	
	_, err = testutil.CreateTestBodyRecord(ctx, testDB.DB, userID, lastWeek, &weight3, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Create a real repository and service
	repo := testutil.NewBodyRecordRepository(testDB.Pool)
	service := application.NewBodyRecordService(repo, logger)

	// Create the handler with the real service
	handler := NewBodyRecordHandler(service, logger)

	// Test date range (last 3 days)
	startDate := today.Add(-3 * 24 * time.Hour)
	endDate := today

	// Create a request
	req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	})

	// Create a context with a user ID
	testCtx := mockContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.GetBodyRecordsByDateRange(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should only include today and yesterday, not last week
	expectedCount := 2
	if len(resp.Msg.BodyRecords) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(resp.Msg.BodyRecords))
	}
}
