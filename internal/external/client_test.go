package external_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RvShivam/postgres-summary-service/internal/external"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchSummary_Success(t *testing.T) {
	// Arrange: spin up a fake external service.
	response := external.SummaryResponse{
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := external.NewClient(srv.URL)

	// Act
	got, err := client.FetchSummary(context.Background(), external.SummaryRequest{
		Host:     "remote-db.example.com",
		Port:     5432,
		User:     "readonly",
		Password: "pass",
		DBName:   "sample",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "sum123", got.SummaryID)
	assert.Len(t, got.Schemas, 1)
	assert.Equal(t, "public", got.Schemas[0].Name)
	assert.Len(t, got.Schemas[0].Tables, 2)
}

func TestFetchSummary_Non200PermanentError(t *testing.T) {
	// A 400 should NOT be retried and should surface immediately.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := external.NewClient(srv.URL)

	_, err := client.FetchSummary(context.Background(), external.SummaryRequest{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestFetchSummary_5xxRetryExhausted(t *testing.T) {
	// A 500 should be retried; after maxRetries we expect an "unavailable" error.
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := external.NewClient(srv.URL)

	_, err := client.FetchSummary(context.Background(), external.SummaryRequest{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable after")
	assert.Equal(t, 3, attempts, "should have attempted exactly maxRetries times")
}

func TestFetchSummary_ContextCancelledDuringRetry(t *testing.T) {
	// Verify that a cancelled context aborts mid-retry without hanging.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := external.NewClient(srv.URL)

	_, err := client.FetchSummary(ctx, external.SummaryRequest{})

	require.Error(t, err)
}

func TestFetchSummary_UnreachableService(t *testing.T) {
	// Point at a URL that nothing is listening on.
	client := external.NewClient("http://127.0.0.1:19999/api/summary")

	_, err := client.FetchSummary(context.Background(), external.SummaryRequest{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable after")
}
