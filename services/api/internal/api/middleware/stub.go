package middleware

import "context"

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	// In a real implementation, this would extract user ID from JWT token or session
	return "stub-user-id"
}
