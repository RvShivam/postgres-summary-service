package service

import (
	"context"
	"time"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/google/uuid"
)

type SyncRequest struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func (s *SummaryService) SyncSummary(
	ctx context.Context,
	req SyncRequest,
) (*domain.Summary, error) {

	// Build request for the external service.
	externalReq := external.SummaryRequest{
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		DBName:   req.DBName,
	}

	// Fetch the database summary.
	externalResp, err := s.client.FetchSummary(ctx, externalReq)
	if err != nil {
		return nil, err
	}

	// Build the aggregate root.
	summary := &domain.Summary{
		ID:                uuid.New(),
		ExternalSummaryID: externalResp.SummaryID,
		Host:              req.Host,
		Port:              req.Port,
		User:              req.User,
		DBName:            req.DBName,
		SyncedAt:          time.Now().UTC(),
		Schemas:           make([]domain.Schema, 0, len(externalResp.Schemas)),
	}

	// Map schemas and tables.
	for _, extSchema := range externalResp.Schemas {

		schema := domain.Schema{
			ID:        uuid.New(),
			SummaryID: summary.ID,
			Name:      extSchema.Name,
			Tables:    make([]domain.Table, 0, len(extSchema.Tables)),
		}

		for _, extTable := range extSchema.Tables {

			table := domain.Table{
				ID:       uuid.New(),
				SchemaID: schema.ID,
				Name:     extTable.Name,
				RowCount: extTable.RowCount,
				SizeMB:   extTable.SizeMB,
			}

			schema.Tables = append(schema.Tables, table)
		}

		summary.Schemas = append(summary.Schemas, schema)
	}

	// Persist the aggregate.
	if err := s.repository.SaveSummary(ctx, summary); err != nil {
		return nil, err
	}

	return summary, nil
}
