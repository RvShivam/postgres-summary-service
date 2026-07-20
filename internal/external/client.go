package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	maxRetries     = 3
	retryBaseDelay = 500 * time.Millisecond
)

type Client interface {
	FetchSummary(
		ctx context.Context,
		req SummaryRequest,
	) (*SummaryResponse, error)
}

type ExternalClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) Client {
	return &ExternalClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ExternalClient) FetchSummary(
	ctx context.Context,
	req SummaryRequest,
) (*SummaryResponse, error) {

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {

		// Back off before retrying (skip on first attempt).
		if attempt > 0 {
			delay := retryBaseDelay * (1 << (attempt - 1)) // 500ms, 1s, 2s …
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL,
			bytes.NewReader(body),
		)
		if err != nil {
			// Request construction failure is not retryable.
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(httpReq)
		if err != nil {
			// Network / timeout error — retry.
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		// 5xx responses are transient — retry.
		if resp.StatusCode >= http.StatusInternalServerError {
			lastErr = fmt.Errorf("external service returned status %d", resp.StatusCode)
			continue
		}

		// Any other non-200 response is a permanent error (4xx etc.) — don't retry.
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("external service returned status %d", resp.StatusCode)
		}

		var summary SummaryResponse

		if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
			return nil, err
		}

		return &summary, nil
	}

	return nil, fmt.Errorf("external service unavailable after %d attempts: %w", maxRetries, lastErr)
}

