package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the risk assessment service
type Config struct {
	Server   ServerConfig   `json:"server"`
	Supabase SupabaseConfig `json:"supabase"`
	Redis    RedisConfig    `json:"redis"`
	ML       MLConfig       `json:"ml"`
	External ExternalConfig `json:"external"`
	Logging  LoggingConfig  `json:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// SupabaseConfig holds Supabase configuration
type SupabaseConfig struct {
	URL            string `json:"url"`
	APIKey         string `json:"api_key"`
	ServiceRoleKey string `json:"service_role_key"`
	JWTSecret      string `json:"jwt_secret"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string `json:"url"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// MLConfig holds machine learning configuration
type MLConfig struct {
	ModelPath     string  `json:"model_path"`
	TrainingData  string  `json:"training_data"`
	BatchSize     int     `json:"batch_size"`
	LearningRate  float64 `json:"learning_rate"`
	MaxIterations int     `json:"max_iterations"`
}

// ExternalConfig holds external API configuration
type ExternalConfig struct {
	ThomsonReuters ThomsonReutersConfig `json:"thomson_reuters"`
	OFAC           OFACConfig           `json:"ofac"`
	WorldCheck     WorldCheckConfig     `json:"worldcheck"`
	NewsAPI        NewsAPIConfig        `json:"news_api"`
	OpenCorporates OpenCorporatesConfig `json:"opencorporates"`
}

// ThomsonReutersConfig holds Thomson Reuters API configuration
type ThomsonReutersConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// OFACConfig holds OFAC API configuration
type OFACConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// WorldCheckConfig holds World-Check API configuration
type WorldCheckConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// NewsAPIConfig holds News API configuration
type NewsAPIConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// OpenCorporatesConfig holds OpenCorporates API configuration
type OpenCorporatesConfig struct {
	APIKey  string        `json:"api_key"`
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvAsString("PORT", "8080"),
			Host:         getEnvAsString("HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Supabase: SupabaseConfig{
			URL:            getEnvAsString("SUPABASE_URL", ""),
			APIKey:         getEnvAsString("SUPABASE_ANON_KEY", getEnvAsString("SUPABASE_API_KEY", "")),
			ServiceRoleKey: getEnvAsString("SUPABASE_SERVICE_ROLE_KEY", ""),
			JWTSecret:      getEnvAsString("SUPABASE_JWT_SECRET", ""),
		},
		Redis: RedisConfig{
			URL:      getEnvAsString("REDIS_URL", "redis://localhost:6379"),
			Password: getEnvAsString("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		ML: MLConfig{
			ModelPath:     getEnvAsString("ML_MODEL_PATH", "./models"),
			TrainingData:  getEnvAsString("ML_TRAINING_DATA", "./data"),
			BatchSize:     getEnvAsInt("ML_BATCH_SIZE", 32),
			LearningRate:  getEnvAsFloat64("ML_LEARNING_RATE", 0.01),
			MaxIterations: getEnvAsInt("ML_MAX_ITERATIONS", 1000),
		},
		External: ExternalConfig{
			ThomsonReuters: ThomsonReutersConfig{
				APIKey:  getEnvAsString("THOMSON_REUTERS_API_KEY", ""),
				BaseURL: getEnvAsString("THOMSON_REUTERS_BASE_URL", "https://api.thomsonreuters.com"),
				Timeout: getEnvAsDuration("THOMSON_REUTERS_TIMEOUT", 30*time.Second),
			},
			OFAC: OFACConfig{
				APIKey:  getEnvAsString("OFAC_API_KEY", ""),
				BaseURL: getEnvAsString("OFAC_BASE_URL", "https://api.treasury.gov"),
				Timeout: getEnvAsDuration("OFAC_TIMEOUT", 30*time.Second),
			},
			WorldCheck: WorldCheckConfig{
				APIKey:  getEnvAsString("WORLDCHECK_API_KEY", ""),
				BaseURL: getEnvAsString("WORLDCHECK_BASE_URL", "https://api.worldcheck.com"),
				Timeout: getEnvAsDuration("WORLDCHECK_TIMEOUT", 30*time.Second),
			},
			NewsAPI: NewsAPIConfig{
				APIKey:  getEnvAsString("NEWS_API_KEY", ""),
				BaseURL: getEnvAsString("NEWS_API_BASE_URL", "https://newsapi.org/v2"),
				Timeout: getEnvAsDuration("NEWS_API_TIMEOUT", 30*time.Second),
			},
			OpenCorporates: OpenCorporatesConfig{
				APIKey:  getEnvAsString("OPENCORPORATES_API_KEY", ""),
				BaseURL: getEnvAsString("OPENCORPORATES_BASE_URL", "https://api.opencorporates.com"),
				Timeout: getEnvAsDuration("OPENCORPORATES_TIMEOUT", 30*time.Second),
			},
		},
		Logging: LoggingConfig{
			Level:  getEnvAsString("LOG_LEVEL", "info"),
			Format: getEnvAsString("LOG_FORMAT", "json"),
		},
	}

	// Validate required configuration
	if cfg.Supabase.URL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if cfg.Supabase.APIKey == "" {
		return nil, fmt.Errorf("SUPABASE_ANON_KEY is required")
	}

	return cfg, nil
}

// Helper functions for environment variable parsing
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
