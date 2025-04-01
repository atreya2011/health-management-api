package mocks

import (
	"context"
	"time"

	"github.com/atreya2011/health-management-api/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockBodyRecordRepository is a mock implementation of domain.BodyRecordRepository
type MockBodyRecordRepository struct {
	mock.Mock
}

// Save mocks the Save method
func (m *MockBodyRecordRepository) Save(ctx context.Context, record *domain.BodyRecord) (*domain.BodyRecord, error) {
	args := m.Called(ctx, record)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*domain.BodyRecord), args.Error(1)
}

// FindByUser mocks the FindByUser method
func (m *MockBodyRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.BodyRecord, error) {
	args := m.Called(ctx, userID, limit, offset)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]*domain.BodyRecord), args.Error(1)
}

// FindByUserAndDateRange mocks the FindByUserAndDateRange method
func (m *MockBodyRecordRepository) FindByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.BodyRecord, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]*domain.BodyRecord), args.Error(1)
}

// CountByUser mocks the CountByUser method
func (m *MockBodyRecordRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
