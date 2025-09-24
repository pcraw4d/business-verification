package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigReloader handles hot-reloading of configuration files
type ConfigReloader struct {
	configPath    string
	environment   string
	loader        *EnhancedConfigLoader
	validator     *ConfigValidator
	watcher       *fsnotify.Watcher
	reloadChan    chan *EnhancedConfig
	errorChan     chan error
	stopChan      chan struct{}
	mu            sync.RWMutex
	currentConfig *EnhancedConfig
	callbacks     []ConfigReloadCallback
	enabled       bool
}

// ConfigReloadCallback is a function that gets called when configuration is reloaded
type ConfigReloadCallback func(oldConfig, newConfig *EnhancedConfig) error

// NewConfigReloader creates a new configuration reloader
func NewConfigReloader(configPath, environment string, loader *EnhancedConfigLoader) (*ConfigReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	validator := NewConfigValidator()

	reloader := &ConfigReloader{
		configPath:  configPath,
		environment: environment,
		loader:      loader,
		validator:   validator,
		watcher:     watcher,
		reloadChan:  make(chan *EnhancedConfig, 1),
		errorChan:   make(chan error, 1),
		stopChan:    make(chan struct{}),
		callbacks:   make([]ConfigReloadCallback, 0),
		enabled:     true,
	}

	return reloader, nil
}

// Start starts the configuration reloader
func (cr *ConfigReloader) Start(ctx context.Context) error {
	// Load initial configuration
	config, err := cr.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load initial configuration: %w", err)
	}

	cr.mu.Lock()
	cr.currentConfig = config
	cr.mu.Unlock()

	// Set up file watching
	if err := cr.setupFileWatching(); err != nil {
		return fmt.Errorf("failed to setup file watching: %w", err)
	}

	// Start the reload loop
	go cr.reloadLoop(ctx)

	// Start the file watcher loop
	go cr.watcherLoop(ctx)

	log.Printf("Configuration reloader started for environment: %s", cr.environment)
	return nil
}

// Stop stops the configuration reloader
func (cr *ConfigReloader) Stop() error {
	close(cr.stopChan)

	if cr.watcher != nil {
		return cr.watcher.Close()
	}

	return nil
}

// GetCurrentConfig returns the current configuration
func (cr *ConfigReloader) GetCurrentConfig() *EnhancedConfig {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.currentConfig
}

// AddReloadCallback adds a callback function that gets called when configuration is reloaded
func (cr *ConfigReloader) AddReloadCallback(callback ConfigReloadCallback) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.callbacks = append(cr.callbacks, callback)
}

// RemoveReloadCallback removes a callback function
func (cr *ConfigReloader) RemoveReloadCallback(callback ConfigReloadCallback) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for i, cb := range cr.callbacks {
		if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
			cr.callbacks = append(cr.callbacks[:i], cr.callbacks[i+1:]...)
			break
		}
	}
}

// Enable enables configuration reloading
func (cr *ConfigReloader) Enable() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.enabled = true
}

// Disable disables configuration reloading
func (cr *ConfigReloader) Disable() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.enabled = false
}

// IsEnabled returns whether configuration reloading is enabled
func (cr *ConfigReloader) IsEnabled() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.enabled
}

// setupFileWatching sets up file watching for configuration files
func (cr *ConfigReloader) setupFileWatching() error {
	// Watch the main config directory
	if err := cr.watcher.Add(cr.configPath); err != nil {
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Watch environment-specific directory if it exists
	envConfigPath := filepath.Join(cr.configPath, "environments")
	if _, err := os.Stat(envConfigPath); err == nil {
		if err := cr.watcher.Add(envConfigPath); err != nil {
			return fmt.Errorf("failed to watch environment config directory: %w", err)
		}
	}

	// Watch specific configuration files
	configFiles := []string{
		"config.yaml",
		"modules.yaml",
		"enhanced_features.yaml",
		"performance.yaml",
		"advanced_monitoring.yaml",
	}

	for _, filename := range configFiles {
		filePath := filepath.Join(cr.configPath, filename)
		if _, err := os.Stat(filePath); err == nil {
			if err := cr.watcher.Add(filePath); err != nil {
				return fmt.Errorf("failed to watch config file %s: %w", filename, err)
			}
		}
	}

	// Watch environment-specific config file
	envConfigFile := filepath.Join(cr.configPath, "environments", fmt.Sprintf("%s.yaml", cr.environment))
	if _, err := os.Stat(envConfigFile); err == nil {
		if err := cr.watcher.Add(envConfigFile); err != nil {
			return fmt.Errorf("failed to watch environment config file: %w", err)
		}
	}

	return nil
}

// watcherLoop handles file system events
func (cr *ConfigReloader) watcherLoop(ctx context.Context) {
	for {
		select {
		case event, ok := <-cr.watcher.Events:
			if !ok {
				return
			}
			cr.handleFileEvent(event)
		case err, ok := <-cr.watcher.Errors:
			if !ok {
				return
			}
			cr.errorChan <- fmt.Errorf("file watcher error: %w", err)
		case <-ctx.Done():
			return
		case <-cr.stopChan:
			return
		}
	}
}

// handleFileEvent handles file system events
func (cr *ConfigReloader) handleFileEvent(event fsnotify.Event) {
	cr.mu.RLock()
	enabled := cr.enabled
	cr.mu.RUnlock()

	if !enabled {
		return
	}

	// Only handle write events for configuration files
	if event.Op&fsnotify.Write == fsnotify.Write {
		// Check if the file is a configuration file
		if cr.isConfigFile(event.Name) {
			log.Printf("Configuration file changed: %s", event.Name)

			// Debounce the reload to avoid multiple reloads for the same change
			go cr.debouncedReload()
		}
	}
}

// isConfigFile checks if a file is a configuration file
func (cr *ConfigReloader) isConfigFile(filename string) bool {
	ext := filepath.Ext(filename)
	if ext != ".yaml" && ext != ".yml" {
		return false
	}

	// Check if it's in the config directory
	relPath, err := filepath.Rel(cr.configPath, filename)
	if err != nil {
		return false
	}

	// Allow main config files and environment-specific files
	configFiles := []string{
		"config.yaml",
		"modules.yaml",
		"enhanced_features.yaml",
		"performance.yaml",
		"advanced_monitoring.yaml",
	}

	for _, configFile := range configFiles {
		if relPath == configFile {
			return true
		}
	}

	// Check for environment-specific files
	envConfigFile := filepath.Join("environments", fmt.Sprintf("%s.yaml", cr.environment))
	if relPath == envConfigFile {
		return true
	}

	return false
}

// debouncedReload debounces configuration reloads
func (cr *ConfigReloader) debouncedReload() {
	// Wait a short time to allow for multiple file changes to complete
	time.Sleep(100 * time.Millisecond)

	// Trigger reload
	select {
	case cr.reloadChan <- nil:
	default:
		// Channel is full, reload is already pending
	}
}

// reloadLoop handles configuration reloading
func (cr *ConfigReloader) reloadLoop(ctx context.Context) {
	for {
		select {
		case <-cr.reloadChan:
			if err := cr.performReload(); err != nil {
				cr.errorChan <- fmt.Errorf("configuration reload failed: %w", err)
			}
		case <-ctx.Done():
			return
		case <-cr.stopChan:
			return
		}
	}
}

// performReload performs the actual configuration reload
func (cr *ConfigReloader) performReload() error {
	cr.mu.RLock()
	enabled := cr.enabled
	oldConfig := cr.currentConfig
	cr.mu.RUnlock()

	if !enabled {
		return nil
	}

	// Load new configuration
	newConfig, err := cr.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load new configuration: %w", err)
	}

	// Validate new configuration
	errors := cr.validator.Validate(newConfig)
	if len(errors) > 0 {
		// Log validation errors but don't fail the reload
		log.Printf("Configuration validation warnings: %s", cr.validator.GetErrorSummary())

		// Check if there are any critical errors
		if cr.validator.HasErrorsByLevel(Error) {
			return fmt.Errorf("configuration validation failed: %s", cr.validator.GetErrorSummary())
		}
	}

	// Update current configuration
	cr.mu.Lock()
	cr.currentConfig = newConfig
	cr.mu.Unlock()

	// Call reload callbacks
	if err := cr.callReloadCallbacks(oldConfig, newConfig); err != nil {
		return fmt.Errorf("failed to execute reload callbacks: %w", err)
	}

	log.Printf("Configuration reloaded successfully")
	return nil
}

// callReloadCallbacks calls all registered reload callbacks
func (cr *ConfigReloader) callReloadCallbacks(oldConfig, newConfig *EnhancedConfig) error {
	cr.mu.RLock()
	callbacks := make([]ConfigReloadCallback, len(cr.callbacks))
	copy(callbacks, cr.callbacks)
	cr.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(oldConfig, newConfig); err != nil {
			return fmt.Errorf("callback execution failed: %w", err)
		}
	}

	return nil
}

// ForceReload forces a configuration reload
func (cr *ConfigReloader) ForceReload() error {
	return cr.performReload()
}

// GetReloadErrors returns any errors that occurred during reloading
func (cr *ConfigReloader) GetReloadErrors() []error {
	var errors []error

	// Drain the error channel
	for {
		select {
		case err := <-cr.errorChan:
			errors = append(errors, err)
		default:
			return errors
		}
	}
}

// GetReloadStatus returns the status of the configuration reloader
func (cr *ConfigReloader) GetReloadStatus() map[string]interface{} {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return map[string]interface{}{
		"enabled":        cr.enabled,
		"config_path":    cr.configPath,
		"environment":    cr.environment,
		"callbacks":      len(cr.callbacks),
		"current_config": cr.currentConfig != nil,
	}
}

// ValidateCurrentConfig validates the current configuration
func (cr *ConfigReloader) ValidateCurrentConfig() []ValidationError {
	cr.mu.RLock()
	config := cr.currentConfig
	cr.mu.RUnlock()

	if config == nil {
		return []ValidationError{
			{
				Field:   "ConfigReloader",
				Value:   nil,
				Message: "no configuration loaded",
				Level:   Error,
			},
		}
	}

	return cr.validator.Validate(config)
}

// GetValidationSummary returns a summary of configuration validation
func (cr *ConfigReloader) GetValidationSummary() string {
	errors := cr.ValidateCurrentConfig()
	if len(errors) == 0 {
		return "Configuration is valid"
	}

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, err := range errors {
		switch err.Level {
		case Error:
			errorCount++
		case Warning:
			warningCount++
		case Info:
			infoCount++
		}
	}

	return fmt.Sprintf("Configuration validation found %d errors, %d warnings, %d info messages",
		errorCount, warningCount, infoCount)
}

// WatchConfigFile adds a specific file to the watcher
func (cr *ConfigReloader) WatchConfigFile(filePath string) error {
	if err := cr.watcher.Add(filePath); err != nil {
		return fmt.Errorf("failed to watch config file %s: %w", filePath, err)
	}
	return nil
}

// UnwatchConfigFile removes a specific file from the watcher
func (cr *ConfigReloader) UnwatchConfigFile(filePath string) error {
	if err := cr.watcher.Remove(filePath); err != nil {
		return fmt.Errorf("failed to unwatch config file %s: %w", filePath, err)
	}
	return nil
}

// GetWatchedFiles returns a list of currently watched files
func (cr *ConfigReloader) GetWatchedFiles() []string {
	// Note: fsnotify doesn't provide a direct way to get watched files
	// This is a simplified implementation that returns the expected files
	files := []string{
		filepath.Join(cr.configPath, "config.yaml"),
		filepath.Join(cr.configPath, "modules.yaml"),
		filepath.Join(cr.configPath, "enhanced_features.yaml"),
		filepath.Join(cr.configPath, "performance.yaml"),
		filepath.Join(cr.configPath, "advanced_monitoring.yaml"),
		filepath.Join(cr.configPath, "environments", fmt.Sprintf("%s.yaml", cr.environment)),
	}

	var watchedFiles []string
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			watchedFiles = append(watchedFiles, file)
		}
	}

	return watchedFiles
}
