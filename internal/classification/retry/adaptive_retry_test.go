package retry

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestAdaptiveRetryStrategy_ShouldRetry(t *testing.T) {
	strategy := NewAdaptiveRetryStrategy(3, 1*time.Second)

	tests := []struct {
		name           string
		err            error
		httpStatusCode int
		wantRetry      bool
		wantCount      int
	}{
		{
			name:           "Permanent error 400 - no retry",
			err:            nil,
			httpStatusCode: 400,
			wantRetry:      false,
			wantCount:      0,
		},
		{
			name:           "Permanent error 403 - no retry",
			err:            nil,
			httpStatusCode: 403,
			wantRetry:      false,
			wantCount:      0,
		},
		{
			name:           "Permanent error 404 - no retry",
			err:            nil,
			httpStatusCode: 404,
			wantRetry:      false,
			wantCount:      0,
		},
		{
			name:           "Rate limited 429 - more retries",
			err:            nil,
			httpStatusCode: 429,
			wantRetry:      true,
			wantCount:      5,
		},
		{
			name:           "Server error 500 - retry",
			err:            nil,
			httpStatusCode: 500,
			wantRetry:      true,
			wantCount:      3,
		},
		{
			name:           "DNS error - retry with more attempts",
			err:            &net.DNSError{},
			httpStatusCode: 0,
			wantRetry:      true,
			wantCount:      4, // defaultMaxRetries + 1
		},
		{
			name:           "Timeout error - retry",
			err:            &timeoutError{},
			httpStatusCode: 0,
			wantRetry:      true,
			wantCount:      3,
		},
		{
			name:           "Network error - retry",
			err:            &net.OpError{},
			httpStatusCode: 0,
			wantRetry:      true,
			wantCount:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retry, count := strategy.ShouldRetry(tt.err, tt.httpStatusCode)
			if retry != tt.wantRetry {
				t.Errorf("ShouldRetry() retry = %v, want %v", retry, tt.wantRetry)
			}
			if count != tt.wantCount {
				t.Errorf("ShouldRetry() count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestAdaptiveRetryStrategy_CalculateBackoff(t *testing.T) {
	strategy := NewAdaptiveRetryStrategy(3, 1*time.Second)

	tests := []struct {
		name      string
		attempt   int
		baseDelay time.Duration
		wantMax   time.Duration // Maximum expected backoff
	}{
		{
			name:      "First attempt",
			attempt:   1,
			baseDelay: 1 * time.Second,
			wantMax:   2 * time.Second,
		},
		{
			name:      "Second attempt",
			attempt:   2,
			baseDelay: 1 * time.Second,
			wantMax:   4 * time.Second,
		},
		{
			name:      "Third attempt",
			attempt:   3,
			baseDelay: 1 * time.Second,
			wantMax:   8 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := strategy.CalculateBackoff(tt.attempt, tt.baseDelay)
			if backoff > tt.wantMax {
				t.Errorf("CalculateBackoff() = %v, want <= %v", backoff, tt.wantMax)
			}
			if backoff > 30*time.Second {
				t.Errorf("CalculateBackoff() = %v, should be capped at 30s", backoff)
			}
		})
	}
}

func TestAdaptiveRetryStrategy_RecordResult(t *testing.T) {
	strategy := NewAdaptiveRetryStrategy(3, 1*time.Second)

	// Record some failures
	strategy.RecordResult(errors.New("test error"), 500, false)
	strategy.RecordResult(errors.New("test error"), 500, false)
	strategy.RecordResult(errors.New("test error"), 500, false)

	// Record a success (different status code)
	strategy.RecordResult(nil, 200, true)

	// Check that history is being tracked for 500 errors
	strategy.historyMutex.RLock()
	stats500, exists500 := strategy.errorHistory["http_500"]
	stats200, exists200 := strategy.errorHistory["http_200"]
	strategy.historyMutex.RUnlock()

	if !exists500 {
		t.Error("Expected error history to be tracked for 500 errors")
	}
	if stats500 == nil {
		t.Error("Expected stats to be non-nil for 500 errors")
	}
	if stats500.TotalAttempts != 3 {
		t.Errorf("Expected 3 total attempts for 500 errors, got %d", stats500.TotalAttempts)
	}
	if stats500.FailureCount != 3 {
		t.Errorf("Expected 3 failures, got %d", stats500.FailureCount)
	}

	// Check 200 success history
	if !exists200 {
		t.Error("Expected success history to be tracked for 200 responses")
	}
	if stats200 == nil {
		t.Error("Expected stats to be non-nil for 200 responses")
	}
	if stats200.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", stats200.SuccessCount)
	}
}

func TestRetryWithStrategy(t *testing.T) {
	strategy := NewAdaptiveRetryStrategy(3, 100*time.Millisecond)

	t.Run("Success on first attempt", func(t *testing.T) {
		attempts := 0
		err := RetryWithStrategy(
			context.Background(),
			strategy,
			func() error {
				attempts++
				return nil
			},
			nil,
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("Success after retries", func(t *testing.T) {
		attempts := 0
		err := RetryWithStrategy(
			context.Background(),
			strategy,
			func() error {
				attempts++
				if attempts < 2 {
					return errors.New("temporary error")
				}
				return nil
			},
			nil,
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("No retry for permanent error", func(t *testing.T) {
		attempts := 0
		statusCode := 400
		err := RetryWithStrategy(
			context.Background(),
			strategy,
			func() error {
				attempts++
				return errors.New("permanent error")
			},
			&statusCode,
		)

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt (no retry), got %d", attempts)
		}
	})
}

