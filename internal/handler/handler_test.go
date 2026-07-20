package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RvShivam/postgres-summary-service/internal/domain"
	"github.com/RvShivam/postgres-summary-service/internal/handler"
	"github.com/RvShivam/postgres-summary-service/internal/mocks"
	"github.com/RvShivam/postgres-summary-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(h *handler.Handler) *gin.Engine {
	r := gin.New()
	r.POST("/summary/sync", h.SyncSummary)
	r.GET("/summaries", h.ListSummaries)
	r.GET("/summaries/:id", h.GetSummary)
	return r
}

// ── POST /summary/sync ────────────────────────────────────────────────────────

func TestSyncSummary_Handler_Success(t *testing.T) {
	mockSvc := &mocks.MockService{}

	id := uuid.New()
	returnedSummary := &domain.Summary{
		ID:                id,
		ExternalSummaryID: "sum123",
		Host:              "remote-db.example.com",
		Port:              5432,
		User:              "readonly",
		DBName:            "sample",
		SyncedAt:          time.Now(),
		Schemas:           []domain.Schema{},
	}

	mockSvc.On("SyncSummary", mock.Anything, service.SyncRequest{
		Host:     "remote-db.example.com",
		Port:     5432,
		User:     "readonly",
		Password: "pass",
		DBName:   "sample",
	}).Return(returnedSummary, nil)

	h := handler.New(mockSvc)
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"host":     "remote-db.example.com",
		"port":     5432,
		"user":     "readonly",
		"password": "pass",
		"dbname":   "sample",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/summary/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestSyncSummary_Handler_InvalidBody(t *testing.T) {
	mockSvc := &mocks.MockService{}
	h := handler.New(mockSvc)
	r := newRouter(h)

	// Missing required fields
	body, _ := json.Marshal(map[string]any{
		"host": "only-host-no-other-fields",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/summary/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "SyncSummary")
}

func TestSyncSummary_Handler_ServiceError(t *testing.T) {
	mockSvc := &mocks.MockService{}

	mockSvc.On("SyncSummary", mock.Anything, mock.Anything).
		Return(nil, errors.New("external service unavailable"))

	h := handler.New(mockSvc)
	r := newRouter(h)

	body, _ := json.Marshal(map[string]any{
		"host":     "remote-db.example.com",
		"port":     5432,
		"user":     "readonly",
		"password": "pass",
		"dbname":   "sample",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/summary/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ── GET /summaries ────────────────────────────────────────────────────────────

func TestListSummaries_Handler_Success(t *testing.T) {
	mockSvc := &mocks.MockService{}

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

	mockSvc.On("ListSummaries", mock.Anything).Return(expected, nil)

	h := handler.New(mockSvc)
	r := newRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/summaries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got []domain.SummaryOverview
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Len(t, got, 1)
}

func TestListSummaries_Handler_ServiceError(t *testing.T) {
	mockSvc := &mocks.MockService{}
	mockSvc.On("ListSummaries", mock.Anything).Return(nil, errors.New("db failure"))

	h := handler.New(mockSvc)
	r := newRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/summaries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ── GET /summaries/:id ────────────────────────────────────────────────────────

func TestGetSummary_Handler_Success(t *testing.T) {
	mockSvc := &mocks.MockService{}

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

	mockSvc.On("GetSummary", mock.Anything, id).Return(expected, nil)

	h := handler.New(mockSvc)
	r := newRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/summaries/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSummary_Handler_InvalidUUID(t *testing.T) {
	mockSvc := &mocks.MockService{}
	h := handler.New(mockSvc)
	r := newRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/summaries/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "GetSummary")
}

func TestGetSummary_Handler_NotFound(t *testing.T) {
	mockSvc := &mocks.MockService{}

	id := uuid.New()
	mockSvc.On("GetSummary", mock.Anything, id).Return(nil, pgx.ErrNoRows)

	h := handler.New(mockSvc)
	r := newRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/summaries/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "summary not found", resp["error"])
}
