package repository

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	SaveSummary(ctx context.Context, summary *domain.Summary) error
	ListSummaries(ctx context.Context) ([]domain.SummaryOverview, error)
	GetSummary(ctx context.Context, id uuid.UUID) (*domain.Summary, error)
}
