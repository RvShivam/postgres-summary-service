package service

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/google/uuid"
)

func (s *SummaryService) GetSummary(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Summary, error) {
	return s.repository.GetSummary(ctx, id)
}
