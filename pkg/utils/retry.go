package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries   int           // Maximum number of retries (default: 3)
	BaseDelay    time.Duration // Base delay for exponential backoff (default: 100ms)
	MaxDelay     time.Duration // Maximum delay between retries (default: 1s)
	IsRetryable  func(error) bool // Function to determine if error is retryable
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		IsRetryable: func(err error) bool {
			return IsRetryableError(err)
		},
	}
}

// IsRetryableError checks if an error is retryable
// Retries on: network errors, timeouts, 5xx HTTP errors
// Does not retry on: 4xx HTTP errors (client errors)
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network errors
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for DNS errors
	if _, ok := err.(*net.DNSError); ok {
		return true
	}

	// Check for HTTP errors (5xx are retryable, 4xx are not)
	if httpErr, ok := err.(*HTTPError); ok {
		// Retry on 5xx errors (server errors)
		if httpErr.StatusCode >= 500 && httpErr.StatusCode < 600 {
			return true
		}
		// Don't retry on 4xx errors (client errors)
		return false
	}

	// Check error message for common retryable patterns
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"temporary failure",
		"i/o timeout",
		"context deadline exceeded",
		"502",
		"503",
		"504",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// RetryWithExponentialBackoff executes an operation with exponential backoff retry logic
// The operation function should return an error if it fails, nil on success
func RetryWithExponentialBackoff(
	ctx context.Context,
	config RetryConfig,
	operation func() error,
) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !config.IsRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate exponential backoff delay: baseDelay * 2^attempt
		delay := config.BaseDelay * time.Duration(1<<uint(attempt))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait before retry (respect context cancellation)
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// RetryHTTPRequest executes an HTTP request with retry logic
// Returns the response and error
func RetryHTTPRequest(
	ctx context.Context,
	config RetryConfig,
	client *http.Client,
	req *http.Request,
) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Create request with context
		reqWithCtx := req.WithContext(ctx)

		// Execute request
		resp, err := client.Do(reqWithCtx)
		if err == nil {
			// Check HTTP status code
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return resp, nil // Success
			}

			// Create HTTP error for status codes
			httpErr := &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    http.StatusText(resp.StatusCode),
			}

			// Close response body if not successful
			resp.Body.Close()

			// Check if error is retryable
			if !config.IsRetryable(httpErr) {
				return nil, httpErr
			}

			lastErr = httpErr
		} else {
			// Network error
			if !config.IsRetryable(err) {
				return nil, err
			}
			lastErr = err
		}

		// Don't retry on last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := config.BaseDelay * time.Duration(1<<uint(attempt))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	if lastResp != nil {
		lastResp.Body.Close()
	}

	return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	// Simple case-insensitive contains check
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}

