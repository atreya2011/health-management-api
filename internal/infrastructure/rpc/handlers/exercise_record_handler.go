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
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	"github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres" // Added
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ExerciseRecordHandler implements the exercise record service RPCs
type ExerciseRecordHandler struct {
	repo *postgres.PgExerciseRecordRepository // Use concrete repository type
	log  *slog.Logger
}

// NewExerciseRecordHandler creates a new exercise record handler
func NewExerciseRecordHandler(repo *postgres.PgExerciseRecordRepository, log *slog.Logger) *ExerciseRecordHandler { // Use concrete repository type
	return &ExerciseRecordHandler{
		repo: repo,
		log:  log,
	}
}

// CreateExerciseRecord creates a new exercise record
func (h *ExerciseRecordHandler) CreateExerciseRecord(ctx context.Context, req *connect.Request[v1.CreateExerciseRecordRequest]) (*connect.Response[v1.CreateExerciseRecordResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Get recorded_at time, default to current time if not provided
	var recordedAt time.Time
	if req.Msg.RecordedAt != nil {
		recordedAt = req.Msg.RecordedAt.AsTime()
	} else {
		recordedAt = time.Now()
	}

	// Convert protobuf wrappers to Go pointers
	var durationMinutes *int32
	var caloriesBurned *int32

	if req.Msg.DurationMinutes != nil {
		d := req.Msg.DurationMinutes.Value
		durationMinutes = &d
	}

	if req.Msg.CaloriesBurned != nil {
		c := req.Msg.CaloriesBurned.Value
		caloriesBurned = &c
	}

	// Construct persistence object (was domain object)
	record := &postgres.ExerciseRecord{ // Use postgres.ExerciseRecord
		ID:              uuid.New(), // Generate ID here
		UserID:          userID,
		ExerciseName:    req.Msg.ExerciseName,
		DurationMinutes: durationMinutes,
		CaloriesBurned:  caloriesBurned,
		RecordedAt:      recordedAt,
		CreatedAt:       time.Now(), // Set timestamp here
		UpdatedAt:       time.Now(), // Set timestamp here
	}

	// Validate the record (moved from service)
	if err := record.Validate(); err != nil {
		h.log.WarnContext(ctx, "Validation failed for exercise record", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid exercise record data: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Creating exercise record", "userID", userID, "exerciseName", req.Msg.ExerciseName)
	savedRecord, err := h.repo.Create(ctx, record) // Changed from exerciseApp.CreateExerciseRecord
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to create exercise record", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create exercise record"))
	}

	// Convert persistence model to protobuf message
	protoRecord := toProtoExerciseRecord(savedRecord) // Use savedRecord (now *postgres.ExerciseRecord)

	// Create response
	res := connect.NewResponse(&v1.CreateExerciseRecordResponse{
		ExerciseRecord: protoRecord,
	})

	return res, nil
}

// ListExerciseRecords lists exercise records for the authenticated user
func (h *ExerciseRecordHandler) ListExerciseRecords(ctx context.Context, req *connect.Request[v1.ListExerciseRecordsRequest]) (*connect.Response[v1.ListExerciseRecordsResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Get pagination parameters
	pageSize := 20  // Default page size
	pageNumber := 1 // Default page number

	if req.Msg.Pagination != nil {
		if req.Msg.Pagination.PageSize > 0 {
			pageSize = int(req.Msg.Pagination.PageSize)
		}
		if req.Msg.Pagination.PageNumber > 0 {
			pageNumber = int(req.Msg.Pagination.PageNumber)
		}
	}
	// Ensure page is at least 1 (from service)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	// Apply max page size (from service)
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset (from service)
	offset := (pageNumber - 1) * pageSize

	// Call repository directly
	h.log.InfoContext(ctx, "Fetching exercise records for user", "userID", userID, "page", pageNumber, "pageSize", pageSize)
	records, err := h.repo.FindByUser(ctx, userID, pageSize, offset) // Changed from exerciseApp.ListExerciseRecords
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to fetch exercise records", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to fetch exercise records"))
	}

	// Get total count (from service)
	total, err := h.repo.CountByUser(ctx, userID)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to count exercise records", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to count exercise records"))
	}

	// Convert persistence models to protobuf messages
	protoRecords := make([]*v1.ExerciseRecord, len(records)) // records is now []*postgres.ExerciseRecord
	for i, record := range records {
		protoRecords[i] = toProtoExerciseRecord(record) // Pass *postgres.ExerciseRecord
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListExerciseRecordsResponse{
		ExerciseRecords: protoRecords,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// DeleteExerciseRecord deletes an exercise record
func (h *ExerciseRecordHandler) DeleteExerciseRecord(ctx context.Context, req *connect.Request[v1.DeleteExerciseRecordRequest]) (*connect.Response[v1.DeleteExerciseRecordResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse record ID
	recordID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid record ID", "recordID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid record ID: %w", err))
	}

	// Call repository directly
	h.log.InfoContext(ctx, "Deleting exercise record", "recordID", recordID, "userID", userID)
	err = h.repo.Delete(ctx, recordID, userID) // Changed from exerciseApp.DeleteExerciseRecord
	if err != nil {
		// Note: Delete doesn't return ErrExerciseRecordNotFound in the repo implementation currently
		// if errors.Is(err, postgres.ErrExerciseRecordNotFound) { // Use postgres error
		// 	h.log.WarnContext(ctx, "Exercise record not found for deletion", "recordID", recordID, "userID", userID)
		// 	return nil, connect.NewError(connect.CodeNotFound, errors.New("exercise record not found"))
		// }
		h.log.ErrorContext(ctx, "Failed to delete exercise record", "recordID", recordID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete exercise record"))
	}

	// Create response
	res := connect.NewResponse(&v1.DeleteExerciseRecordResponse{
		Success: true,
	})

	return res, nil
}

// toProtoExerciseRecord converts a postgres.ExerciseRecord to a v1.ExerciseRecord
func toProtoExerciseRecord(record *postgres.ExerciseRecord) *v1.ExerciseRecord { // Accept *postgres.ExerciseRecord
	protoRecord := &v1.ExerciseRecord{
		Id:           record.ID.String(),
		UserId:       record.UserID.String(),
		ExerciseName: record.ExerciseName,
		RecordedAt:   timestamppb.New(record.RecordedAt),
		CreatedAt:    timestamppb.New(record.CreatedAt),
		UpdatedAt:    timestamppb.New(record.UpdatedAt),
	}

	// Handle optional fields
	if record.DurationMinutes != nil {
		protoRecord.DurationMinutes = &wrapperspb.Int32Value{Value: *record.DurationMinutes}
	}

	if record.CaloriesBurned != nil {
		protoRecord.CaloriesBurned = &wrapperspb.Int32Value{Value: *record.CaloriesBurned}
	}

	return protoRecord
}
