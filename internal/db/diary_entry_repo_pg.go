package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDiaryEntryNotFound is returned when a diary entry is not found
var ErrDiaryEntryNotFound = errors.New("diary entry not found")

// PgDiaryEntryRepository provides database operations for DiaryEntry
type PgDiaryEntryRepository struct {
	q *db.Queries
}

// NewPgDiaryEntryRepository creates a new PostgreSQL diary entry repository
func NewPgDiaryEntryRepository(pool *pgxpool.Pool) *PgDiaryEntryRepository { // Return exported type
	return &PgDiaryEntryRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new diary entry
func (r *PgDiaryEntryRepository) Create(ctx context.Context, userID uuid.UUID, title *string, content string, entryDate time.Time) (db.DiaryEntry, error) {
	var titleVal pgtype.Text
	if title != nil {
		titleVal = pgtype.Text{String: *title, Valid: true}
	}

	pgDate := pgtype.Date{Time: entryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    userID,
		Title:     titleVal,
		Content:   content,
		EntryDate: pgDate,
	}

	dbEntry, err := r.q.CreateDiaryEntry(ctx, params)
	if err != nil {
		return db.DiaryEntry{}, fmt.Errorf("failed to create diary entry: %w", err)
	}

	// Return generated struct directly
	return dbEntry, nil
}

// Update updates an existing diary entry
func (r *PgDiaryEntryRepository) Update(ctx context.Context, id, userID uuid.UUID, title *string, content string) (db.DiaryEntry, error) {
	var titleVal pgtype.Text
	if title != nil {
		titleVal = pgtype.Text{String: *title, Valid: true}
	}

	params := db.UpdateDiaryEntryParams{
		ID:      id,
		Title:   titleVal,
		Content: content,
		UserID:  userID, // Need UserID to ensure user owns the entry being updated
	}

	dbEntry, err := r.q.UpdateDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DiaryEntry{}, ErrDiaryEntryNotFound // Return zero value and local error
		}
		return db.DiaryEntry{}, fmt.Errorf("failed to update diary entry: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbEntry, nil
}

// FindByID retrieves a diary entry by ID and user ID
func (r *PgDiaryEntryRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (db.DiaryEntry, error) {
	params := db.GetDiaryEntryByIDParams{
		ID:     id,
		UserID: userID,
	}

	dbEntry, err := r.q.GetDiaryEntryByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.DiaryEntry{}, ErrDiaryEntryNotFound // Return zero value and local error
		}
		return db.DiaryEntry{}, fmt.Errorf("failed to find diary entry: %w", err) // Use fmt.Errorf
	}

	// Return generated struct directly
	return dbEntry, nil
}

// FindByUser retrieves paginated diary entries for a user
func (r *PgDiaryEntryRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]db.DiaryEntry, error) {
	params := db.ListDiaryEntriesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbEntries, err := r.q.ListDiaryEntriesByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list diary entries: %w", err)
	}

	// Return generated structs directly
	return dbEntries, nil
}

// Delete deletes a diary entry by ID and user ID
func (r *PgDiaryEntryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	params := db.DeleteDiaryEntryParams{
		ID:     id,
		UserID: userID,
	}

	err := r.q.DeleteDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Consider returning ErrDiaryEntryNotFound here too for consistency?
			// For now, matching original behavior which returned nil on no rows for delete.
			return nil
		}
		return fmt.Errorf("failed to delete diary entry: %w", err) // Use fmt.Errorf
	}

	return nil
}

// CountByUser returns the total number of diary entries for a user
func (r *PgDiaryEntryRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q.CountDiaryEntriesByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count diary entries: %w", err)
	}

	return count, nil
}

// Removed toLocalDiaryEntry function as it's no longer needed
