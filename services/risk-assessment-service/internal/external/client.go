package external

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client provides a base HTTP client for external API integrations
type Client struct {
	httpClient *http.Client
	logger     *zap.Logger
	baseURL    string
	apiKey     string
	timeout    time.Duration
}

// Config holds configuration for external API clients
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new external API client
func NewClient(config Config, logger *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:  logger,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		timeout: config.Timeout,
	}
}

// Get performs a GET request with retry logic
func (c *Client) Get(ctx context.Context, endpoint string, params map[string]string) (*http.Response, error) {
	req, err := c.buildRequest(ctx, "GET", endpoint, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	return c.doRequest(req)
}

// Post performs a POST request with retry logic
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	req, err := c.buildRequest(ctx, "POST", endpoint, nil, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	return c.doRequest(req)
}

// buildRequest builds an HTTP request
func (c *Client) buildRequest(ctx context.Context, method, endpoint string, params map[string]string, body interface{}) (*http.Request, error) {
	url := c.baseURL + endpoint

	// Add query parameters
	if len(params) > 0 {
		url += "?"
		first := true
		for key, value := range params {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// Add API key if available
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	req.Header.Set("User-Agent", "KYB-Risk-Assessment-Service/1.0")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// doRequest performs the HTTP request with retry logic
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, err)
			}

			// Exponential backoff
			delay := baseDelay * time.Duration(1<<attempt)
			c.logger.Warn("Request failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
				zap.Error(err))

			time.Sleep(delay)
			continue
		}

		// Check for HTTP error status codes
		if resp.StatusCode >= 400 {
			resp.Body.Close()
			if attempt == maxRetries {
				return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
			}

			// Retry on server errors (5xx)
			if resp.StatusCode >= 500 {
				delay := baseDelay * time.Duration(1<<attempt)
				c.logger.Warn("Server error, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("status_code", resp.StatusCode),
					zap.Duration("delay", delay))

				time.Sleep(delay)
				continue
			}

			// Don't retry on client errors (4xx)
			return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after all retries")
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
