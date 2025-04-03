package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestExerciseRecord_Validate(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	
	// Helper function to create a valid exercise record
	createValidRecord := func() *ExerciseRecord {
		duration := int32(30)
		calories := int32(250)
		return &ExerciseRecord{
			ID:              uuid.New(),
			UserID:          uuid.New(),
			ExerciseName:    "Running",
			DurationMinutes: &duration,
			CaloriesBurned:  &calories,
			RecordedAt:      now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}

	tests := []struct {
		name    string
		modify  func(*ExerciseRecord)
		wantErr bool
	}{
		{
			name:    "Valid exercise record",
			modify:  func(er *ExerciseRecord) {},
			wantErr: false,
		},
		{
			name: "Empty exercise name",
			modify: func(er *ExerciseRecord) {
				er.ExerciseName = ""
			},
			wantErr: true,
		},
		{
			name: "Exercise name too long",
			modify: func(er *ExerciseRecord) {
				// Create a string longer than 100 characters
				er.ExerciseName = string(make([]byte, 101))
			},
			wantErr: true,
		},
		{
			name: "Negative duration",
			modify: func(er *ExerciseRecord) {
				negativeDuration := int32(-10)
				er.DurationMinutes = &negativeDuration
			},
			wantErr: true,
		},
		{
			name: "Zero duration",
			modify: func(er *ExerciseRecord) {
				zeroDuration := int32(0)
				er.DurationMinutes = &zeroDuration
			},
			wantErr: true,
		},
		{
			name: "Duration too long",
			modify: func(er *ExerciseRecord) {
				longDuration := int32(1500) // 25 hours
				er.DurationMinutes = &longDuration
			},
			wantErr: true,
		},
		{
			name: "Negative calories",
			modify: func(er *ExerciseRecord) {
				negativeCalories := int32(-100)
				er.CaloriesBurned = &negativeCalories
			},
			wantErr: true,
		},
		{
			name: "Calories too high",
			modify: func(er *ExerciseRecord) {
				highCalories := int32(12000)
				er.CaloriesBurned = &highCalories
			},
			wantErr: true,
		},
		{
			name: "Future recorded date",
			modify: func(er *ExerciseRecord) {
				er.RecordedAt = tomorrow
			},
			wantErr: true,
		},
		{
			name: "Nil duration is valid",
			modify: func(er *ExerciseRecord) {
				er.DurationMinutes = nil
			},
			wantErr: false,
		},
		{
			name: "Nil calories is valid",
			modify: func(er *ExerciseRecord) {
				er.CaloriesBurned = nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			er := createValidRecord()
			tt.modify(er)
			
			err := er.Validate()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ExerciseRecord.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
