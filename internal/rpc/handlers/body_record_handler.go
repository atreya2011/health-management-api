package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	// "github.com/atreya2011/health-management-api/internal/application" // Removed
	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	"github.com/atreya2011/health-management-api/internal/auth"
	postgres "github.com/atreya2011/health-management-api/internal/db" // Added
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// BodyRecordHandler implements the body record service RPCs
type BodyRecordHandler struct {
	repo *postgres.PgBodyRecordRepository // Use concrete repository type
	log  *slog.Logger
}

// NewBodyRecordHandler creates a new body record handler
func NewBodyRecordHandler(repo *postgres.PgBodyRecordRepository, log *slog.Logger) *BodyRecordHandler { // Use concrete repository type
	return &BodyRecordHandler{
		repo: repo,
		log:  log,
	}
}

// CreateBodyRecord creates or updates a body record for a specific date
func (h *BodyRecordHandler) CreateBodyRecord(ctx context.Context, req *connect.Request[v1.CreateBodyRecordRequest]) (*connect.Response[v1.CreateBodyRecordResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Msg.Date)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid date format", "date", req.Msg.Date, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid date format: %w", err))
	}

	// Convert protobuf wrappers to Go pointers
	var weight *float64
	var bodyFat *float64

	if req.Msg.WeightKg != nil {
		w := req.Msg.WeightKg.Value
		weight = &w
	}

	if req.Msg.BodyFatPercentage != nil {
		bf := req.Msg.BodyFatPercentage.Value
		bodyFat = &bf
	}

	// Construct persistence object (was domain object)
	record := &postgres.BodyRecord{ // Use postgres.BodyRecord
		UserID:            userID,
		Date:              date,
		WeightKg:          weight,
		BodyFatPercentage: bodyFat,
	}

	// Validate the record (moved from service)
	if err := record.Validate(); err != nil {
		h.log.WarnContext(ctx, "Validation failed for body record", "userID", userID, "error", err)
		// Use CodeInvalidArgument for validation errors
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid body record data: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Saving body record", "userID", userID, "date", date)
	savedRecord, err := h.repo.Save(ctx, record) // Changed from bodyRecordApp.CreateOrUpdateBodyRecord
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to save body record", "userID", userID, "error", err)
		// Use CodeInternal for persistence errors
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to save body record"))
	}

	// Convert persistence model to protobuf message
	protoRecord := toProtoBodyRecord(savedRecord) // Use savedRecord (now *postgres.BodyRecord)

	// Create response
	res := connect.NewResponse(&v1.CreateBodyRecordResponse{
		BodyRecord: protoRecord,
	})

	return res, nil
}

// ListBodyRecords lists body records for the authenticated user
func (h *BodyRecordHandler) ListBodyRecords(ctx context.Context, req *connect.Request[v1.ListBodyRecordsRequest]) (*connect.Response[v1.ListBodyRecordsResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Get pagination parameters & apply defaults (moved from service)
	pageSize := 20  // Default page size
	pageNumber := 1 // Default page number

	if req.Msg.Pagination != nil {
		if req.Msg.Pagination.PageSize > 0 {
			pageSize = int(req.Msg.Pagination.PageSize)
			// Apply max page size (from service)
			if pageSize > 100 {
				pageSize = 100
			}
		}
		if req.Msg.Pagination.PageNumber > 0 {
			pageNumber = int(req.Msg.Pagination.PageNumber)
		}
	}
	// Ensure page is at least 1 (from service)
	if pageNumber <= 0 {
		pageNumber = 1
	}

	// Calculate offset (from service)
	offset := (pageNumber - 1) * pageSize

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching body records for user", "userID", userID, "page", pageNumber, "pageSize", pageSize)
	records, err := h.repo.FindByUser(ctx, userID, pageSize, offset) // Changed from bodyRecordApp.GetBodyRecordsForUser
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch body records", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch body records"))
	}

	// Get total count (from service)
	total, err := h.repo.CountByUser(ctx, userID)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count body records", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count body records"))
	}

	// Convert persistence models to protobuf messages
	protoRecords := make([]*v1.BodyRecord, len(records)) // records is now []*postgres.BodyRecord
	for i, record := range records {
		protoRecords[i] = toProtoBodyRecord(record) // Pass *postgres.BodyRecord
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListBodyRecordsResponse{
		BodyRecords: protoRecords,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// GetBodyRecordsByDateRange retrieves body records for a specific date range
func (h *BodyRecordHandler) GetBodyRecordsByDateRange(ctx context.Context, req *connect.Request[v1.GetBodyRecordsByDateRangeRequest]) (*connect.Response[v1.GetBodyRecordsByDateRangeResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.Msg.StartDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid start date format", "startDate", req.Msg.StartDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid start date format: %w", err))
	}

	endDate, err := time.Parse("2006-01-02", req.Msg.EndDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid end date format", "endDate", req.Msg.EndDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid end date format: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching body records for user by date range", "userID", userID, "startDate", startDate, "endDate", endDate)
	records, err := h.repo.FindByUserAndDateRange(ctx, userID, startDate, endDate) // Changed from bodyRecordApp.GetBodyRecordsForUserDateRange
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch body records by date range", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch body records by date range"))
	}

	// Convert persistence models to protobuf messages
	protoRecords := make([]*v1.BodyRecord, len(records)) // records is now []*postgres.BodyRecord
	for i, record := range records {
		protoRecords[i] = toProtoBodyRecord(record) // Pass *postgres.BodyRecord
	}

	// Create response
	res := connect.NewResponse(&v1.GetBodyRecordsByDateRangeResponse{
		BodyRecords: protoRecords,
	})

	return res, nil
}

// toProtoBodyRecord converts a postgres.BodyRecord to a v1.BodyRecord
func toProtoBodyRecord(record *postgres.BodyRecord) *v1.BodyRecord { // Accept *postgres.BodyRecord
	protoRecord := &v1.BodyRecord{
		Id:        record.ID.String(),
		UserId:    record.UserID.String(),
		Date:      record.Date.Format("2006-01-02"),
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}

	// Handle optional fields
	if record.WeightKg != nil {
		protoRecord.WeightKg = &wrapperspb.DoubleValue{Value: *record.WeightKg}
	}

	if record.BodyFatPercentage != nil {
		protoRecord.BodyFatPercentage = &wrapperspb.DoubleValue{Value: *record.BodyFatPercentage}
	}

	return protoRecord
}
