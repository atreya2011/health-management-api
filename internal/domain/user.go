package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID
	SubjectID string // Subject identifier from JWT, keep internal, don't expose in API directly
	CreatedAt time.Time
	UpdatedAt time.Time
}
