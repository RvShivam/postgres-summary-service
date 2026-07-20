package mocks

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/stretchr/testify/mock"
)

// MockExternalClient is a testify mock for external.Client.
type MockExternalClient struct {
	mock.Mock
}

func (m *MockExternalClient) FetchSummary(
	ctx context.Context,
	req external.SummaryRequest,
) (*external.SummaryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*external.SummaryResponse), args.Error(1)
}
