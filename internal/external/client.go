package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external service returned status %d", resp.StatusCode)
	}

	var summary SummaryResponse

	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}
