package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/repo"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/wrapperspb"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
)

func TestCreateDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	entryDateStr := entryDate.Format("2006-01-02")

	testCases := []struct {
		name         string
		req          *v1.CreateDiaryEntryRequest
		expectError  bool
		expectedResp *v1.CreateDiaryEntryResponse
	}{
		{
			name: "Success - Full Entry",
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Test Diary Entry"),
				Content:   "This is a test diary entry content.",
				EntryDate: entryDateStr,
			},
			expectError: false,
			expectedResp: &v1.CreateDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     wrapperspb.String("Test Diary Entry"),
					Content:   "This is a test diary entry content.",
					EntryDate: entryDateStr,
				},
			},
		},
		{
			name: "Success - No Title",
			req: &v1.CreateDiaryEntryRequest{
				Title:     nil,
				Content:   "Content without a title.",
				EntryDate: entryDateStr,
			},
			expectError: false,
			expectedResp: &v1.CreateDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     nil,
					Content:   "Content without a title.",
					EntryDate: entryDateStr,
				},
			},
		},
		{
			name: "Error - Missing Content",
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Title Only"),
				Content:   "",
				EntryDate: entryDateStr,
			},
			expectError:  true,
			expectedResp: nil,
		},
		{
			name: "Error - Invalid Date Format",
			req: &v1.CreateDiaryEntryRequest{
				Title:     wrapperspb.String("Bad Date"),
				Content:   "Some content",
				EntryDate: "2023/01/01",
			},
			expectError:  true,
			expectedResp: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := repo.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			req := connect.NewRequest(tc.req)
			resp, err := handler.CreateDiaryEntry(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.DiaryEntry)
				require.NotEmpty(t, resp.Msg.DiaryEntry.Id)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.DiaryEntry{}, "id", "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("CreateDiaryEntry response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestUpdateDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	originalTitle := "Original Title"
	originalContent := "Original content."

	entryDateStr := entryDate.Format("2006-01-02")

	testCases := []struct {
		name         string
		setup        func(t *testing.T, ctx context.Context) db.DiaryEntry
		req          *v1.UpdateDiaryEntryRequest
		expectError  bool
		expectedResp *v1.UpdateDiaryEntryResponse
		verifyAfter func(t *testing.T, handler *DiaryHandler, testCtx context.Context, entryID uuid.UUID)
	}{
		{
			name: "Success - Update Title and Content",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				entry, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entry
			},
			req: &v1.UpdateDiaryEntryRequest{
				Title:   wrapperspb.String("Updated Title"),
				Content: "Updated content.",
			},
			expectError: false,
			expectedResp: &v1.UpdateDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     wrapperspb.String("Updated Title"),
					Content:   "Updated content.",
					EntryDate: entryDateStr,
				},
			},
			verifyAfter: nil,
		},
		{
			name: "Success - Update Only Title",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				entry, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entry
			},
			req: &v1.UpdateDiaryEntryRequest{
				Title:   wrapperspb.String("New Title Only"),
				Content: originalContent,
			},
			expectError: false,
			expectedResp: &v1.UpdateDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     wrapperspb.String("New Title Only"),
					Content:   originalContent,
					EntryDate: entryDateStr,
				},
			},
			verifyAfter: nil,
		},
		{
			name: "Success - Update Only Content",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				entry, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, originalTitle, originalContent, entryDate)
				require.NoError(t, err)
				return entry
			},
			req: &v1.UpdateDiaryEntryRequest{
				Content: "New Content Only",
				Title:   wrapperspb.String(originalTitle),
			},
			expectError: false,
			expectedResp: &v1.UpdateDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     wrapperspb.String(originalTitle),
					Content:   "New Content Only",
					EntryDate: entryDateStr,
				},
			},
			verifyAfter: nil,
		},
		{
			name: "Error - Update Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{ID: uuid.New()}
			},
			req: &v1.UpdateDiaryEntryRequest{
				Title:   wrapperspb.String("Doesn't Matter"),
				Content: "Doesn't Matter",
			},
			expectError:  true,
			expectedResp: nil,
			verifyAfter:  nil,
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{}
			},
			req: &v1.UpdateDiaryEntryRequest{
				Id:      "invalid-uuid",
				Title:   wrapperspb.String("Doesn't Matter"),
				Content: "Doesn't Matter",
			},
			expectError:  true,
			expectedResp: nil,
			verifyAfter:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := repo.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			dbEntry := tc.setup(t, ctx)
			req := connect.NewRequest(tc.req)
			if tc.req.Id == "" {
				req.Msg.Id = dbEntry.ID.String()
				if tc.expectedResp != nil && tc.expectedResp.DiaryEntry != nil {
					tc.expectedResp.DiaryEntry.Id = dbEntry.ID.String()
				}
			}

			resp, err := handler.UpdateDiaryEntry(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.DiaryEntry{}, "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("UpdateDiaryEntry response mismatch (-want +got):\n%s", diff)
				}
			}

			if tc.verifyAfter != nil {
				tc.verifyAfter(t, handler, testCtx, dbEntry.ID)
			}
		})
	}
}

func TestGetDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	testTitle := "Test Title"
	testContent := "Test content."

	entryDateStr := entryDate.Format("2006-01-02")

	testCases := []struct {
		name         string
		setup        func(t *testing.T, ctx context.Context) db.DiaryEntry
		reqID        func(entry db.DiaryEntry) string
		expectError  bool
		expectedResp *v1.GetDiaryEntryResponse
	}{
		{
			name: "Success - Get Existing Entry",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				entry, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, testTitle, testContent, entryDate)
				require.NoError(t, err)
				return entry
			},
			reqID:       func(entry db.DiaryEntry) string { return entry.ID.String() },
			expectError: false,
			expectedResp: &v1.GetDiaryEntryResponse{
				DiaryEntry: &v1.DiaryEntry{
					UserId:    testUserID.String(),
					Title:     wrapperspb.String(testTitle),
					Content:   testContent,
					EntryDate: entryDateStr,
				},
			},
		},
		{
			name: "Error - Get Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{ID: uuid.New()}
			},
			reqID:        func(entry db.DiaryEntry) string { return entry.ID.String() },
			expectError:  true,
			expectedResp: nil,
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{}
			},
			reqID:        func(entry db.DiaryEntry) string { return "invalid-uuid" },
			expectError:  true,
			expectedResp: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := repo.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			dbEntry := tc.setup(t, ctx)
			reqIDStr := tc.reqID(dbEntry)
			req := connect.NewRequest(&v1.GetDiaryEntryRequest{
				Id: reqIDStr,
			})

			if !tc.expectError && tc.expectedResp != nil && tc.expectedResp.DiaryEntry != nil {
				tc.expectedResp.DiaryEntry.Id = reqIDStr
			}

			resp, err := handler.GetDiaryEntry(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.DiaryEntry{}, "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("GetDiaryEntry response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListDiaryEntries(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewDiaryEntryRepository(testPool)
	handler := NewDiaryHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test entries
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	entryToday, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, "Today's Entry", "Content for today", today)
	require.NoError(t, err)
	entryYesterday, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, "Yesterday's Entry", "Content for yesterday", yesterday)
	require.NoError(t, err)

	// Convert to proto
	protoToday := ToProtoDiaryEntry(entryToday)
	protoYesterday := ToProtoDiaryEntry(entryYesterday)

	testCases := []struct {
		name         string
		req          *v1.ListDiaryEntriesRequest
		expectError  bool
		expectedResp *v1.ListDiaryEntriesResponse
	}{
		{
			name: "Default Pagination",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListDiaryEntriesResponse{
				DiaryEntries: []*v1.DiaryEntry{protoToday, protoYesterday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListDiaryEntriesResponse{
				DiaryEntries: []*v1.DiaryEntry{protoToday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectError: false,
			expectedResp: &v1.ListDiaryEntriesResponse{
				DiaryEntries: []*v1.DiaryEntry{protoYesterday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 2,
				},
			},
		},
		{
			name: "Pagination - Page 3 Size 1 (Empty)",
			req: &v1.ListDiaryEntriesRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 3},
			},
			expectError: false,
			expectedResp: &v1.ListDiaryEntriesResponse{
				DiaryEntries: []*v1.DiaryEntry{},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 3,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.req)
			resp, err := handler.ListDiaryEntries(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.DiaryEntry{}, "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
					cmpopts.SortSlices(func(a, b *v1.DiaryEntry) bool {
						return a.EntryDate > b.EntryDate
					}),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("ListDiaryEntries response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteDiaryEntry(t *testing.T) {
	entryDate := time.Now().UTC().Truncate(24 * time.Hour)
	titleToDelete := "Entry to Delete"
	contentToDelete := "This entry will be deleted."

	testCases := []struct {
		name         string
		setup        func(t *testing.T, ctx context.Context) db.DiaryEntry
		reqID        func(entry db.DiaryEntry) string
		expectError  bool
		expectedResp *v1.DeleteDiaryEntryResponse
		verifyAfter  func(t *testing.T, handler *DiaryHandler, testCtx context.Context, entryID uuid.UUID)
	}{
		{
			name: "Success - Delete Existing Entry",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				entry, err := testutil.CreateTestDiaryEntry(ctx, testQueries, testUserID, titleToDelete, contentToDelete, entryDate)
				require.NoError(t, err)
				return entry
			},
			reqID:       func(entry db.DiaryEntry) string { return entry.ID.String() },
			expectError: false,
			expectedResp: &v1.DeleteDiaryEntryResponse{
				Success: true,
			},
			verifyAfter: func(t *testing.T, handler *DiaryHandler, testCtx context.Context, entryID uuid.UUID) {
				getReq := connect.NewRequest(&v1.GetDiaryEntryRequest{Id: entryID.String()})
				_, getErr := handler.GetDiaryEntry(testCtx, getReq)
				require.Error(t, getErr, "Expected error when getting deleted entry, got nil")
			},
		},
		{
			name: "Error - Delete Non-existent Entry",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{ID: uuid.New()}
			},
			reqID:        func(entry db.DiaryEntry) string { return entry.ID.String() },
			expectError:  true,
			expectedResp: nil,
			verifyAfter:  nil,
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) db.DiaryEntry {
				return db.DiaryEntry{}
			},
			reqID:        func(entry db.DiaryEntry) string { return "invalid-uuid" },
			expectError:  true,
			expectedResp: nil,
			verifyAfter:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := repo.NewDiaryEntryRepository(testPool)
			handler := NewDiaryHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			dbEntry := tc.setup(t, ctx)
			reqIDStr := tc.reqID(dbEntry)
			req := connect.NewRequest(&v1.DeleteDiaryEntryRequest{
				Id: reqIDStr,
			})

			resp, err := handler.DeleteDiaryEntry(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("DeleteDiaryEntry response mismatch (-want +got):\n%s", diff)
				}
			}

			if tc.verifyAfter != nil {
				tc.verifyAfter(t, handler, testCtx, dbEntry.ID)
			}
		})
	}
}
