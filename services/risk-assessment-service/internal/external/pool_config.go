package external

import (
	"fmt"
	"time"
)

// DefaultPoolConfigs provides default configurations for different external API providers
var DefaultPoolConfigs = map[string]*PoolConfig{
	"thomson_reuters": {
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             30 * time.Second,
		MaxRetries:          3,
		RetryDelay:          2 * time.Second,
	},
	"ofac": {
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     30,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             20 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
	},
	"news_api": {
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             15 * time.Second,
		MaxRetries:          2,
		RetryDelay:          1 * time.Second,
	},
	"opencorporates": {
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             15 * time.Second,
		MaxRetries:          2,
		RetryDelay:          1 * time.Second,
	},
	"worldcheck": {
		MaxIdleConns:        30,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     30,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             20 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
	},
	"government": {
		MaxIdleConns:        15,
		MaxIdleConnsPerHost: 3,
		MaxConnsPerHost:     15,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             25 * time.Second,
		MaxRetries:          2,
		RetryDelay:          2 * time.Second,
	},
}

// GetDefaultConfig returns the default configuration for a provider
func GetDefaultConfig(provider string) *PoolConfig {
	if config, exists := DefaultPoolConfigs[provider]; exists {
		// Return a copy to avoid modifying the original
		return &PoolConfig{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.IdleConnTimeout,
			DisableKeepAlives:   config.DisableKeepAlives,
			Timeout:             config.Timeout,
			MaxRetries:          config.MaxRetries,
			RetryDelay:          config.RetryDelay,
		}
	}

	// Return a generic default configuration
	return &PoolConfig{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     60 * time.Second,
		DisableKeepAlives:   false,
		Timeout:             30 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
	}
}

// ValidateConfig validates a pool configuration
func ValidateConfig(config *PoolConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MaxIdleConns < 0 {
		return fmt.Errorf("MaxIdleConns must be non-negative")
	}

	if config.MaxIdleConnsPerHost < 0 {
		return fmt.Errorf("MaxIdleConnsPerHost must be non-negative")
	}

	if config.MaxConnsPerHost < 0 {
		return fmt.Errorf("MaxConnsPerHost must be non-negative")
	}

	if config.IdleConnTimeout < 0 {
		return fmt.Errorf("IdleConnTimeout must be non-negative")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("Timeout must be non-negative")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries must be non-negative")
	}

	if config.RetryDelay < 0 {
		return fmt.Errorf("RetryDelay must be non-negative")
	}

	// Validate relationships
	if config.MaxIdleConnsPerHost > config.MaxIdleConns {
		return fmt.Errorf("MaxIdleConnsPerHost cannot be greater than MaxIdleConns")
	}

	if config.MaxConnsPerHost > config.MaxIdleConns {
		return fmt.Errorf("MaxConnsPerHost cannot be greater than MaxIdleConns")
	}

	return nil
}

// MergeConfigs merges a base configuration with overrides
func MergeConfigs(base, override *PoolConfig) *PoolConfig {
	if base == nil {
		base = &PoolConfig{}
	}
	if override == nil {
		return base
	}

	merged := &PoolConfig{
		MaxIdleConns:        base.MaxIdleConns,
		MaxIdleConnsPerHost: base.MaxIdleConnsPerHost,
		MaxConnsPerHost:     base.MaxConnsPerHost,
		IdleConnTimeout:     base.IdleConnTimeout,
		DisableKeepAlives:   base.DisableKeepAlives,
		Timeout:             base.Timeout,
		MaxRetries:          base.MaxRetries,
		RetryDelay:          base.RetryDelay,
	}

	// Apply overrides
	if override.MaxIdleConns != 0 {
		merged.MaxIdleConns = override.MaxIdleConns
	}
	if override.MaxIdleConnsPerHost != 0 {
		merged.MaxIdleConnsPerHost = override.MaxIdleConnsPerHost
	}
	if override.MaxConnsPerHost != 0 {
		merged.MaxConnsPerHost = override.MaxConnsPerHost
	}
	if override.IdleConnTimeout != 0 {
		merged.IdleConnTimeout = override.IdleConnTimeout
	}
	if override.Timeout != 0 {
		merged.Timeout = override.Timeout
	}
	if override.MaxRetries != 0 {
		merged.MaxRetries = override.MaxRetries
	}
	if override.RetryDelay != 0 {
		merged.RetryDelay = override.RetryDelay
	}
	// DisableKeepAlives is a boolean, so we check if it's explicitly set
	merged.DisableKeepAlives = override.DisableKeepAlives

	return merged
}

// GetRecommendedConfig returns a recommended configuration based on provider characteristics
func GetRecommendedConfig(provider string, characteristics ProviderCharacteristics) *PoolConfig {
	base := GetDefaultConfig(provider)

	// Adjust based on characteristics
	if characteristics.HighVolume {
		base.MaxIdleConns *= 2
		base.MaxIdleConnsPerHost *= 2
		base.MaxConnsPerHost *= 2
	}

	if characteristics.HighLatency {
		base.Timeout *= 2
		base.IdleConnTimeout *= 2
	}

	if characteristics.Unreliable {
		base.MaxRetries += 1
		base.RetryDelay *= 2
	}

	if characteristics.RateLimited {
		base.MaxIdleConnsPerHost = max(1, base.MaxIdleConnsPerHost/2)
		base.MaxConnsPerHost = max(1, base.MaxConnsPerHost/2)
	}

	return base
}

// ProviderCharacteristics represents characteristics of an external API provider
type ProviderCharacteristics struct {
	HighVolume  bool `json:"high_volume"`
	HighLatency bool `json:"high_latency"`
	Unreliable  bool `json:"unreliable"`
	RateLimited bool `json:"rate_limited"`
	Expensive   bool `json:"expensive"`
}

// GetProviderCharacteristics returns characteristics for known providers
func GetProviderCharacteristics(provider string) ProviderCharacteristics {
	switch provider {
	case "thomson_reuters":
		return ProviderCharacteristics{
			HighVolume:  true,
			HighLatency: true,
			Unreliable:  false,
			RateLimited: true,
			Expensive:   true,
		}
	case "ofac":
		return ProviderCharacteristics{
			HighVolume:  false,
			HighLatency: false,
			Unreliable:  false,
			RateLimited: false,
			Expensive:   false,
		}
	case "news_api":
		return ProviderCharacteristics{
			HighVolume:  true,
			HighLatency: false,
			Unreliable:  true,
			RateLimited: true,
			Expensive:   false,
		}
	case "opencorporates":
		return ProviderCharacteristics{
			HighVolume:  true,
			HighLatency: false,
			Unreliable:  true,
			RateLimited: true,
			Expensive:   false,
		}
	case "worldcheck":
		return ProviderCharacteristics{
			HighVolume:  false,
			HighLatency: true,
			Unreliable:  false,
			RateLimited: true,
			Expensive:   true,
		}
	case "government":
		return ProviderCharacteristics{
			HighVolume:  false,
			HighLatency: true,
			Unreliable:  true,
			RateLimited: false,
			Expensive:   false,
		}
	default:
		return ProviderCharacteristics{
			HighVolume:  false,
			HighLatency: false,
			Unreliable:  true,
			RateLimited: false,
			Expensive:   false,
		}
	}
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
