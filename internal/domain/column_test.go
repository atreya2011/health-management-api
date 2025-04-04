package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestColumnValidate(t *testing.T) {
	tests := []struct {
		name    string
		column  Column
		wantErr bool
	}{
		{
			name: "valid column",
			column: Column{
				ID:      uuid.New(),
				Title:   "Test Title",
				Content: "Test Content",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			column: Column{
				ID:      uuid.New(),
				Title:   "",
				Content: "Test Content",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			column: Column{
				ID:      uuid.New(),
				Title:   "Test Title",
				Content: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.column.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Column.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestColumnIsPublished(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		column   Column
		expected bool
	}{
		{
			name: "published in the past",
			column: Column{
				ID:          uuid.New(),
				Title:       "Test Title",
				Content:     "Test Content",
				PublishedAt: &past,
			},
			expected: true,
		},
		{
			name: "published in the future",
			column: Column{
				ID:          uuid.New(),
				Title:       "Test Title",
				Content:     "Test Content",
				PublishedAt: &future,
			},
			expected: false,
		},
		{
			name: "not published",
			column: Column{
				ID:          uuid.New(),
				Title:       "Test Title",
				Content:     "Test Content",
				PublishedAt: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.column.IsPublished()
			if result != tt.expected {
				t.Errorf("Column.IsPublished() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
