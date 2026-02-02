package parallel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(cfg Config) *Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: base,
		apiKey:  cfg.APIKey,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) ListMonitors(ctx context.Context) (*ListMonitorsResponse, error) {
	var out ListMonitorsResponse
	if err := c.do(ctx, http.MethodGet, "/v1alpha/monitors", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateMonitor(ctx context.Context, req CreateMonitorRequest) (*MonitorResponse, error) {
	var out MonitorResponse
	if err := c.do(ctx, http.MethodPost, "/v1alpha/monitors", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateMonitorWebhook(ctx context.Context, monitorID string, webhook *MonitorWebhook) (*MonitorResponse, error) {
	var out MonitorResponse
	body := UpdateMonitorWebhookRequest{Webhook: webhook}
	path := fmt.Sprintf("/v1alpha/monitors/%s", monitorID)
	if err := c.do(ctx, http.MethodPost, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateMonitor(ctx context.Context, monitorID string, req UpdateMonitorRequest) (*MonitorResponse, error) {
	var out MonitorResponse
	path := fmt.Sprintf("/v1alpha/monitors/%s", monitorID)
	if err := c.do(ctx, http.MethodPost, path, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteMonitor(ctx context.Context, monitorID string) (*MonitorResponse, error) {
	var out MonitorResponse
	path := fmt.Sprintf("/v1alpha/monitors/%s", monitorID)
	if err := c.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, r)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, b)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
