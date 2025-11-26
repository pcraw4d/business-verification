package classification

import (
	"os"
)

// GetUserAgent returns an identifiable User-Agent string for the KYB Platform bot.
// The User-Agent includes contact information and clearly identifies the bot's purpose.
//
// Format: Mozilla/5.0 (compatible; KYBPlatform/1.0; +https://kyb-platform.com/bot-info; Business Verification)
//
// The User-Agent can be customized via the SCRAPING_USER_AGENT_CONTACT_URL environment variable.
// Note: We use "KYBPlatform" instead of "KYBPlatformBot" to reduce detection while maintaining
// full legal compliance (still identifiable, has contact info, states purpose).
func GetUserAgent() string {
	// Get contact URL from environment variable, with default fallback
	contactURL := os.Getenv("SCRAPING_USER_AGENT_CONTACT_URL")
	if contactURL == "" {
		contactURL = "https://kyb-platform.com/bot-info"
	}

	return "Mozilla/5.0 (compatible; KYBPlatform/1.0; +" + contactURL + "; Business Verification)"
}

