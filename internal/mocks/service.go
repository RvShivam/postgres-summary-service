package mocks

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockService is a testify mock for service.Service.
type MockService struct {
	mock.Mock
}

func (m *MockService) SyncSummary(
	ctx context.Context,
	req service.SyncRequest,
) (*domain.Summary, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Summary), args.Error(1)
}

func (m *MockService) ListSummaries(ctx context.Context) ([]domain.SummaryOverview, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SummaryOverview), args.Error(1)
}

func (m *MockService) GetSummary(ctx context.Context, id uuid.UUID) (*domain.Summary, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Summary), args.Error(1)
}
