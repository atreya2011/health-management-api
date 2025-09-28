package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/auth"
	"github.com/atreya2011/health-management-api/internal/clock"
	"github.com/atreya2011/health-management-api/internal/repo"
	db "github.com/atreya2011/health-management-api/internal/repo/gen"
	v1 "github.com/atreya2011/health-management-api/internal/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ExerciseRecordHandler implements the exercise record service RPCs
type ExerciseRecordHandler struct {
	repo  *repo.ExerciseRecordRepository // Use concrete repository type
	log   *slog.Logger
	clock clock.Clock
}

// NewExerciseRecordHandler creates a new exercise record handler
func NewExerciseRecordHandler(repo *repo.ExerciseRecordRepository, log *slog.Logger, clock clock.Clock) *ExerciseRecordHandler {
	return &ExerciseRecordHandler{
		repo:  repo,
		log:   log,
		clock: clock,
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
		recordedAt = h.clock.Now()
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

	// Re-implement validation logic here
	exerciseName := req.Msg.ExerciseName
	if exerciseName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("exercise name cannot be empty"))
	}
	if len(exerciseName) > 100 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("exercise name exceeds maximum allowed length (100 characters)"))
	}
	if durationMinutes != nil {
		duration := *durationMinutes
		if duration <= 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("duration must be positive"))
		}
		if duration > 1440 { // 24 hours in minutes
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("duration exceeds maximum allowed value (24 hours)"))
		}
	}
	if caloriesBurned != nil {
		calories := *caloriesBurned
		if calories < 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("calories burned cannot be negative"))
		}
		if calories > 10000 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("calories burned exceeds maximum allowed value"))
		}
	}
	if recordedAt.After(h.clock.Now()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("recorded date cannot be in the future"))
	}
	// Removed instantiation of repo.ExerciseRecord

	// Call repository directly with new signature, passing current time from clock
	now := h.clock.Now()
	h.log.InfoContext(ctx, "Creating exercise record", "userID", userID, "exerciseName", exerciseName, "now", now)
	savedRecord, err := h.repo.Create(ctx, userID, exerciseName, durationMinutes, caloriesBurned, recordedAt, now)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to create exercise record", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create exercise record"))
	}

	// Convert persistence model to protobuf message
	protoRecord := ToProtoExerciseRecord(savedRecord) // Use savedRecord (now db.ExerciseRecord)

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
	protoRecords := make([]*v1.ExerciseRecord, len(records)) // records is now []db.ExerciseRecord
	for i, record := range records {
		protoRecords[i] = ToProtoExerciseRecord(record) // Pass db.ExerciseRecord
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
	err = h.repo.Delete(ctx, recordID, userID)
	if err != nil {
		// Check if the error is ErrExerciseRecordNotFound from the repository
		if errors.Is(err, repo.ErrExerciseRecordNotFound) {
			h.log.WarnContext(ctx, "Exercise record not found or not owned by user during deletion", "recordID", recordID, "userID", userID)
			// Return NotFound error to the client
			return nil, connect.NewError(connect.CodeNotFound, errors.New("exercise record not found"))
		}
		// Handle other potential errors
		h.log.ErrorContext(ctx, "Failed to delete exercise record", "recordID", recordID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete exercise record"))
	}

	// Create response
	res := connect.NewResponse(&v1.DeleteExerciseRecordResponse{
		Success: true,
	})

	return res, nil
}

// ToProtoExerciseRecord converts a db.ExerciseRecord (sqlc generated) to a v1.ExerciseRecord
func ToProtoExerciseRecord(record db.ExerciseRecord) *v1.ExerciseRecord { // Accept db.ExerciseRecord
	protoRecord := &v1.ExerciseRecord{
		Id:           record.ID.String(),
		UserId:       record.UserID.String(),
		ExerciseName: record.ExerciseName,
		RecordedAt:   timestamppb.New(record.RecordedAt),
		CreatedAt:    timestamppb.New(record.CreatedAt),
		UpdatedAt:    timestamppb.New(record.UpdatedAt),
	}

	// Handle pgtype.Int4 for optional fields
	if record.DurationMinutes.Valid {
		protoRecord.DurationMinutes = &wrapperspb.Int32Value{Value: record.DurationMinutes.Int32}
	}

	if record.CaloriesBurned.Valid {
		protoRecord.CaloriesBurned = &wrapperspb.Int32Value{Value: record.CaloriesBurned.Int32}
	}

	return protoRecord
}
