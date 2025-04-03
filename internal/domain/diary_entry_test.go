package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDiaryEntry_Validate(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	
	// Helper function to create a valid diary entry
	createValidEntry := func() *DiaryEntry {
		title := "Test Title"
		return &DiaryEntry{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Title:     &title,
			Content:   "This is a valid content for testing.",
			EntryDate: now,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	tests := []struct {
		name    string
		modify  func(*DiaryEntry)
		wantErr bool
	}{
		{
			name:    "Valid diary entry",
			modify:  func(de *DiaryEntry) {},
			wantErr: false,
		},
		{
			name: "Empty content",
			modify: func(de *DiaryEntry) {
				de.Content = ""
			},
			wantErr: true,
		},
		{
			name: "Whitespace content",
			modify: func(de *DiaryEntry) {
				de.Content = "   "
			},
			wantErr: true,
		},
		{
			name: "Content too long",
			modify: func(de *DiaryEntry) {
				// Create a string longer than 10000 characters
				de.Content = string(make([]byte, 10001))
			},
			wantErr: true,
		},
		{
			name: "Title too long",
			modify: func(de *DiaryEntry) {
				longTitle := string(make([]byte, 201))
				de.Title = &longTitle
			},
			wantErr: true,
		},
		{
			name: "Future entry date",
			modify: func(de *DiaryEntry) {
				de.EntryDate = tomorrow
			},
			wantErr: true,
		},
		{
			name: "Nil title is valid",
			modify: func(de *DiaryEntry) {
				de.Title = nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			de := createValidEntry()
			tt.modify(de)
			
			err := de.Validate()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("DiaryEntry.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
