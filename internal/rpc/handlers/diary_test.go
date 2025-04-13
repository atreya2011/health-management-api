package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDiaryHandler_CreateDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	entryDateStr := entryDate.Format("2006-01-02")

	testCases := []struct {
		name        string
		req         *v1.CreateDiaryEntryRequest
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.CreateDiaryEntryResponse], err error)
	}{
		{
			name: "Success - Full Entry",
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Test Diary Entry"),
				Content:   "This is a test diary entry content.",
				EntryDate: entryDateStr,
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateDiaryEntryResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.DiaryEntry)
				assert.Equal(t, testUserID.String(), resp.Msg.DiaryEntry.UserId)
				assert.Equal(t, entryDateStr, resp.Msg.DiaryEntry.EntryDate)
				require.NotNil(t, resp.Msg.DiaryEntry.Title)
				assert.Equal(t, "Test Diary Entry", resp.Msg.DiaryEntry.Title.Value)
				assert.Equal(t, "This is a test diary entry content.", resp.Msg.DiaryEntry.Content)
			},
		},
		{
			name: "Success - No Title",
			req: &v1.CreateDiaryEntryRequest{
				Title:     nil, // No title provided
				Content:   "Content without a title.",
				EntryDate: entryDateStr,
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.CreateDiaryEntryResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg.DiaryEntry)
				assert.Nil(t, resp.Msg.DiaryEntry.Title)
				assert.Equal(t, "Content without a title.", resp.Msg.DiaryEntry.Content)
			},
		},
		{
			name: "Error - Missing Content", // Assuming content is required by validation
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Title Only"),
				Content:   "", // Empty content
				EntryDate: entryDateStr,
			},
			expectError: true, // Expect validation error
			verify: func(t *testing.T, resp *connect.Response[v1.CreateDiaryEntryResponse], err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
			},
		},
		{
			name: "Error - Invalid Date Format",
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Bad Date"),
				Content:   "Some content",
				EntryDate: "2023/01/01", // Invalid format
			},
			expectError: true, // Expect date parsing error
			verify: func(t *testing.T, resp *connect.Response[v1.CreateDiaryEntryResponse], err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := testutil.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			req := connect.NewRequest(tc.req)
			resp, err := handler.CreateDiaryEntry(testCtx, req)

			tc.verify(t, resp, err)
		})
	}
}

func TestDiaryHandler_UpdateDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	originalTitle := "Original Title"
	originalContent := "Original content."

	testCases := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context) uuid.UUID // Returns the ID of the entry to update
		req         *v1.UpdateDiaryEntryRequest
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID)
	}{
		{
			name: "Success - Update Title and Content",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				entryID, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entryID
			},
			req: &v1.UpdateDiaryEntryRequest{
				// ID will be set dynamically in the loop
				Title:   wrapperspb.String("Updated Title"),
				Content: "Updated content.",
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.DiaryEntry)
				assert.Equal(t, entryID.String(), resp.Msg.DiaryEntry.Id)
				require.NotNil(t, resp.Msg.DiaryEntry.Title)
				assert.Equal(t, "Updated Title", resp.Msg.DiaryEntry.Title.Value)
				assert.Equal(t, "Updated content.", resp.Msg.DiaryEntry.Content)
				// Ensure date and user ID remain unchanged
				assert.Equal(t, entryDate.Format("2006-01-02"), resp.Msg.DiaryEntry.EntryDate)
				assert.Equal(t, testUserID.String(), resp.Msg.DiaryEntry.UserId)
			},
		},
		{
			name: "Success - Update Only Title",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				entryID, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entryID
			},
			req: &v1.UpdateDiaryEntryRequest{
				Title: wrapperspb.String("New Title Only"),
				Content: originalContent, // Include original content to pass validation if needed
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID) {
				require.NoError(t, err, "UpdateDiaryEntry failed unexpectedly")

				// Fetch the entry again to verify the actual state
				handler := NewDiaryHandler(testutil.NewDiaryEntryRepository(testPool), testLogger)
				getReq := connect.NewRequest(&v1.GetDiaryEntryRequest{Id: entryID.String()})
				getResp, getErr := handler.GetDiaryEntry(newTestContext(context.Background()), getReq)

				require.NoError(t, getErr, "Failed to get entry after update")
				require.NotNil(t, getResp)
				require.NotNil(t, getResp.Msg)
				require.NotNil(t, getResp.Msg.DiaryEntry)
				require.NotNil(t, getResp.Msg.DiaryEntry.Title)
				assert.Equal(t, "New Title Only", getResp.Msg.DiaryEntry.Title.Value)
				assert.Equal(t, originalContent, getResp.Msg.DiaryEntry.Content) // Content should remain original
			},
		},
		{
			name: "Success - Update Only Content",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				entryID, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entryID
			},
			req: &v1.UpdateDiaryEntryRequest{
				Content: "New Content Only",
				Title: wrapperspb.String(originalTitle), // Include original title
			},
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID) {
				require.NoError(t, err, "UpdateDiaryEntry failed unexpectedly")

				// Fetch the entry again to verify the actual state
				handler := NewDiaryHandler(testutil.NewDiaryEntryRepository(testPool), testLogger)
				getReq := connect.NewRequest(&v1.GetDiaryEntryRequest{Id: entryID.String()})
				getResp, getErr := handler.GetDiaryEntry(newTestContext(context.Background()), getReq)

				require.NoError(t, getErr, "Failed to get entry after update")
				require.NotNil(t, getResp)
				require.NotNil(t, getResp.Msg)
				require.NotNil(t, getResp.Msg.DiaryEntry)
				require.NotNil(t, getResp.Msg.DiaryEntry.Title)
				assert.Equal(t, originalTitle, getResp.Msg.DiaryEntry.Title.Value) // Title should remain original
				assert.Equal(t, "New Content Only", getResp.Msg.DiaryEntry.Content)
			},
		},
		{
			name: "Error - Update Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				// No setup needed, use a random ID
				return uuid.New()
			},
			req: &v1.UpdateDiaryEntryRequest{
				Title:   wrapperspb.String("Doesn't Matter"),
				Content: "Doesn't Matter",
			},
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID) {
				require.Error(t, err) // Expect not found error
				assert.Nil(t, resp)
			},
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				// No setup needed
				return uuid.Nil // Return Nil UUID as placeholder, actual invalid ID is in req
			},
			req: &v1.UpdateDiaryEntryRequest{
				Id:      "invalid-uuid",
				Title:   wrapperspb.String("Doesn't Matter"),
				Content: "Doesn't Matter",
			},
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.UpdateDiaryEntryResponse], err error, entryID uuid.UUID) {
				require.Error(t, err) // Expect invalid argument error
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := testutil.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			entryID := tc.setup(t, ctx)
			req := connect.NewRequest(tc.req)
			// Set the ID in the request if it wasn't invalid format test
			if tc.req.Id == "" {
				req.Msg.Id = entryID.String()
			}

			resp, err := handler.UpdateDiaryEntry(testCtx, req)
			tc.verify(t, resp, err, entryID)
		})
	}
}

func TestDiaryHandler_GetDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	testTitle := "Test Title"
	testContent := "Test content."

	testCases := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context) uuid.UUID // Returns the ID of the entry to get
		reqID       func(entryID uuid.UUID) string                    // Function to generate request ID string
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.GetDiaryEntryResponse], err error)
	}{
		{
			name: "Success - Get Existing Entry",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				entryID, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, testTitle, testContent, entryDate)
				require.NoError(t, err)
				return entryID
			},
			reqID: func(entryID uuid.UUID) string { return entryID.String() },
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.GetDiaryEntryResponse], err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.DiaryEntry)
				// ID check happens implicitly via setup/reqID
				require.NotNil(t, resp.Msg.DiaryEntry.Title)
				assert.Equal(t, testTitle, resp.Msg.DiaryEntry.Title.Value)
				assert.Equal(t, testContent, resp.Msg.DiaryEntry.Content)
				assert.Equal(t, entryDate.Format("2006-01-02"), resp.Msg.DiaryEntry.EntryDate)
				assert.Equal(t, testUserID.String(), resp.Msg.DiaryEntry.UserId)
			},
		},
		{
			name: "Error - Get Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				// No setup needed, just generate a random ID
				return uuid.New()
			},
			reqID: func(entryID uuid.UUID) string { return entryID.String() },
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.GetDiaryEntryResponse], err error) {
				require.Error(t, err) // Expect not found error
				assert.Nil(t, resp)
			},
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				return uuid.Nil // Placeholder, not used
			},
			reqID: func(entryID uuid.UUID) string { return "invalid-uuid" },
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.GetDiaryEntryResponse], err error) {
				require.Error(t, err) // Expect invalid argument error
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := testutil.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			entryID := tc.setup(t, ctx)
			req := connect.NewRequest(&v1.GetDiaryEntryRequest{
				Id: tc.reqID(entryID),
			})

			resp, err := handler.GetDiaryEntry(testCtx, req)
			tc.verify(t, resp, err)
		})
	}
}

func TestDiaryHandler_ListDiaryEntries(t *testing.T) {
	resetDB(t, testPool)
	repo := testutil.NewDiaryEntryRepository(testPool)
	handler := NewDiaryHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test entries
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	_, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, "Today's Entry", "Content for today", today)
	require.NoError(t, err)
	_, err = testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, "Yesterday's Entry", "Content for yesterday", yesterday)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		req         *v1.ListDiaryEntriesRequest
		expectLen   int
		expectTotal int32
		expectPage  int32
	}{
		{
			name: "Default Pagination",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectLen:   2,
			expectTotal: 2,
			expectPage:  1,
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectLen:   1,
			expectTotal: 2,
			expectPage:  1,
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectLen:   1,
			expectTotal: 2,
			expectPage:  2,
		},
		{
			name: "Pagination - Page 3 Size 1 (Empty)",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 3},
			},
			expectLen:   0,
			expectTotal: 2,
			expectPage:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListDiaryEntries(testCtx, req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.DiaryEntries, tc.expectLen)
			require.NotNil(t, resp.Msg.Pagination)
			assert.Equal(t, tc.expectTotal, resp.Msg.Pagination.TotalItems)
			assert.Equal(t, tc.expectPage, resp.Msg.Pagination.CurrentPage)
		})
	}
}

func TestDiaryHandler_DeleteDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	titleToDelete := "Entry to Delete"
	contentToDelete := "This entry will be deleted."

	testCases := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context) uuid.UUID // Returns the ID of the entry to delete
		reqID       func(entryID uuid.UUID) string                    // Function to generate request ID string
		expectError bool
		verify      func(t *testing.T, resp *connect.Response[v1.DeleteDiaryEntryResponse], err error, entryID uuid.UUID, handler *DiaryHandler, testCtx context.Context)
	}{
		{
			name: "Success - Delete Existing Entry",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				entryID, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, titleToDelete, contentToDelete, entryDate)
				require.NoError(t, err)
				return entryID
			},
			reqID: func(entryID uuid.UUID) string { return entryID.String() },
			expectError: false,
			verify: func(t *testing.T, resp *connect.Response[v1.DeleteDiaryEntryResponse], err error, entryID uuid.UUID, handler *DiaryHandler, testCtx context.Context) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				assert.True(t, resp.Msg.Success)

				// Verify deletion by trying to get it
				getReq := connect.NewRequest(&v1.GetDiaryEntryRequest{Id: entryID.String()})
				_, getErr := handler.GetDiaryEntry(testCtx, getReq)
				require.Error(t, getErr, "Expected error when getting deleted entry, got nil")
			},
		},
		{
			name: "Error - Delete Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				// No setup needed, just generate a random ID
				return uuid.New()
			},
			reqID: func(entryID uuid.UUID) string { return entryID.String() },
			expectError: true, // Expecting an error now for non-existent delete
			verify: func(t *testing.T, resp *connect.Response[v1.DeleteDiaryEntryResponse], err error, entryID uuid.UUID, handler *DiaryHandler, testCtx context.Context) {
				// Expect a specific error (e.g., NotFound) when trying to delete a non-existent entry.
				// The exact error code might depend on the handler implementation.
				require.Error(t, err, "Expected an error when deleting non-existent entry")
				// Optionally check for a specific error code if known:
				// assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
				assert.Nil(t, resp, "Response should be nil on error")
			},
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) uuid.UUID {
				return uuid.Nil // Placeholder, not used
			},
			reqID: func(entryID uuid.UUID) string { return "invalid-uuid" },
			expectError: true,
			verify: func(t *testing.T, resp *connect.Response[v1.DeleteDiaryEntryResponse], err error, entryID uuid.UUID, handler *DiaryHandler, testCtx context.Context) {
				require.Error(t, err) // Expect invalid argument error
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := testutil.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			entryID := tc.setup(t, ctx)
			req := connect.NewRequest(&v1.DeleteDiaryEntryRequest{
				Id: tc.reqID(entryID),
			})

			resp, err := handler.DeleteDiaryEntry(testCtx, req)
			tc.verify(t, resp, err, entryID, handler, testCtx)
		})
	}
}
