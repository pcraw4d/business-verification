package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/cache"
	"kyb-platform/services/risk-assessment-service/internal/config"
)

func main() {
	// Parse command line flags
	var (
		configFile = flag.String("config", "configs/cache_config.yaml", "Configuration file path")
		once       = flag.Bool("once", false, "Run once and exit")
		interval   = flag.Duration("interval", 5*time.Minute, "Prefetch interval")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Initialize logger
	var logger *zap.Logger
	var err error

	if *verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting cache warmer",
		zap.String("config", *configFile),
		zap.Bool("once", *once),
		zap.Duration("interval", *interval))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create cache coordinator configuration
	coordinatorConfig := &cache.CacheCoordinatorConfig{
		L1Config: &cache.MemoryCacheConfig{
			Capacity:        1000,
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
			EnableStats:     true,
		},
		L2Config: &cache.RedisCacheConfig{
			Addrs:             []string{cfg.Redis.URL},
			Password:          cfg.Redis.Password,
			DB:                cfg.Redis.DB,
			PoolSize:          cfg.Redis.PoolSize,
			MinIdleConns:      cfg.Redis.MinIdleConns,
			MaxRetries:        cfg.Redis.MaxRetries,
			DialTimeout:       cfg.Redis.DialTimeout,
			ReadTimeout:       cfg.Redis.ReadTimeout,
			WriteTimeout:      cfg.Redis.WriteTimeout,
			PoolTimeout:       cfg.Redis.PoolTimeout,
			IdleTimeout:       cfg.Redis.IdleTimeout,
			MaxConnAge:        30 * time.Minute,
			DefaultTTL:        5 * time.Minute,
			KeyPrefix:         "ra:",
			EnableMetrics:     true,
			EnableCompression: false,
		},
		EnableL1:       true,
		EnableL2:       true,
		EnableFallback: true,
		SyncInterval:   1 * time.Minute,
	}

	// Create cache coordinator
	coordinator, err := cache.NewCacheCoordinator(coordinatorConfig, logger)
	if err != nil {
		logger.Fatal("Failed to create cache coordinator", zap.Error(err))
	}
	defer coordinator.Close()

	// Create prefetch strategy
	prefetchConfig := &cache.PrefetchConfig{
		Enabled:           true,
		MaxPrefetchItems:  1000,
		PrefetchInterval:  *interval,
		PopularThreshold:  10,
		PriorityThreshold: 5,
		TTLMultiplier:     1.5,
	}

	prefetchStrategy := cache.NewPrefetchStrategy(coordinator, prefetchConfig, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, stopping cache warmer")
		cancel()
	}()

	// Perform initial cache warmup
	logger.Info("Performing initial cache warmup")
	if err := prefetchStrategy.WarmupCache(ctx); err != nil {
		logger.Error("Initial cache warmup failed", zap.Error(err))
	} else {
		logger.Info("Initial cache warmup completed successfully")
	}

	// Print initial statistics
	printStats(coordinator, prefetchStrategy, logger)

	if *once {
		logger.Info("Cache warmer completed (once mode)")
		return
	}

	// Start periodic prefetch routine
	logger.Info("Starting periodic prefetch routine", zap.Duration("interval", *interval))

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Cache warmer stopped")
			return
		case <-ticker.C:
			logger.Info("Starting periodic prefetch")

			// Perform prefetch operations
			if err := prefetchStrategy.PrefetchMLModelResults(ctx); err != nil {
				logger.Warn("ML model prefetch failed", zap.Error(err))
			}

			if err := prefetchStrategy.PrefetchIndustryData(ctx); err != nil {
				logger.Warn("Industry data prefetch failed", zap.Error(err))
			}

			if err := prefetchStrategy.PrefetchCountryData(ctx); err != nil {
				logger.Warn("Country data prefetch failed", zap.Error(err))
			}

			// Print statistics
			printStats(coordinator, prefetchStrategy, logger)

			logger.Info("Periodic prefetch completed")
		}
	}
}

func printStats(coordinator *cache.CacheCoordinator, strategy *cache.PrefetchStrategy, logger *zap.Logger) {
	// Get coordinator stats
	coordStats := coordinator.Stats()
	logger.Info("Cache Coordinator Stats",
		zap.Int64("total_hits", coordStats.TotalHits),
		zap.Int64("total_misses", coordStats.TotalMisses),
		zap.Float64("hit_rate", coordStats.HitRate),
		zap.Int64("l1_hits", coordStats.L1Hits),
		zap.Int64("l2_hits", coordStats.L2Hits),
		zap.Int64("l1_sets", coordStats.L1Sets),
		zap.Int64("l2_sets", coordStats.L2Sets))

	// Get L1 cache stats
	if l1Stats := coordinator.GetL1Stats(); l1Stats != nil {
		logger.Info("L1 Cache Stats",
			zap.Int64("hits", l1Stats.Hits),
			zap.Int64("misses", l1Stats.Misses),
			zap.Float64("hit_rate", l1Stats.HitRate),
			zap.Int("size", l1Stats.Size),
			zap.Int("capacity", l1Stats.Capacity),
			zap.Int64("evictions", l1Stats.Evictions))
	}

	// Get L2 cache stats
	if l2Stats := coordinator.GetL2Stats(); l2Stats != nil {
		logger.Info("L2 Cache Stats",
			zap.Int64("hits", l2Stats.Hits),
			zap.Int64("misses", l2Stats.Misses),
			zap.Float64("hit_rate", l2Stats.HitRate),
			zap.Int64("sets", l2Stats.Sets),
			zap.Int64("deletes", l2Stats.Deletes),
			zap.Int64("errors", l2Stats.Errors))
	}

	// Get prefetch stats
	prefetchStats := strategy.GetStats()
	logger.Info("Prefetch Stats",
		zap.Int64("total_prefetches", prefetchStats.TotalPrefetches),
		zap.Int64("successful_prefetches", prefetchStats.SuccessfulPrefetches),
		zap.Int64("failed_prefetches", prefetchStats.FailedPrefetches),
		zap.Duration("prefetch_time", prefetchStats.PrefetchTime),
		zap.Time("last_prefetch", prefetchStats.LastPrefetch))
}
