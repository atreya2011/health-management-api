package handlers

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/repo"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/atreya2011/health-management-api/internal/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreateBodyRecord(t *testing.T) {
	date := time.Now().UTC().Truncate(24 * time.Hour)
	dateStr := date.Format("2006-01-02")

	testCases := []struct {
		name         string
		req          *v1.CreateBodyRecordRequest
		expectError  bool
		expectedResp *v1.CreateBodyRecordResponse
	}{
		{
			name: "Success",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: 75.5},
			},
			expectError: false,
			expectedResp: &v1.CreateBodyRecordResponse{
				BodyRecord: &v1.BodyRecord{
					UserId:     testUserID.String(),
					Date:       dateStr,
					WeightKg:   &wrapperspb.DoubleValue{Value: 75.5},
					CreatedAt:  timestamppb.Now(),
					UpdatedAt:  timestamppb.Now(),
				},
			},
		},
		{
			name: "Error - Invalid Weight",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: -10.0},
			},
			expectError:  true,
			expectedResp: nil,
		},
		{
			name: "Success - Only Weight",
			req: &v1.CreateBodyRecordRequest{
				Date:     dateStr,
				WeightKg: &wrapperspb.DoubleValue{Value: 76.0},
			},
			expectError: false,
			expectedResp: &v1.CreateBodyRecordResponse{
				BodyRecord: &v1.BodyRecord{
					UserId:     testUserID.String(),
					Date:       dateStr,
					WeightKg:   &wrapperspb.DoubleValue{Value: 76.0},
					CreatedAt:  timestamppb.Now(),
					UpdatedAt:  timestamppb.Now(),
				},
			},
		},
		{
			name: "Success - With Body Fat",
			req: &v1.CreateBodyRecordRequest{
				Date:              dateStr,
				WeightKg:          &wrapperspb.DoubleValue{Value: 75.0},
				BodyFatPercentage: &wrapperspb.DoubleValue{Value: 15.5},
			},
			expectError: false,
			expectedResp: &v1.CreateBodyRecordResponse{
				BodyRecord: &v1.BodyRecord{
					UserId:            testUserID.String(),
					Date:              dateStr,
					WeightKg:          &wrapperspb.DoubleValue{Value: 75.0},
					BodyFatPercentage: &wrapperspb.DoubleValue{Value: 15.5},
					CreatedAt:         timestamppb.Now(),
					UpdatedAt:         timestamppb.Now(),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, testPool)
			repo := repo.NewBodyRecordRepository(testPool)
			handler := NewBodyRecordHandler(repo, testLogger)
			ctx := context.Background()
			testCtx := newTestContext(ctx)

			req := connect.NewRequest(tc.req)
			resp, err := handler.CreateBodyRecord(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.BodyRecord)
				require.NotEmpty(t, resp.Msg.BodyRecord.Id)

				cmpOpts := []cmp.Option{
					protocmp.Transform(),
					protocmp.IgnoreFields(&v1.BodyRecord{}, "id", "created_at", "updated_at"),
					cmpopts.EquateApproxTime(time.Second),
				}

				if diff := cmp.Diff(tc.expectedResp, resp.Msg, cmpOpts...); diff != "" {
					t.Errorf("CreateBodyRecord response mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestListBodyRecords(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	weight1 := 75.5
	weight2 := 76.0
	bodyFat := 15.5
	record1, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	require.NoError(t, err, "Failed to create test body record 1")
	record2, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	require.NoError(t, err, "Failed to create test body record 2")

	protoRecord1 := ToProtoBodyRecord(record1)
	protoRecord2 := ToProtoBodyRecord(record2)

	testCases := []struct {
		name         string
		req          *v1.ListBodyRecordsRequest
		expectError  bool
		expectedResp *v1.ListBodyRecordsResponse
	}{
		{
			name: "Default Pagination",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 10, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListBodyRecordsResponse{
				BodyRecords: []*v1.BodyRecord{protoRecord1, protoRecord2},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  1,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 1 Size 1",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 1},
			},
			expectError: false,
			expectedResp: &v1.ListBodyRecordsResponse{
				BodyRecords: []*v1.BodyRecord{protoRecord1},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 1,
				},
			},
		},
		{
			name: "Pagination - Page 2 Size 1",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 2},
			},
			expectError: false,
			expectedResp: &v1.ListBodyRecordsResponse{
				BodyRecords: []*v1.BodyRecord{protoRecord2},
				Pagination: &v1.PageResponse{
					TotalItems:  2,
					TotalPages:  2,
					CurrentPage: 2,
				},
			},
		},
		{
			name: "Pagination - Page 3 Size 1 (Empty)",
			req: &v1.ListBodyRecordsRequest{
				Pagination: &v1.PageRequest{PageSize: 1, PageNumber: 3},
			},
			expectError: false,
			expectedResp: &v1.ListBodyRecordsResponse{
				BodyRecords: []*v1.BodyRecord{},
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
			resp, err := handler.ListBodyRecords(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				if diff := cmp.Diff(tc.expectedResp.Pagination, resp.Msg.Pagination, protocmp.Transform()); diff != "" {
					t.Errorf("ListBodyRecords pagination mismatch (-want +got):\n%s", diff)
				}

				assert.Len(t, resp.Msg.BodyRecords, len(tc.expectedResp.BodyRecords), "Incorrect number of body records returned")

				if len(tc.expectedResp.BodyRecords) > 0 && len(resp.Msg.BodyRecords) > 0 {
					var foundToday, foundYesterday bool
					for _, actualRecord := range resp.Msg.BodyRecords {
						if actualRecord.Date == today.Format("2006-01-02") {
							require.NotNil(t, actualRecord.WeightKg, "Today's weight should not be nil")
							assert.Equal(t, weight1, actualRecord.WeightKg.Value, "Today's weight mismatch")
							assert.Nil(t, actualRecord.BodyFatPercentage, "Today's body fat should be nil")
							foundToday = true
						} else if actualRecord.Date == yesterday.Format("2006-01-02") {
							require.NotNil(t, actualRecord.WeightKg, "Yesterday's weight should not be nil")
							assert.Equal(t, weight2, actualRecord.WeightKg.Value, "Yesterday's weight mismatch")
							require.NotNil(t, actualRecord.BodyFatPercentage, "Yesterday's body fat should not be nil")
							assert.Equal(t, bodyFat, actualRecord.BodyFatPercentage.Value, "Yesterday's body fat mismatch")
							foundYesterday = true
						}
					}
					if tc.name == "Default Pagination" {
						assert.True(t, foundToday, "Did not find today's record in default pagination")
						assert.True(t, foundYesterday, "Did not find yesterday's record in default pagination")
					}
				}
			}
		})
	}
}

func TestGetBodyRecordsByDateRange(t *testing.T) {
	resetDB(t, testPool)
	repo := repo.NewBodyRecordRepository(testPool)
	handler := NewBodyRecordHandler(repo, testLogger)
	ctx := context.Background()
	testCtx := newTestContext(ctx)

	// Setup: Create test records
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lastWeek := today.Add(-7 * 24 * time.Hour)
	weight1, weight2, weight3 := 75.5, 76.0, 77.0
	bodyFat := 15.5
	recordToday, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, today, &weight1, nil)
	require.NoError(t, err)
	recordYesterday, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, yesterday, &weight2, &bodyFat)
	require.NoError(t, err)
	recordLastWeek, err := testutil.CreateTestBodyRecord(ctx, testQueries, testUserID, lastWeek, &weight3, nil)
	require.NoError(t, err)

	// Convert to proto
	protoToday := ToProtoBodyRecord(recordToday)
	protoYesterday := ToProtoBodyRecord(recordYesterday)
	protoLastWeek := ToProtoBodyRecord(recordLastWeek)

	testCases := []struct {
		name         string
		startDate    time.Time
		endDate      time.Time
		expectError  bool
		expectedResp *v1.GetBodyRecordsByDateRangeResponse
	}{
		{
			name:        "Range including today and yesterday",
			startDate:   today.Add(-3 * 24 * time.Hour),
			endDate:     today,
			expectError: false,
			expectedResp: &v1.GetBodyRecordsByDateRangeResponse{
				BodyRecords: []*v1.BodyRecord{protoToday, protoYesterday},
			},
		},
		{
			name:        "Range including only today",
			startDate:   today,
			endDate:     today,
			expectError: false,
			expectedResp: &v1.GetBodyRecordsByDateRangeResponse{
				BodyRecords: []*v1.BodyRecord{protoToday},
			},
		},
		{
			name:        "Range including all three",
			startDate:   lastWeek,
			endDate:     today,
			expectError: false,
			expectedResp: &v1.GetBodyRecordsByDateRangeResponse{
				BodyRecords: []*v1.BodyRecord{protoToday, protoYesterday, protoLastWeek},
			},
		},
		{
			name:        "Range with no records",
			startDate:   today.Add(24 * time.Hour),
			endDate:     today.Add(48 * time.Hour),
			expectError: false,
			expectedResp: &v1.GetBodyRecordsByDateRangeResponse{
				BodyRecords: []*v1.BodyRecord{},
			},
		},
		{
			name:        "Range including only last week",
			startDate:   lastWeek,
			endDate:     lastWeek,
			expectError: false,
			expectedResp: &v1.GetBodyRecordsByDateRangeResponse{
				BodyRecords: []*v1.BodyRecord{protoLastWeek},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&v1.GetBodyRecordsByDateRangeRequest{
				StartDate: tc.startDate.Format("2006-01-02"),
				EndDate:   tc.endDate.Format("2006-01-02"),
			})
			resp, err := handler.GetBodyRecordsByDateRange(testCtx, req)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)

				assert.Len(t, resp.Msg.BodyRecords, len(tc.expectedResp.BodyRecords), "Incorrect number of body records returned for date range")

				expectedDates := make(map[string]bool)
				for _, rec := range tc.expectedResp.BodyRecords {
					expectedDates[rec.Date] = true
				}

				foundDates := make(map[string]bool)
				for _, actualRecord := range resp.Msg.BodyRecords {
					foundDates[actualRecord.Date] = true
					switch actualRecord.Date {
					case today.Format("2006-01-02"):
						require.NotNil(t, actualRecord.WeightKg)
						assert.Equal(t, weight1, actualRecord.WeightKg.Value)
						assert.Nil(t, actualRecord.BodyFatPercentage)
					case yesterday.Format("2006-01-02"):
						require.NotNil(t, actualRecord.WeightKg)
						assert.Equal(t, weight2, actualRecord.WeightKg.Value)
						require.NotNil(t, actualRecord.BodyFatPercentage)
						assert.Equal(t, bodyFat, actualRecord.BodyFatPercentage.Value)
					case lastWeek.Format("2006-01-02"):
						require.NotNil(t, actualRecord.WeightKg)
						assert.Equal(t, weight3, actualRecord.WeightKg.Value)
						assert.Nil(t, actualRecord.BodyFatPercentage)
					}
				}

				for date := range expectedDates {
					assert.True(t, foundDates[date], "Expected record for date %s not found", date)
				}
				for date := range foundDates {
					assert.True(t, expectedDates[date], "Found unexpected record for date %s", date)
				}
			}
		})
	}
}
