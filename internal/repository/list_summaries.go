package repository

import (
	"context"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
)

const listSummariesQuery = `
SELECT
	id,
	external_summary_id,
	host,
	port,
	user_name,
	db_name,
	synced_at
FROM summaries
ORDER BY synced_at DESC
`

func (r *PostgresRepository) ListSummaries(
	ctx context.Context,
) ([]domain.SummaryOverview, error) {

	rows, err := r.pool.Query(ctx, listSummariesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.SummaryOverview

	for rows.Next() {
		var summary domain.SummaryOverview

		err := rows.Scan(
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

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}
