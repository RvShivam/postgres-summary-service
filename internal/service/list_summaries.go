package service

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
)

func (s *SummaryService) ListSummaries(
	ctx context.Context,
) ([]domain.SummaryOverview, error) {
	return s.repository.ListSummaries(ctx)
}
