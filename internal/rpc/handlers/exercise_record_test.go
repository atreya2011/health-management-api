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
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	db "github.com/atreya2011/health-management-api/internal/repo/gen"
)

func TestCreateExerciseRecord(t *testing.T) {
	// Set a fixed time for the test
	fixedTime := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	mockClock.SetTime(fixedTime) // Use the global mockClock

	recordedAt := mockClock.Now().UTC() // Use mockClock
	recordedAtPb := timestamppb.New(recordedAt)
	fixedTimestampPb := timestamppb.New(fixedTime)

	testCases := []struct {
		name         string
		req          *v1.CreateExerciseRecordRequest
		expectError  bool
		expectedResp *v1.CreateExerciseRecordResponse
	}{
		{
			name: "Success - Full Record",
			req: &v1.CreateExerciseRecordRequest{
				ExerciseName:    "Running",
				DurationMinutes: wrapperspb.Int32(30),
				CaloriesBurned:  wrapperspb.Int32(250),
				RecordedAt:      recordedAtPb,
			},
			expectError: false,
			expectedResp: &v1.CreateExerciseRecordResponse{
				ExerciseRecord: &v1.ExerciseRecord{
					UserId:          testUserID.String(),
					ExerciseName:    "Running",
					DurationMinutes: wrapperspb.Int32(30),
					CaloriesBurned:  wrapperspb.Int32(250),
					RecordedAt:      recordedAtPb,
					// Add expected timestamps
					CreatedAt: fixedTimestampPb,
					UpdatedAt: fixedTimestampPb,
				},
			},
		},
		{
			name: "Success - Only Name and Time",
			req: &v1.CreateExerciseRecordRequest{
				ExerciseName: "Walking",
				RecordedAt:   recordedAtPb,
			},
			expectError: false,
			expectedResp: &v1.CreateExerciseRecordResponse{
				ExerciseRecord: &v1.ExerciseRecord{
					UserId:       testUserID.String(),
					ExerciseName: "Walking",
					RecordedAt:   recordedAtPb,
					// Add expected timestamps
					CreatedAt: fixedTimestampPb,
					UpdatedAt: fixedTimestampPb,
				},
			},
		},
		{
			name: "Error - Invalid Duration",
			req: &v1.CreateExerciseRecordRequest{
				ExerciseName:    "Cycling",
				DurationMinutes: wrapperspb.Int32(-10),
				RecordedAt:      recordedAtPb,
			},
			expectError:  true,
			expectedResp: nil,
		},
		{
			name: "Error - Missing Exercise Name",
			req: &v1.CreateExerciseRecordRequest{
				ExerciseName: "",
				RecordedAt:   recordedAtPb,
			},
			expectError:  true,
			expectedResp: nil,
		},
		{
			name: "Success - Missing RecordedAt",
			req: &v1.CreateExerciseRecordRequest{
				ExerciseName: "Swimming",
				RecordedAt:   nil, // RecordedAt will be set to mockClock.Now() by the handler
			},
			expectError: false,
			expectedResp: &v1.CreateExerciseRecordResponse{
				ExerciseRecord: &v1.ExerciseRecord{
					UserId:       testUserID.String(),
					ExerciseName: "Swimming",
					RecordedAt:   fixedTimestampPb, // Expect handler to set this to mockClock.Now()
					// Add expected timestamps
					CreatedAt: fixedTimestampPb,
					UpdatedAt: fixedTimestampPb,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			exerciseRepo := repo.NewExerciseRecordRepository(testPool)
			handler := NewExerciseRecordHandler(exerciseRepo, testLogger, mockClock) // Pass mockClock
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			req := connect.NewRequest(tc.req)
			resp, err := handler.CreateExerciseRecord(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.ExerciseRecord)
				require.NotEmpty(t, resp.Msg.ExerciseRecord.Id)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.ExerciseRecord{}, "id"), // Only ignore ID
					// cmpopts.EquateApproxTime(time.Second), // Remove time approximation
				}
				// Special handling for nil RecordedAt case - expect it to be set by handler
				if tc.req.RecordedAt == nil && tc.expectedResp != nil && tc.expectedResp.ExerciseRecord != nil {
					// We already set the expected RecordedAt to fixedTimestampPb above
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("CreateExerciseRecord response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListExerciseRecords(t *testing.T) {
	resetDB(t, testPool)
	exerciseRepo := repo.NewExerciseRecordRepository(testPool)
	handler := NewExerciseRecordHandler(exerciseRepo, testLogger, mockClock) // Pass mockClock
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test records using mock clock
	mockClock.SetTime(time.Date(2024, 1, 15, 14, 10, 0, 0, time.UTC)) // Set time for setup
	today := mockClock.Now().UTC()
	yesterday := today.Add(-24 * time.Hour)
	duration1, duration2 := int32(30), int32(45)
	calories1, calories2 := int32(250), int32(350)
	// Assuming CreateTestExerciseRecord uses the time passed to it correctly
	now := mockClock.Now()                                                                                                             // Get current mock time
	recordToday, err := testutil.CreateTestExerciseRecord(ctx, testQueries, testUserID, "Running", &duration1, &calories1, today, now) // Pass now
	require.NoError(t, err)
	recordYesterday, err := testutil.CreateTestExerciseRecord(ctx, testQueries, testUserID, "Weight Training", &duration2, &calories2, yesterday, now) // Pass now
	require.NoError(t, err)

	protoToday := ToProtoExerciseRecord(recordToday)
	protoYesterday := ToProtoExerciseRecord(recordYesterday)

	testCases := []struct {
		name         string
		req          *v1.ListExerciseRecordsRequest
		expectError  bool
		expectedResp *v1.ListExerciseRecordsResponse
	}{
		{
			name: "Default Pagination",
			req: &v1.ListExerciseRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListExerciseRecordsResponse{
				ExerciseRecords: []*v1.ExerciseRecord{protoToday, protoYesterday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListExerciseRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListExerciseRecordsResponse{
				ExerciseRecords: []*v1.ExerciseRecord{protoToday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListExerciseRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectError: false,
			expectedResp: &v1.ListExerciseRecordsResponse{
				ExerciseRecords: []*v1.ExerciseRecord{protoYesterday},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 2,
				},
			},
		},
		{
			name: "Pagination - Page 3 Size 1 (Empty)",
			req: &v1.ListExerciseRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 3},
			},
			expectError: false,
			expectedResp: &v1.ListExerciseRecordsResponse{
				ExerciseRecords: []*v1.ExerciseRecord{},
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
			resp, err := handler.ListExerciseRecords(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					// protocmp.IgnoreFields(&v1.ExerciseRecord{}, "created_at", "updated_at"), // Compare timestamps now
					// cmpopts.EquateApproxTime(time.Second), // Remove time approximation
					cmpopts.SortSlices(func(a, b *v1.ExerciseRecord) bool {
						// Keep sorting by recorded time for consistent order
						if a.RecordedAt == nil || b.RecordedAt == nil {
							return a.RecordedAt != nil // Treat nil as earliest
						}
						return a.RecordedAt.AsTime().After(b.RecordedAt.AsTime())
					}),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("ListExerciseRecords response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestDeleteExerciseRecord(t *testing.T) {
	// Set a fixed time for the test
	fixedTime := time.Date(2024, 1, 15, 14, 20, 0, 0, time.UTC)
	mockClock.SetTime(fixedTime) // Use the global mockClock

	recordedAt := mockClock.Now().UTC() // Use mockClock
	duration := int32(30)
	calories := int32(250)

	testCases := []struct {
		name         string
		setup        func(t *testing.T, ctx context.Context) db.ExerciseRecord
		reqID        func(record db.ExerciseRecord) string
		expectError  bool
		expectedResp *v1.DeleteExerciseRecordResponse
		verifyAfter  func(t *testing.T, handler *ExerciseRecordHandler, testCtx context.Context, recordID uuid.UUID)
	}{
		{
			name: "Success - Delete Existing Record",
			setup: func(t *testing.T, ctx context.Context) db.ExerciseRecord {
				now := mockClock.Now()                                                                                                                    // Get current mock time for setup
				record, err := testutil.CreateTestExerciseRecord(ctx, testQueries, testUserID, "Record to Delete", &duration, &calories, recordedAt, now) // Pass now
				require.NoError(t, err)
				return record
			},
			reqID:       func(record db.ExerciseRecord) string { return record.ID.String() },
			expectError: false,
			expectedResp: &v1.DeleteExerciseRecordResponse{
				Success: true,
			},
			verifyAfter: func(t *testing.T, handler *ExerciseRecordHandler, testCtx context.Context, recordID uuid.UUID) {
				listReq := connect.NewRequest(&v1.ListExerciseRecordsRequest{Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1}})
				listResp, listErr := handler.ListExerciseRecords(testCtx, listReq)
				require.NoError(t, listErr)
				require.NotNil(t, listResp)
				require.NotNil(t, listResp.Msg)
				assert.Len(t, listResp.Msg.ExerciseRecords, 0, "Expected 0 records after deletion")
			},
		},
		{
			name: "Error - Delete Non-existent Record",
			setup: func(t *testing.T, ctx context.Context) db.ExerciseRecord {
				return db.ExerciseRecord{ID: uuid.New()}
			},
			reqID:       func(record db.ExerciseRecord) string { return record.ID.String() },
			expectError: false,
			expectedResp: &v1.DeleteExerciseRecordResponse{
				Success: true,
			},
			verifyAfter: nil,
		},
		{
			name: "Error - Invalid ID Format",
			setup: func(t *testing.T, ctx context.Context) db.ExerciseRecord {
				return db.ExerciseRecord{}
			},
			reqID:        func(record db.ExerciseRecord) string { return "invalid-uuid" },
			expectError:  true,
			expectedResp: nil,
			verifyAfter:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			exerciseRepo := repo.NewExerciseRecordRepository(testPool)
			handler := NewExerciseRecordHandler(exerciseRepo, testLogger, mockClock) // Pass mockClock
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			dbRecord := tc.setup(t, ctx)
			reqIDStr := tc.reqID(dbRecord)
			req := connect.NewRequest(&v1.DeleteExerciseRecordRequest{
				Id: reqIDStr,
			})

			resp, err := handler.DeleteExerciseRecord(testCtx, req)

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
					t.Errorf("DeleteExerciseRecord response mismatch (-want +got):\n%s", diff)
				}
			}

			if tc.verifyAfter != nil {
				tc.verifyAfter(t, handler, testCtx, dbRecord.ID)
			}
		})
	}
}
