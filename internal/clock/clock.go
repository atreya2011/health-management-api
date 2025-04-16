package clock

import "time"

// Clock provides an interface for time operations.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the standard time package.
type RealClock struct{}

// Now returns the current time using time.Now().
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// NewRealClock creates a new RealClock.
func NewRealClock() Clock {
	return &RealClock{}
}

// MockClock implements Clock with a settable time for testing.
type MockClock struct {
	mockTime time.Time
}

// Now returns the mock time.
// Returns UTC by default for consistency, matching many existing usages.
// Tests can override with SetTime if specific zones are needed.
func (c *MockClock) Now() time.Time {
	if c.mockTime.IsZero() {
		// Provide a default non-zero time if not set, to avoid potential issues.
		// Using a fixed date helps make tests more deterministic if time isn't explicitly set.
		return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return c.mockTime
}

// SetTime sets the time for the MockClock.
func (c *MockClock) SetTime(t time.Time) {
	c.mockTime = t
}

// NewMockClock creates a new MockClock initialized with the given time.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{mockTime: t}
}

// NewDefaultMockClock creates a MockClock with a default fixed time.
func NewDefaultMockClock() *MockClock {
	// Using a fixed date helps make tests more deterministic.
	return NewMockClock(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
}
