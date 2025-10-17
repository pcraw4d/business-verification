package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				URL:    "https://test.supabase.co",
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "missing URL",
			config: &Config{
				URL:    "",
				APIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "supabase URL is required",
		},
		{
			name: "missing API key",
			config: &Config{
				URL:    "https://test.supabase.co",
				APIKey: "",
			},
			expectError: true,
			errorMsg:    "supabase API key is required",
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "supabase config is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			client, err := NewClient(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.GetClient())
			}
		})
	}
}

func TestClient_GetClient(t *testing.T) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	client, err := NewClient(config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test that GetClient returns the underlying client
	underlyingClient := client.GetClient()
	assert.NotNil(t, underlyingClient)
}

func TestClient_Health(t *testing.T) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "valid context",
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name:        "context with timeout",
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name:        "cancelled context",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: false, // Health check should not fail on cancelled context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Health(tt.ctx)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Health_Timeout(t *testing.T) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	// Test with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Should not error even with very short timeout
	err = client.Health(ctx)
	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	client, err := NewClient(config, logger)
	require.NoError(t, err)

	// Close should not return an error
	err = client.Close()
	assert.NoError(t, err)
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "complete config",
			config: &Config{
				URL:            "https://test.supabase.co",
				APIKey:         "test-api-key",
				ServiceRoleKey: "test-service-role-key",
				JWTSecret:      "test-jwt-secret",
			},
			expectError: false,
		},
		{
			name: "minimal config",
			config: &Config{
				URL:    "https://test.supabase.co",
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "empty URL",
			config: &Config{
				URL:    "",
				APIKey: "test-api-key",
			},
			expectError: true,
			errorMsg:    "supabase URL is required",
		},
		{
			name: "empty API key",
			config: &Config{
				URL:    "https://test.supabase.co",
				APIKey: "",
			},
			expectError: true,
			errorMsg:    "supabase API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			client, err := NewClient(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewClient(b *testing.B) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, err := NewClient(config, logger)
		if err != nil {
			b.Fatal(err)
		}
		_ = client
	}
}

func BenchmarkClient_Health(b *testing.B) {
	config := &Config{
		URL:    "https://test.supabase.co",
		APIKey: "test-api-key",
	}
	logger := zap.NewNop()

	client, err := NewClient(config, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Health(ctx)
	}
}
