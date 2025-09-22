package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"kyb-platform/internal/api/compatibility"
	"go.uber.org/zap"
)

// APIVersioningConfig holds configuration for API versioning middleware
type APIVersioningConfig struct {
	// Version Detection
	EnableURLVersioning    bool `json:"enable_url_versioning" yaml:"enable_url_versioning"`
	EnableHeaderVersioning bool `json:"enable_header_versioning" yaml:"enable_header_versioning"`
	EnableQueryVersioning  bool `json:"enable_query_versioning" yaml:"enable_query_versioning"`
	EnableAcceptVersioning bool `json:"enable_accept_versioning" yaml:"enable_accept_versioning"`

	// Version Headers
	VersionHeaderName   string `json:"version_header_name" yaml:"version_header_name"`
	QueryVersionParam   string `json:"query_version_param" yaml:"query_version_param"`
	AcceptVersionPrefix string `json:"accept_version_prefix" yaml:"accept_version_prefix"`

	// Behavior
	StrictVersioning      bool `json:"strict_versioning" yaml:"strict_versioning"`
	AllowVersionFallback  bool `json:"allow_version_fallback" yaml:"allow_version_fallback"`
	RemoveVersionFromPath bool `json:"remove_version_from_path" yaml:"remove_version_from_path"`

	// Error Handling
	ReturnVersionErrors bool `json:"return_version_errors" yaml:"return_version_errors"`
	LogVersionFailures  bool `json:"log_version_failures" yaml:"log_version_failures"`

	// Deprecation
	EnableDeprecationWarnings bool `json:"enable_deprecation_warnings" yaml:"enable_deprecation_warnings"`
	DeprecationWarningDays    int  `json:"deprecation_warning_days" yaml:"deprecation_warning_days"`

	// Client Validation
	EnableClientValidation bool   `json:"enable_client_validation" yaml:"enable_client_validation"`
	ClientVersionHeader    string `json:"client_version_header" yaml:"client_version_header"`
}

// APIVersioningMiddleware provides comprehensive API versioning
type APIVersioningMiddleware struct {
	config         *APIVersioningConfig
	versionManager *compatibility.VersionManager
	logger         *zap.Logger
	versionRegex   *regexp.Regexp
}

// VersionInfo holds version information for the request
type VersionInfo struct {
	RequestedVersion string
	ResolvedVersion  string
	IsDeprecated     bool
	DeprecationDate  *time.Time
	SunsetDate       *time.Time
	MigrationGuide   string
	ClientVersion    string
	IsValidClient    bool
}

// VersionError represents a version-related error
type VersionError struct {
	Type              string   `json:"type"`
	Message           string   `json:"message"`
	Code              string   `json:"code"`
	RequestedVersion  string   `json:"requested_version,omitempty"`
	SupportedVersions []string `json:"supported_versions,omitempty"`
	MigrationGuide    string   `json:"migration_guide,omitempty"`
}

func (e *VersionError) Error() string {
	return e.Message
}

// NewAPIVersioningMiddleware creates a new API versioning middleware
func NewAPIVersioningMiddleware(config *APIVersioningConfig, versionManager *compatibility.VersionManager, logger *zap.Logger) *APIVersioningMiddleware {
	if config == nil {
		config = GetDefaultAPIVersioningConfig()
	}

	if versionManager == nil {
		panic("version manager cannot be nil")
	}

	// Compile version regex
	versionRegex := regexp.MustCompile(`^/v(\d+)(/.*)?$`)

	return &APIVersioningMiddleware{
		config:         config,
		versionManager: versionManager,
		logger:         logger,
		versionRegex:   versionRegex,
	}
}

// Middleware applies API versioning to HTTP requests
func (m *APIVersioningMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Detect and validate version
		versionInfo, err := m.detectAndValidateVersion(r)
		if err != nil {
			m.handleVersionError(w, r, err)
			return
		}

		// Add version info to context
		ctx := context.WithValue(r.Context(), "version_info", versionInfo)
		ctx = context.WithValue(ctx, "api_version", versionInfo.ResolvedVersion)

		// Rewrite path if needed
		if m.config.RemoveVersionFromPath {
			r.URL.Path = m.removeVersionFromPath(r.URL.Path)
		}

		// Add version headers to response
		m.addVersionHeaders(w, versionInfo)

		// Add deprecation warnings if needed
		if m.config.EnableDeprecationWarnings && versionInfo.IsDeprecated {
			m.addDeprecationWarnings(w, versionInfo)
		}

		// Continue with next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// detectAndValidateVersion detects and validates the API version
func (m *APIVersioningMiddleware) detectAndValidateVersion(r *http.Request) (*VersionInfo, error) {
	var requestedVersion string
	var detectionMethod string

	// Try different detection methods in order of preference
	if m.config.EnableURLVersioning {
		if version := m.extractVersionFromURL(r.URL.Path); version != "" {
			requestedVersion = version
			detectionMethod = "url"
		}
	}

	if requestedVersion == "" && m.config.EnableHeaderVersioning {
		if version := r.Header.Get(m.config.VersionHeaderName); version != "" {
			requestedVersion = version
			detectionMethod = "header"
		}
	}

	if requestedVersion == "" && m.config.EnableQueryVersioning {
		if version := r.URL.Query().Get(m.config.QueryVersionParam); version != "" {
			requestedVersion = version
			detectionMethod = "query"
		}
	}

	if requestedVersion == "" && m.config.EnableAcceptVersioning {
		if version := m.extractVersionFromAccept(r.Header.Get("Accept")); version != "" {
			requestedVersion = version
			detectionMethod = "accept"
		}
	}

	// If no version detected, use default
	if requestedVersion == "" {
		requestedVersion = m.versionManager.GetDefaultVersion(r.Context())
		detectionMethod = "default"
	}

	// Validate version
	resolvedVersion, err := m.resolveVersion(requestedVersion)
	if err != nil {
		return nil, &VersionError{
			Type:              "unsupported_version",
			Message:           fmt.Sprintf("Unsupported API version: %s", requestedVersion),
			Code:              "UNSUPPORTED_VERSION",
			RequestedVersion:  requestedVersion,
			SupportedVersions: m.getSupportedVersions(),
		}
	}

	// Get version info
	apiVersion, err := m.versionManager.GetVersion(r.Context(), resolvedVersion)
	if err != nil {
		return nil, &VersionError{
			Type:             "invalid_version",
			Message:          fmt.Sprintf("Invalid API version: %s", resolvedVersion),
			Code:             "INVALID_VERSION",
			RequestedVersion: requestedVersion,
		}
	}

	// Validate client version if enabled
	var clientVersion string
	var isValidClient bool
	if m.config.EnableClientValidation {
		clientVersion = r.Header.Get(m.config.ClientVersionHeader)
		if clientVersion != "" {
			err := m.versionManager.ValidateClientVersion(r.Context(), resolvedVersion, clientVersion)
			isValidClient = err == nil
		}
	}

	versionInfo := &VersionInfo{
		RequestedVersion: requestedVersion,
		ResolvedVersion:  resolvedVersion,
		IsDeprecated:     apiVersion.DeprecatedAt != nil,
		DeprecationDate:  apiVersion.DeprecatedAt,
		MigrationGuide:   apiVersion.MigrationGuide,
		ClientVersion:    clientVersion,
		IsValidClient:    isValidClient,
	}

	// Calculate sunset date if deprecated
	if apiVersion.DeprecatedAt != nil {
		sunsetDate := apiVersion.DeprecatedAt.Add(6 * 30 * 24 * time.Hour) // 6 months
		versionInfo.SunsetDate = &sunsetDate
	}

	// Log version detection
	if m.config.LogVersionFailures {
		m.logger.Info("API version detected",
			zap.String("requested_version", requestedVersion),
			zap.String("resolved_version", resolvedVersion),
			zap.String("detection_method", detectionMethod),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Bool("is_deprecated", versionInfo.IsDeprecated),
		)
	}

	return versionInfo, nil
}

// extractVersionFromURL extracts version from URL path
func (m *APIVersioningMiddleware) extractVersionFromURL(path string) string {
	matches := m.versionRegex.FindStringSubmatch(path)
	if len(matches) >= 2 {
		return "v" + matches[1]
	}
	return ""
}

// extractVersionFromAccept extracts version from Accept header
func (m *APIVersioningMiddleware) extractVersionFromAccept(acceptHeader string) string {
	if acceptHeader == "" {
		return ""
	}

	parts := strings.Split(acceptHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, m.config.AcceptVersionPrefix) {
			// Extract version from Accept header like: application/vnd.kyb-platform.v2+json
			start := strings.Index(part, m.config.AcceptVersionPrefix)
			if start != -1 {
				versionStart := start + len(m.config.AcceptVersionPrefix)
				versionEnd := strings.Index(part[versionStart:], "+")
				if versionEnd == -1 {
					versionEnd = strings.Index(part[versionStart:], ";")
				}
				if versionEnd == -1 {
					versionEnd = len(part[versionStart:])
				}
				return "v" + part[versionStart:versionStart+versionEnd]
			}
		}
	}
	return ""
}

// resolveVersion resolves the requested version to a supported version
func (m *APIVersioningMiddleware) resolveVersion(requestedVersion string) (string, error) {
	// Check if version is supported
	if m.versionManager.IsVersionSupported(context.Background(), requestedVersion) {
		return requestedVersion, nil
	}

	// If strict versioning is enabled, return error
	if m.config.StrictVersioning {
		return "", fmt.Errorf("version %s is not supported", requestedVersion)
	}

	// If fallback is allowed, try to find a compatible version
	if m.config.AllowVersionFallback {
		// Try current version first
		currentVersion := m.versionManager.GetCurrentVersion(context.Background())
		if m.versionManager.IsVersionSupported(context.Background(), currentVersion) {
			return currentVersion, nil
		}

		// Try default version
		defaultVersion := m.versionManager.GetDefaultVersion(context.Background())
		if m.versionManager.IsVersionSupported(context.Background(), defaultVersion) {
			return defaultVersion, nil
		}
	}

	return "", fmt.Errorf("no compatible version found for %s", requestedVersion)
}

// removeVersionFromPath removes version prefix from URL path
func (m *APIVersioningMiddleware) removeVersionFromPath(path string) string {
	matches := m.versionRegex.FindStringSubmatch(path)
	if len(matches) >= 3 && matches[2] != "" {
		return matches[2] // Return the part after version
	}
	return path
}

// addVersionHeaders adds version-related headers to response
func (m *APIVersioningMiddleware) addVersionHeaders(w http.ResponseWriter, versionInfo *VersionInfo) {
	w.Header().Set("X-API-Version", versionInfo.ResolvedVersion)
	w.Header().Set("X-API-Version-Requested", versionInfo.RequestedVersion)

	if versionInfo.IsDeprecated {
		w.Header().Set("X-API-Deprecated", "true")
		if versionInfo.DeprecationDate != nil {
			w.Header().Set("X-API-Deprecated-At", versionInfo.DeprecationDate.Format(time.RFC3339))
		}
		if versionInfo.SunsetDate != nil {
			w.Header().Set("X-API-Sunset-Date", versionInfo.SunsetDate.Format(time.RFC3339))
		}
		if versionInfo.MigrationGuide != "" {
			w.Header().Set("X-API-Migration-Guide", versionInfo.MigrationGuide)
		}
	}

	if versionInfo.ClientVersion != "" {
		w.Header().Set("X-Client-Version", versionInfo.ClientVersion)
		if !versionInfo.IsValidClient {
			w.Header().Set("X-Client-Version-Warning", "Client version may not be fully compatible")
		}
	}
}

// addDeprecationWarnings adds deprecation warnings to response
func (m *APIVersioningMiddleware) addDeprecationWarnings(w http.ResponseWriter, versionInfo *VersionInfo) {
	if versionInfo.DeprecationDate == nil {
		return
	}

	// Check if we should show warning based on deprecation warning days
	daysSinceDeprecation := int(time.Since(*versionInfo.DeprecationDate).Hours() / 24)
	if daysSinceDeprecation <= m.config.DeprecationWarningDays {
		warning := fmt.Sprintf("API version %s is deprecated and will be removed on %s",
			versionInfo.ResolvedVersion,
			versionInfo.SunsetDate.Format("2006-01-02"))

		w.Header().Set("X-API-Deprecation-Warning", warning)
	}
}

// handleVersionError handles version-related errors
func (m *APIVersioningMiddleware) handleVersionError(w http.ResponseWriter, r *http.Request, err error) {
	if m.config.LogVersionFailures {
		m.logger.Warn("API version error",
			zap.Error(err),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
		)
	}

	if !m.config.ReturnVersionErrors {
		// Return 404 for unsupported versions
		http.NotFound(w, r)
		return
	}

	// Return detailed error response
	var versionError *VersionError
	if ve, ok := err.(*VersionError); ok {
		versionError = ve
	} else {
		versionError = &VersionError{
			Type:    "version_error",
			Message: err.Error(),
			Code:    "VERSION_ERROR",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":   versionError,
		"success": false,
	}

	json.NewEncoder(w).Encode(response)
}

// getSupportedVersions returns list of supported versions
func (m *APIVersioningMiddleware) getSupportedVersions() []string {
	versions, err := m.versionManager.ListVersions(context.Background())
	if err != nil {
		return []string{}
	}

	var supported []string
	for _, version := range versions {
		if m.versionManager.IsVersionSupported(context.Background(), version.Version) {
			supported = append(supported, version.Version)
		}
	}

	return supported
}

// GetVersionInfo retrieves version info from request context
func GetVersionInfo(ctx context.Context) *VersionInfo {
	if versionInfo, ok := ctx.Value("version_info").(*VersionInfo); ok {
		return versionInfo
	}
	return nil
}

// GetAPIVersion retrieves API version from request context
func GetAPIVersion(ctx context.Context) string {
	if version, ok := ctx.Value("api_version").(string); ok {
		return version
	}
	return ""
}

// GetDefaultAPIVersioningConfig returns a default API versioning configuration
func GetDefaultAPIVersioningConfig() *APIVersioningConfig {
	return &APIVersioningConfig{
		EnableURLVersioning:       true,
		EnableHeaderVersioning:    true,
		EnableQueryVersioning:     true,
		EnableAcceptVersioning:    true,
		VersionHeaderName:         "X-API-Version",
		QueryVersionParam:         "version",
		AcceptVersionPrefix:       "vnd.kyb-platform.v",
		StrictVersioning:          false,
		AllowVersionFallback:      true,
		RemoveVersionFromPath:     true,
		ReturnVersionErrors:       true,
		LogVersionFailures:        true,
		EnableDeprecationWarnings: true,
		DeprecationWarningDays:    30,
		EnableClientValidation:    true,
		ClientVersionHeader:       "X-Client-Version",
	}
}

// GetStrictAPIVersioningConfig returns a strict API versioning configuration
func GetStrictAPIVersioningConfig() *APIVersioningConfig {
	return &APIVersioningConfig{
		EnableURLVersioning:       true,
		EnableHeaderVersioning:    true,
		EnableQueryVersioning:     false,
		EnableAcceptVersioning:    true,
		VersionHeaderName:         "X-API-Version",
		QueryVersionParam:         "version",
		AcceptVersionPrefix:       "vnd.kyb-platform.v",
		StrictVersioning:          true,
		AllowVersionFallback:      false,
		RemoveVersionFromPath:     true,
		ReturnVersionErrors:       true,
		LogVersionFailures:        true,
		EnableDeprecationWarnings: true,
		DeprecationWarningDays:    7,
		EnableClientValidation:    true,
		ClientVersionHeader:       "X-Client-Version",
	}
}

// GetPermissiveAPIVersioningConfig returns a permissive API versioning configuration
func GetPermissiveAPIVersioningConfig() *APIVersioningConfig {
	return &APIVersioningConfig{
		EnableURLVersioning:       true,
		EnableHeaderVersioning:    true,
		EnableQueryVersioning:     true,
		EnableAcceptVersioning:    true,
		VersionHeaderName:         "X-API-Version",
		QueryVersionParam:         "version",
		AcceptVersionPrefix:       "vnd.kyb-platform.v",
		StrictVersioning:          false,
		AllowVersionFallback:      true,
		RemoveVersionFromPath:     false,
		ReturnVersionErrors:       false,
		LogVersionFailures:        false,
		EnableDeprecationWarnings: false,
		DeprecationWarningDays:    0,
		EnableClientValidation:    false,
		ClientVersionHeader:       "X-Client-Version",
	}
}
