package repository

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
)

const insertSummaryQuery = `
INSERT INTO summaries (
	id,
	external_summary_id,
	host,
	port,
	user_name,
	db_name,
	synced_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

const insertSchemaQuery = `
INSERT INTO schemas (
	id,
	summary_id,
	name
)
VALUES ($1, $2, $3)
`

const insertTableQuery = `
INSERT INTO tables (
	id,
	schema_id,
	name,
	row_count,
	size_mb
)
VALUES ($1, $2, $3, $4, $5)
`

func (r *PostgresRepository) SaveSummary(
	ctx context.Context,
	summary *domain.Summary,
) error {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		insertSummaryQuery,
		summary.ID,
		summary.ExternalSummaryID,
		summary.Host,
		summary.Port,
		summary.User,
		summary.DBName,
		summary.SyncedAt,
	)
	if err != nil {
		return err
	}

	for _, schema := range summary.Schemas {

		_, err = tx.Exec(
			ctx,
			insertSchemaQuery,
			schema.ID,
			summary.ID,
			schema.Name,
		)
		if err != nil {
			return err
		}

		for _, table := range schema.Tables {

			_, err = tx.Exec(
				ctx,
				insertTableQuery,
				table.ID,
				schema.ID,
				table.Name,
				table.RowCount,
				table.SizeMB,
			)
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
