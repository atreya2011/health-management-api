package log

import (
	"log/slog"
	"os"
)

// NewLogger creates a new structured logger
func NewLogger() *slog.Logger {
	// Create a JSON handler with default options
	handler := slog.NewJSONHandler(os.Stdout, nil)
	
	// Create a new logger with the handler
	logger := slog.New(handler)
	
	return logger
}
