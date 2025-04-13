package postgres

import (
	"context"
	"errors" // Use standard errors
	"fmt"
	"strings" // Added for validation
	"time"

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	db "github.com/atreya2011/health-management-api/internal/infrastructure/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pkg/errors" // Removed, use fmt.Errorf with %w
)

// ErrDiaryEntryNotFound is returned when a diary entry is not found (Moved from domain)
var ErrDiaryEntryNotFound = errors.New("diary entry not found")

// DiaryEntry represents a diary entry for a user (Moved from domain)
type DiaryEntry struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     *string
	Content   string
	EntryDate time.Time // Store as time.Time (YYYY-MM-DD 00:00:00 UTC)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate performs validation on the diary entry (Moved from domain)
func (de *DiaryEntry) Validate() error {
	// Validate content (required)
	if strings.TrimSpace(de.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// Validate content length
	if len(de.Content) > 10000 {
		return errors.New("content exceeds maximum allowed length (10000 characters)")
	}

	// Validate title length (if provided)
	if de.Title != nil && len(*de.Title) > 200 {
		return errors.New("title exceeds maximum allowed length (200 characters)")
	}

	// Validate entry date is not in the future
	now := time.Now()
	if de.EntryDate.After(now) {
		return errors.New("entry date cannot be in the future")
	}

	return nil
}

// PgDiaryEntryRepository provides database operations for DiaryEntry (Exported)
type PgDiaryEntryRepository struct { // Renamed to export
	q *db.Queries
}

// NewPgDiaryEntryRepository creates a new PostgreSQL diary entry repository
func NewPgDiaryEntryRepository(pool *pgxpool.Pool) *PgDiaryEntryRepository { // Return exported type
	return &PgDiaryEntryRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new diary entry
func (r *PgDiaryEntryRepository) Create(ctx context.Context, entry *DiaryEntry) (*DiaryEntry, error) { // Use local DiaryEntry
	var titleVal pgtype.Text
	if entry.Title != nil {
		titleVal = pgtype.Text{String: *entry.Title, Valid: true}
	}

	pgDate := pgtype.Date{Time: entry.EntryDate, Valid: true}

	params := db.CreateDiaryEntryParams{
		UserID:    entry.UserID,
		Title:     titleVal,
		Content:   entry.Content,
		EntryDate: pgDate,
	}

	dbEntry, err := r.q.CreateDiaryEntry(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create diary entry: %w", err)
	}

	return toLocalDiaryEntry(dbEntry), nil // Use local conversion func
}

// Update updates an existing diary entry
func (r *PgDiaryEntryRepository) Update(ctx context.Context, entry *DiaryEntry) (*DiaryEntry, error) { // Use local DiaryEntry
	var titleVal pgtype.Text
	if entry.Title != nil {
		titleVal = pgtype.Text{String: *entry.Title, Valid: true}
	}

	params := db.UpdateDiaryEntryParams{
		ID:      entry.ID,
		Title:   titleVal,
		Content: entry.Content,
		UserID:  entry.UserID,
	}

	dbEntry, err := r.q.UpdateDiaryEntry(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDiaryEntryNotFound // Use local error
		}
		return nil, fmt.Errorf("failed to update diary entry: %w", err) // Use fmt.Errorf
	}

	return toLocalDiaryEntry(dbEntry), nil // Use local conversion func
}

// FindByID retrieves a diary entry by ID and user ID
func (r *PgDiaryEntryRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*DiaryEntry, error) { // Use local DiaryEntry
	params := db.GetDiaryEntryByIDParams{
		ID:     id,
		UserID: userID,
	}

	dbEntry, err := r.q.GetDiaryEntryByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDiaryEntryNotFound // Use local error
		}
		return nil, fmt.Errorf("failed to find diary entry: %w", err) // Use fmt.Errorf
	}

	return toLocalDiaryEntry(dbEntry), nil // Use local conversion func
}

// FindByUser retrieves paginated diary entries for a user
func (r *PgDiaryEntryRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*DiaryEntry, error) { // Use local DiaryEntry slice
	params := db.ListDiaryEntriesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbEntries, err := r.q.ListDiaryEntriesByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list diary entries: %w", err)
	}

	entries := make([]*DiaryEntry, len(dbEntries)) // Use local DiaryEntry slice
	for i, dbEntry := range dbEntries {
		entries[i] = toLocalDiaryEntry(dbEntry) // Use local conversion func
	}

	return entries, nil
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

// toLocalDiaryEntry converts a db.DiaryEntry (sqlc-generated) to a local DiaryEntry
func toLocalDiaryEntry(dbEntry db.DiaryEntry) *DiaryEntry { // Return local DiaryEntry
	var title *string
	if dbEntry.Title.Valid {
		title = &dbEntry.Title.String
	}

	var entryDateVal time.Time
	if dbEntry.EntryDate.Valid {
		entryDateVal = dbEntry.EntryDate.Time
	}

	return &DiaryEntry{ // Use local DiaryEntry
		ID:        dbEntry.ID,
		UserID:    dbEntry.UserID,
		Title:     title,
		Content:   dbEntry.Content,
		EntryDate: entryDateVal,
		CreatedAt: dbEntry.CreatedAt,
		UpdatedAt: dbEntry.UpdatedAt,
	}
}
