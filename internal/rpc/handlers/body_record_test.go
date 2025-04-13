package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	// "github.com/atreya2011/health-management-api/internal/auth" // Provided by main_test.go
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	// "github.com/google/uuid" // Provided by main_test.go's testUserID
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// testContext is now defined in main_test.go

func TestBodyRecordHandler_CreateBodyRecord(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// Setup and user creation are handled by TestMain

	// Use global testPool and testLogger
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)

	// Use background context
	ctx := context.Background()

	// Test data
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := 75.5

	// Create a request
	req := connect.NewRequest(&v1.CreateBodyRecordRequest{
		Date:     date.Format("2006-01-02"),
		WeightKg: &wrapperspb.DoubleValue{Value: weight},
	})

	// Create a test context using the helper from main_test.go (injects global testUserID)
	testCtx := newTestContext(ctx)

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

	// Use global testUserID for verification
	if resp.Msg.BodyRecord.UserId != testUserID.String() {
		t.Errorf("Expected UserID %v, got %v", testUserID.String(), resp.Msg.BodyRecord.UserId)
	}

	if resp.Msg.BodyRecord.Date != date.Format("2006-01-02") {
		t.Errorf("Expected Date %v, got %v", date.Format("2006-01-02"), resp.Msg.BodyRecord.Date)
	}

	if resp.Msg.BodyRecord.WeightKg == nil || resp.Msg.BodyRecord.WeightKg.Value != weight {
		t.Errorf("Expected WeightKg %v, got %v", weight, resp.Msg.BodyRecord.WeightKg)
	}
}

func TestBodyRecordHandler_CreateBodyRecord_Error(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// Setup and user creation are handled by TestMain

	// Use global testPool and testLogger
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)

	// Use background context
	ctx := context.Background()

	// Test data - invalid weight to trigger validation error
	date := time.Now().UTC().Truncate(24 * time.Hour)
	weight := -10.0 // Negative weight should fail validation

	// Create a request
	req := connect.NewRequest(&v1.CreateBodyRecordRequest{
		Date:     date.Format("2006-01-02"),
		WeightKg: &wrapperspb.DoubleValue{Value: weight},
	})

	// Create a test context using the helper from main_test.go
	testCtx := newTestContext(ctx)

	// Call the method being tested
	resp, err := handler.CreateBodyRecord(testCtx, req)

	// Check for errors
	if err == nil {
		t.Error("Expected error for invalid weight, got nil")
	}

	// Verify the response
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
}

func TestBodyRecordHandler_ListBodyRecords(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// Setup and user creation are handled by TestMain

	// Use global testPool and testLogger
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)

	// Use background context
	ctx := context.Background()

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	weight1 := 75.5
	weight2 := 76.0
	bodyFat := 15.5

	// Use global testQueries and testUserID
	_, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Handler already created above using global resources

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

	// Create a test context using the helper from main_test.go
	testCtx := newTestContext(ctx)

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

func TestBodyRecordHandler_GetBodyRecordsByDateRange(t *testing.T) {
	resetDB(t, testPool) // Reset DB state for this test
	// Setup and user creation are handled by TestMain

	// Use global testPool and testLogger
	repo := testutil.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)

	// Use background context
	ctx := context.Background()

	// Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)

	weight1 := 75.5
	weight2 := 76.0
	weight3 := 77.0
	bodyFat := 15.5

	// Use global testQueries and testUserID
	_, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	_, err = testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, lastWeek, &weight3, nil)
	if err != nil {
		t.Fatalf("Failed to create test body record: %v", err)
	}

	// Handler already created above using global resources

	// Test date range (last 3 days)
	startDate := today.Add(-3 * 24 * time.Hour)
	endDate := today

	// Create a request
	req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	})

	// Create a test context using the helper from main_test.go
	testCtx := newTestContext(ctx)

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
