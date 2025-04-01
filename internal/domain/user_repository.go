package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user record
	Create(ctx context.Context, user *User) error
	
	// FindByID retrieves a user by their internal UUID
	// Returns ErrUserNotFound if not found
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	
	// FindBySubjectID retrieves a user by their JWT Subject claim
	// Returns ErrUserNotFound if not found
	FindBySubjectID(ctx context.Context, subjectID string) (*User, error)
}
