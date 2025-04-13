package handlers

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDiaryHandler_CreateDiaryEntry(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a real repository
	repo := testutil.NewDiaryEntryRepository(testDB.Pool)

	// Create the handler with the real repository
	handler := NewDiaryHandler(repo, logger)

	// Test data
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	title := "Test Diary Entry"
	content := "This is a test diary entry content."

	// Create a request
	req := connect.NewRequest(&v1.CreateDiaryEntryRequest{
		Title:     &wrapperspb.StringValue{Value: title},
		Content:   content,
		EntryDate: entryDate.Format("2006-01-02"),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.CreateDiaryEntry(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil || resp.Msg.DiaryEntry == nil {
		t.Fatal("Expected response with diary entry, got nil")
	}

	if resp.Msg.DiaryEntry.UserId != userID.String() {
		t.Errorf("Expected UserID %v, got %v", userID.String(), resp.Msg.DiaryEntry.UserId)
	}

	if resp.Msg.DiaryEntry.EntryDate != entryDate.Format("2006-01-02") {
		t.Errorf("Expected EntryDate %v, got %v", entryDate.Format("2006-01-02"), resp.Msg.DiaryEntry.EntryDate)
	}

	if resp.Msg.DiaryEntry.Title == nil || resp.Msg.DiaryEntry.Title.Value != title {
		t.Errorf("Expected Title %v, got %v", title, resp.Msg.DiaryEntry.Title)
	}

	if resp.Msg.DiaryEntry.Content != content {
		t.Errorf("Expected Content %v, got %v", content, resp.Msg.DiaryEntry.Content)
	}
}

func TestDiaryHandler_UpdateDiaryEntry(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test diary entry
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	title := "Original Title"
	content := "Original content."

	entryID, err := testutil.CreateTestDiaryEntry(ctx, testDB.Queries, userID, title, content, entryDate) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test diary entry: %v", err)
	}

	// Create a real repository
	repo := testutil.NewDiaryEntryRepository(testDB.Pool)

	// Create the handler with the real repository
	handler := NewDiaryHandler(repo, logger)

	// Updated data
	updatedTitle := "Updated Title"
	updatedContent := "Updated content."

	// Create a request
	req := connect.NewRequest(&v1.UpdateDiaryEntryRequest{
		Id:      entryID.String(),
		Title:   &wrapperspb.StringValue{Value: updatedTitle},
		Content: updatedContent,
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.UpdateDiaryEntry(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil || resp.Msg.DiaryEntry == nil {
		t.Fatal("Expected response with diary entry, got nil")
	}

	if resp.Msg.DiaryEntry.Id != entryID.String() {
		t.Errorf("Expected ID %v, got %v", entryID.String(), resp.Msg.DiaryEntry.Id)
	}

	if resp.Msg.DiaryEntry.Title == nil || resp.Msg.DiaryEntry.Title.Value != updatedTitle {
		t.Errorf("Expected Title %v, got %v", updatedTitle, resp.Msg.DiaryEntry.Title)
	}

	if resp.Msg.DiaryEntry.Content != updatedContent {
		t.Errorf("Expected Content %v, got %v", updatedContent, resp.Msg.DiaryEntry.Content)
	}
}

func TestDiaryHandler_GetDiaryEntry(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test diary entry
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	title := "Test Title"
	content := "Test content."

	entryID, err := testutil.CreateTestDiaryEntry(ctx, testDB.Queries, userID, title, content, entryDate) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test diary entry: %v", err)
	}

	// Create a real repository
	repo := testutil.NewDiaryEntryRepository(testDB.Pool)

	// Create the handler with the real repository
	handler := NewDiaryHandler(repo, logger)

	// Create a request
	req := connect.NewRequest(&v1.GetDiaryEntryRequest{
		Id: entryID.String(),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.GetDiaryEntry(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil || resp.Msg.DiaryEntry == nil {
		t.Fatal("Expected response with diary entry, got nil")
	}

	if resp.Msg.DiaryEntry.Id != entryID.String() {
		t.Errorf("Expected ID %v, got %v", entryID.String(), resp.Msg.DiaryEntry.Id)
	}

	if resp.Msg.DiaryEntry.Title == nil || resp.Msg.DiaryEntry.Title.Value != title {
		t.Errorf("Expected Title %v, got %v", title, resp.Msg.DiaryEntry.Title)
	}

	if resp.Msg.DiaryEntry.Content != content {
		t.Errorf("Expected Content %v, got %v", content, resp.Msg.DiaryEntry.Content)
	}
}

func TestDiaryHandler_ListDiaryEntries(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test diary entries
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	_, err = testutil.CreateTestDiaryEntry(ctx, testDB.Queries, userID, "Today's Entry", "Content for today", today) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test diary entry: %v", err)
	}

	_, err = testutil.CreateTestDiaryEntry(ctx, testDB.Queries, userID, "Yesterday's Entry", "Content for yesterday", yesterday) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test diary entry: %v", err)
	}

	// Create a real repository
	repo := testutil.NewDiaryEntryRepository(testDB.Pool)

	// Create the handler with the real repository
	handler := NewDiaryHandler(repo, logger)

	// Test parameters
	pageSize := int32(10)
	pageNumber := int32(1)

	// Create a request
	req := connect.NewRequest(&v1.ListDiaryEntriesRequest{
		Pagination: &v1.PageRequest{
			PageSize:   pageSize,
			PageNumber: pageNumber,
		},
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.ListDiaryEntries(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	expectedCount := 2
	if len(resp.Msg.DiaryEntries) != expectedCount {
		t.Errorf("Expected %d entries, got %d", expectedCount, len(resp.Msg.DiaryEntries))
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

func TestDiaryHandler_DeleteDiaryEntry(t *testing.T) {
	// Set up the test database
	testDB := testutil.SetupTestDatabase(t)
	defer testDB.TeardownTestDatabase(t)

	// Create a logger that writes to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Create a test user
	ctx := context.Background()
	userID, err := testutil.CreateTestUser(ctx, testDB.Queries) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test diary entry
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	title := "Entry to Delete"
	content := "This entry will be deleted."

	entryID, err := testutil.CreateTestDiaryEntry(ctx, testDB.Queries, userID, title, content, entryDate) // Pass testDB.Queries
	if err != nil {
		t.Fatalf("Failed to create test diary entry: %v", err)
	}

	// Create a real repository
	repo := testutil.NewDiaryEntryRepository(testDB.Pool)

	// Create the handler with the real repository
	handler := NewDiaryHandler(repo, logger)

	// Create a request
	req := connect.NewRequest(&v1.DeleteDiaryEntryRequest{
		Id: entryID.String(),
	})

	// Create a context with a user ID
	testCtx := testContext{
		Context: ctx,
		userID:  userID,
	}

	// Call the method being tested
	resp, err := handler.DeleteDiaryEntry(testCtx, req)

	// Check for errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the response
	if resp == nil || resp.Msg == nil {
		t.Fatal("Expected response, got nil")
	}

	if !resp.Msg.Success {
		t.Errorf("Expected success to be true, got false")
	}

	// Verify the entry was deleted by trying to get it
	getReq := connect.NewRequest(&v1.GetDiaryEntryRequest{
		Id: entryID.String(),
	})

	_, err = handler.GetDiaryEntry(testCtx, getReq)
	if err == nil {
		t.Error("Expected error when getting deleted entry, got nil")
	}
}
