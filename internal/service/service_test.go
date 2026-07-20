package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/RvShivam/postgres-summary-service/internal/mocks"
	"github.com/RvShivam/postgres-summary-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── SyncSummary ───────────────────────────────────────────────────────────────

func TestSyncSummary_Success(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	externalResp := &external.SummaryResponse{
		SummaryID: "sum123",
		Schemas: []external.SchemaSummary{
			{
				Name: "public",
				Tables: []external.TableSummary{
					{Name: "users", RowCount: 1240, SizeMB: 12.5},
					{Name: "orders", RowCount: 580, SizeMB: 8.3},
				},
			},
		},
	}

	mockClient.On("FetchSummary", mock.Anything, mock.AnythingOfType("external.SummaryRequest")).
		Return(externalResp, nil)

	mockRepo.On("SaveSummary", mock.Anything, mock.AnythingOfType("*domain.Summary")).
		Return(nil)

	svc := service.New(mockRepo, mockClient)

	got, err := svc.SyncSummary(context.Background(), service.SyncRequest{
		Host:     "remote-db.example.com",
		Port:     5432,
		User:     "readonly",
		Password: "pass",
		DBName:   "sample",
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "sum123", got.ExternalSummaryID)
	assert.Equal(t, "remote-db.example.com", got.Host)
	assert.Equal(t, "sample", got.DBName)
	// Password must NOT be stored on the domain object.
	assert.Len(t, got.Schemas, 1)
	assert.Equal(t, "public", got.Schemas[0].Name)
	assert.Len(t, got.Schemas[0].Tables, 2)

	mockClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestSyncSummary_ExternalClientError(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	mockClient.On("FetchSummary", mock.Anything, mock.Anything).
		Return(nil, errors.New("external service unavailable after 3 attempts"))

	svc := service.New(mockRepo, mockClient)

	got, err := svc.SyncSummary(context.Background(), service.SyncRequest{
		Host:   "remote-db.example.com",
		Port:   5432,
		User:   "readonly",
		DBName: "sample",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	// Repository should never be called if the external call fails.
	mockRepo.AssertNotCalled(t, "SaveSummary")
}

func TestSyncSummary_RepositorySaveError(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	mockClient.On("FetchSummary", mock.Anything, mock.Anything).
		Return(&external.SummaryResponse{SummaryID: "sum456", Schemas: nil}, nil)

	mockRepo.On("SaveSummary", mock.Anything, mock.Anything).
		Return(errors.New("db write failed"))

	svc := service.New(mockRepo, mockClient)

	got, err := svc.SyncSummary(context.Background(), service.SyncRequest{
		Host:   "remote-db.example.com",
		Port:   5432,
		User:   "readonly",
		DBName: "sample",
	})

	require.Error(t, err)
	assert.Nil(t, got)
}

// ── ListSummaries ─────────────────────────────────────────────────────────────

func TestListSummaries_Success(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	expected := []domain.SummaryOverview{
		{
			ID:                uuid.New(),
			ExternalSummaryID: "sum123",
			Host:              "host1",
			Port:              5432,
			User:              "user1",
			DBName:            "db1",
			SyncedAt:          time.Now(),
		},
	}

	mockRepo.On("ListSummaries", mock.Anything).Return(expected, nil)

	svc := service.New(mockRepo, mockClient)
	got, err := svc.ListSummaries(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}

func TestListSummaries_RepositoryError(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	mockRepo.On("ListSummaries", mock.Anything).
		Return(nil, errors.New("db read failed"))

	svc := service.New(mockRepo, mockClient)
	got, err := svc.ListSummaries(context.Background())

	require.Error(t, err)
	assert.Nil(t, got)
}

// ── GetSummary ────────────────────────────────────────────────────────────────

func TestGetSummary_Success(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	id := uuid.New()
	expected := &domain.Summary{
		ID:                id,
		ExternalSummaryID: "sum123",
		Host:              "host1",
		Port:              5432,
		User:              "user1",
		DBName:            "db1",
		SyncedAt:          time.Now(),
		Schemas:           []domain.Schema{},
	}

	mockRepo.On("GetSummary", mock.Anything, id).Return(expected, nil)

	svc := service.New(mockRepo, mockClient)
	got, err := svc.GetSummary(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}

func TestGetSummary_NotFound(t *testing.T) {
	mockClient := &mocks.MockExternalClient{}
	mockRepo := &mocks.MockRepository{}

	id := uuid.New()

	mockRepo.On("GetSummary", mock.Anything, id).
		Return(nil, errors.New("no rows in result set"))

	svc := service.New(mockRepo, mockClient)
	got, err := svc.GetSummary(context.Background(), id)

	require.Error(t, err)
	assert.Nil(t, got)
}
