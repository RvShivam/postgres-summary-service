package service

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/RvShivam/postgres-summary-service/internal/repository"
	"github.com/google/uuid"
)

type Service interface {
	SyncSummary(
		ctx context.Context,
		req SyncRequest,
	) (*domain.Summary, error)

	ListSummaries(
		ctx context.Context,
	) ([]domain.SummaryOverview, error)

	GetSummary(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.Summary, error)
}

type SummaryService struct {
	repository repository.Repository
	client     external.Client
}

func New(
	repository repository.Repository,
	client external.Client,
) Service {
	return &SummaryService{
		repository: repository,
		client:     client,
	}
}
