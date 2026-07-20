package mocks

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a testify mock for repository.Repository.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveSummary(ctx context.Context, summary *domain.Summary) error {
	args := m.Called(ctx, summary)
	return args.Error(0)
}

func (m *MockRepository) ListSummaries(ctx context.Context) ([]domain.SummaryOverview, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SummaryOverview), args.Error(1)
}

func (m *MockRepository) GetSummary(ctx context.Context, id uuid.UUID) (*domain.Summary, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Summary), args.Error(1)
}
