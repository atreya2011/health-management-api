package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/atreya2011/health-management-api/internal/application"
	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/atreya2011/health-management-api/internal/infrastructure/auth"
	v1 "github.com/atreya2011/health-management-api/internal/infrastructure/rpc/gen/healthapp/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// DiaryHandler implements the diary service
type DiaryHandler struct {
	diaryApp application.DiaryService
	log      *slog.Logger
}

// NewDiaryHandler creates a new diary handler
func NewDiaryHandler(diaryApp application.DiaryService, log *slog.Logger) *DiaryHandler {
	return &DiaryHandler{
		diaryApp: diaryApp,
		log:      log,
	}
}

// CreateDiaryEntry creates a new diary entry
func (h *DiaryHandler) CreateDiaryEntry(ctx context.Context, req *connect.Request[v1.CreateDiaryEntryRequest]) (*connect.Response[v1.CreateDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse date
	entryDate, err := time.Parse("2006-01-02", req.Msg.EntryDate)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid date format", "entryDate", req.Msg.EntryDate, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid date format: %w", err))
	}

	// Convert protobuf wrapper to Go pointer
	var title *string
	if req.Msg.Title != nil {
		t := req.Msg.Title.Value
		title = &t
	}

	// Call application service
	h.log.InfoContext(ctx, "Creating diary entry", "userID", userID, "entryDate", entryDate)
	entry, err := h.diaryApp.CreateDiaryEntry(ctx, userID, title, req.Msg.Content, entryDate)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to create diary entry", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create diary entry"))
	}

	// Convert domain model to protobuf message
	protoEntry := toProtoDiaryEntry(entry)

	// Create response
	res := connect.NewResponse(&v1.CreateDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// UpdateDiaryEntry updates an existing diary entry
func (h *DiaryHandler) UpdateDiaryEntry(ctx context.Context, req *connect.Request[v1.UpdateDiaryEntryRequest]) (*connect.Response[v1.UpdateDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Convert protobuf wrapper to Go pointer
	var title *string
	if req.Msg.Title != nil {
		t := req.Msg.Title.Value
		title = &t
	}

	// Call application service
	h.log.InfoContext(ctx, "Updating diary entry", "entryID", entryID, "userID", userID)
	entry, err := h.diaryApp.UpdateDiaryEntry(ctx, entryID, userID, title, req.Msg.Content)
	if err != nil {
		if errors.Is(err, domain.ErrDiaryEntryNotFound) {
			h.log.WarnContext(ctx, "Diary entry not found", "entryID", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		}
		h.log.ErrorContext(ctx, "Failed to update diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to update diary entry"))
	}

	// Convert domain model to protobuf message
	protoEntry := toProtoDiaryEntry(entry)

	// Create response
	res := connect.NewResponse(&v1.UpdateDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// ListDiaryEntries lists diary entries for the authenticated user
func (h *DiaryHandler) ListDiaryEntries(ctx context.Context, req *connect.Request[v1.ListDiaryEntriesRequest]) (*connect.Response[v1.ListDiaryEntriesResponse], error) {
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

	// Call application service
	h.log.InfoContext(ctx, "Listing diary entries", "userID", userID, "pageSize", pageSize, "pageNumber", pageNumber)
	entries, total, err := h.diaryApp.ListDiaryEntries(ctx, userID, pageNumber, pageSize)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to list diary entries", "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to list diary entries"))
	}

	// Convert domain models to protobuf messages
	protoEntries := make([]*v1.DiaryEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = toProtoDiaryEntry(entry)
	}

	// Calculate pagination response
	totalPages := (int(total) + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Create response
	res := connect.NewResponse(&v1.ListDiaryEntriesResponse{
		DiaryEntries: protoEntries,
		Pagination: &v1.PageResponse{
			TotalItems:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(pageNumber),
		},
	})

	return res, nil
}

// GetDiaryEntry retrieves a specific diary entry by ID
func (h *DiaryHandler) GetDiaryEntry(ctx context.Context, req *connect.Request[v1.GetDiaryEntryRequest]) (*connect.Response[v1.GetDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Call application service
	h.log.InfoContext(ctx, "Getting diary entry", "entryID", entryID, "userID", userID)
	entry, err := h.diaryApp.GetDiaryEntry(ctx, entryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrDiaryEntryNotFound) {
			h.log.WarnContext(ctx, "Diary entry not found", "entryID", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		}
		h.log.ErrorContext(ctx, "Failed to get diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to get diary entry"))
	}

	// Convert domain model to protobuf message
	protoEntry := toProtoDiaryEntry(entry)

	// Create response
	res := connect.NewResponse(&v1.GetDiaryEntryResponse{
		DiaryEntry: protoEntry,
	})

	return res, nil
}

// DeleteDiaryEntry deletes a diary entry
func (h *DiaryHandler) DeleteDiaryEntry(ctx context.Context, req *connect.Request[v1.DeleteDiaryEntryRequest]) (*connect.Response[v1.DeleteDiaryEntryResponse], error) {
	// Get user ID from context
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		h.log.ErrorContext(ctx, "User ID not found in context")
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}

	// Parse entry ID
	entryID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		h.log.WarnContext(ctx, "Invalid entry ID", "entryID", req.Msg.Id, "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid entry ID: %w", err))
	}

	// Call application service
	h.log.InfoContext(ctx, "Deleting diary entry", "entryID", entryID, "userID", userID)
	err = h.diaryApp.DeleteDiaryEntry(ctx, entryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrDiaryEntryNotFound) {
			h.log.WarnContext(ctx, "Diary entry not found for deletion", "entryID", entryID, "userID", userID)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("diary entry not found"))
		}
		h.log.ErrorContext(ctx, "Failed to delete diary entry", "entryID", entryID, "userID", userID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to delete diary entry"))
	}

	// Create response
	res := connect.NewResponse(&v1.DeleteDiaryEntryResponse{
		Success: true,
	})

	return res, nil
}

// toProtoDiaryEntry converts a domain.DiaryEntry to a v1.DiaryEntry
func toProtoDiaryEntry(entry *domain.DiaryEntry) *v1.DiaryEntry {
	protoEntry := &v1.DiaryEntry{
		Id:        entry.ID.String(),
		UserId:    entry.UserID.String(),
		Content:   entry.Content,
		EntryDate: entry.EntryDate.Format("2006-01-02"),
		CreatedAt: timestamppb.New(entry.CreatedAt),
		UpdatedAt: timestamppb.New(entry.UpdatedAt),
	}

	// Handle optional fields
	if entry.Title != nil {
		protoEntry.Title = &wrapperspb.StringValue{Value: *entry.Title}
	}

	return protoEntry
}
