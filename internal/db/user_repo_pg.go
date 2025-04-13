package postgres

import (
	"context"
	"errors"
	"fmt"

	db "github.com/atreya2011/health-management-api/internal/db/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// PgUserRepository provides database operations for User
type PgUserRepository struct {
	q *db.Queries
}

// NewPgUserRepository creates a new PostgreSQL user repository
func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository { // Return exported type
	return &PgUserRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new user record
func (r *PgUserRepository) Create(ctx context.Context, subjectID string) (db.User, error) {
	dbUser, err := r.q.CreateUser(ctx, subjectID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			// If user already exists, try to fetch them instead of returning an error
			existingUser, findErr := r.FindBySubjectID(ctx, subjectID)
			if findErr != nil {
				// Return original creation error if fetching also fails
				return db.User{}, fmt.Errorf("user with subject ID %s already exists, but failed to retrieve: %w; original error: %w", subjectID, findErr, err)
			}
			return existingUser, nil // Return existing user
		}
		return db.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	// Return the newly created user directly
	return dbUser, nil
}

// FindByID retrieves a user by their internal UUID
func (r *PgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound // Return zero value and local error
		}
		return db.User{}, fmt.Errorf("failed to get user by ID: %w", err)
	}
	// Return generated struct directly
	return dbUser, nil
}

// FindBySubjectID retrieves a user by their JWT Subject claim
func (r *PgUserRepository) FindBySubjectID(ctx context.Context, subjectID string) (db.User, error) {
	dbUser, err := r.q.GetUserBySubjectID(ctx, subjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrUserNotFound // Return zero value and local error
		}
		return db.User{}, fmt.Errorf("failed to get user by subject ID: %w", err)
	}
	// Return generated struct directly
	return dbUser, nil
}

// Removed toLocalUser function as it's no longer needed
