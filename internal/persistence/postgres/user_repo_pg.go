package postgres

import (
	"context"
	"errors"
	"fmt"
	"time" // Added for User struct

	// "github.com/atreya2011/health-management-api/internal/domain" // Removed
	db "github.com/atreya2011/health-management-api/internal/persistence/postgres/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUserNotFound is returned when a user is not found (Moved from domain)
var ErrUserNotFound = errors.New("user not found")

// User represents a user in the system (Moved from domain)
type User struct {
	ID        uuid.UUID
	SubjectID string // Subject identifier from JWT, keep internal, don't expose in API directly
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PgUserRepository provides database operations for User (Exported)
type PgUserRepository struct { // Renamed to export
	q *db.Queries
}

// NewPgUserRepository creates a new PostgreSQL user repository
func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository { // Return exported type
	return &PgUserRepository{ // Use exported type
		q: db.New(pool),
	}
}

// Create creates a new user record
func (r *PgUserRepository) Create(ctx context.Context, user *User) error { // Use local User
	dbUser, err := r.q.CreateUser(ctx, user.SubjectID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return fmt.Errorf("user with this subject ID already exists: %w", err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt

	return nil
}

// FindByID retrieves a user by their internal UUID
func (r *PgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) { // Use local User
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound // Use local error
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return toLocalUser(dbUser), nil // Use local conversion func
}

// FindBySubjectID retrieves a user by their JWT Subject claim
func (r *PgUserRepository) FindBySubjectID(ctx context.Context, subjectID string) (*User, error) { // Use local User
	dbUser, err := r.q.GetUserBySubjectID(ctx, subjectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound // Use local error
		}
		return nil, fmt.Errorf("failed to get user by subject ID: %w", err)
	}
	return toLocalUser(dbUser), nil // Use local conversion func
}

// toLocalUser converts a db.User (sqlc-generated) to a local User
func toLocalUser(dbUser db.User) *User { // Return local User
	return &User{ // Use local User
		ID:        dbUser.ID,
		SubjectID: dbUser.SubjectID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}
