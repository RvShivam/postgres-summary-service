package repository

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/google/uuid"
)

const getSummaryQuery = `
SELECT
	id,
	external_summary_id,
	host,
	port,
	user_name,
	db_name,
	synced_at
FROM summaries
WHERE id = $1
`

const getSchemasQuery = `
SELECT
	id,
	name
FROM schemas
WHERE summary_id = $1
ORDER BY name
`

const getTablesQuery = `
SELECT
	id,
	name,
	row_count,
	size_mb
FROM tables
WHERE schema_id = $1
ORDER BY name
`

func (r *PostgresRepository) GetSummary(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Summary, error) {

	var summary domain.Summary

	err := r.pool.QueryRow(ctx, getSummaryQuery, id).Scan(
		&summary.ID,
		&summary.ExternalSummaryID,
		&summary.Host,
		&summary.Port,
		&summary.User,
		&summary.DBName,
		&summary.SyncedAt,
	)
	if err != nil {
		return nil, err
	}

	schemaRows, err := r.pool.Query(ctx, getSchemasQuery, summary.ID)
	if err != nil {
		return nil, err
	}
	defer schemaRows.Close()

	for schemaRows.Next() {
		var schema domain.Schema

		err := schemaRows.Scan(
			&schema.ID,
			&schema.Name,
		)
		if err != nil {
			return nil, err
		}

		tableRows, err := r.pool.Query(ctx, getTablesQuery, schema.ID)
		if err != nil {
			return nil, err
		}

		for tableRows.Next() {
			var table domain.Table

			err := tableRows.Scan(
				&table.ID,
				&table.Name,
				&table.RowCount,
				&table.SizeMB,
			)
			if err != nil {
				tableRows.Close()
				return nil, err
			}

			schema.Tables = append(schema.Tables, table)
		}

		if err := tableRows.Err(); err != nil {
			tableRows.Close()
			return nil, err
		}

		tableRows.Close()

		summary.Schemas = append(summary.Schemas, schema)
	}

	if err := schemaRows.Err(); err != nil {
		return nil, err
	}

	return &summary, nil
}
