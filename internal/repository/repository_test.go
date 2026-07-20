package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/repository"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── SaveSummary ───────────────────────────────────────────────────────────────

func TestSaveSummary_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	summaryID := uuid.New()
	schemaID := uuid.New()
	tableID := uuid.New()
	now := time.Now().UTC()

	summary := &domain.Summary{
		ID:                summaryID,
		ExternalSummaryID: "sum123",
		Host:              "remote-db.example.com",
		Port:              5432,
		User:              "readonly",
		DBName:            "sample",
		SyncedAt:          now,
		Schemas: []domain.Schema{
			{
				ID:        schemaID,
				SummaryID: summaryID,
				Name:      "public",
				Tables: []domain.Table{
					{
						ID:       tableID,
						SchemaID: schemaID,
						Name:     "users",
						RowCount: 1240,
						SizeMB:   12.5,
					},
				},
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO summaries").
		WithArgs(summaryID, "sum123", "remote-db.example.com", 5432, "readonly", "sample", now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO schemas").
		WithArgs(schemaID, summaryID, "public").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO tables").
		WithArgs(tableID, schemaID, "users", int64(1240), 12.5).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = repo.SaveSummary(context.Background(), summary)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveSummary_BeginTransactionError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	mock.ExpectBegin().WillReturnError(errors.New("connection lost"))

	err = repo.SaveSummary(context.Background(), &domain.Summary{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")
}

func TestSaveSummary_InsertSummaryError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO summaries").
		WillReturnError(errors.New("unique constraint violation"))
	mock.ExpectRollback()

	summary := &domain.Summary{ID: uuid.New(), SyncedAt: time.Now()}
	err = repo.SaveSummary(context.Background(), summary)

	require.Error(t, err)
}

// ── ListSummaries ─────────────────────────────────────────────────────────────

func TestListSummaries_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	id1 := uuid.New()
	now := time.Now().UTC()

	rows := pgxmock.NewRows([]string{
		"id", "external_summary_id", "host", "port", "user_name", "db_name", "synced_at",
	}).AddRow(id1, "sum123", "host1", 5432, "user1", "db1", now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	result, err := repo.ListSummaries(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, id1, result[0].ID)
	assert.Equal(t, "sum123", result[0].ExternalSummaryID)
	assert.Equal(t, "host1", result[0].Host)
	assert.Equal(t, "db1", result[0].DBName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListSummaries_Empty(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	rows := pgxmock.NewRows([]string{
		"id", "external_summary_id", "host", "port", "user_name", "db_name", "synced_at",
	})
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	result, err := repo.ListSummaries(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListSummaries_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	result, err := repo.ListSummaries(context.Background())

	require.Error(t, err)
	assert.Nil(t, result)
}

// ── GetSummary ────────────────────────────────────────────────────────────────

func TestGetSummary_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	summaryID := uuid.New()
	schemaID := uuid.New()
	tableID := uuid.New()
	now := time.Now().UTC()

	// Expect the summary row query.
	summaryRow := pgxmock.NewRows([]string{
		"id", "external_summary_id", "host", "port", "user_name", "db_name", "synced_at",
	}).AddRow(summaryID, "sum123", "host1", 5432, "user1", "db1", now)
	mock.ExpectQuery("SELECT").WithArgs(summaryID).WillReturnRows(summaryRow)

	// Expect the schemas query.
	schemaRows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(schemaID, "public")
	mock.ExpectQuery("SELECT").WithArgs(summaryID).WillReturnRows(schemaRows)

	// Expect the tables query for the "public" schema.
	tableRows := pgxmock.NewRows([]string{"id", "name", "row_count", "size_mb"}).
		AddRow(tableID, "users", int64(1240), 12.5)
	mock.ExpectQuery("SELECT").WithArgs(schemaID).WillReturnRows(tableRows)

	result, err := repo.GetSummary(context.Background(), summaryID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, summaryID, result.ID)
	assert.Len(t, result.Schemas, 1)
	assert.Equal(t, "public", result.Schemas[0].Name)
	assert.Len(t, result.Schemas[0].Tables, 1)
	assert.Equal(t, "users", result.Schemas[0].Tables[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSummary_SummaryNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := repository.NewPostgresRepository(mock)

	id := uuid.New()
	mock.ExpectQuery("SELECT").WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "external_summary_id", "host", "port", "user_name", "db_name", "synced_at",
		}))

	result, err := repo.GetSummary(context.Background(), id)

	require.Error(t, err)
	assert.Nil(t, result)
}
