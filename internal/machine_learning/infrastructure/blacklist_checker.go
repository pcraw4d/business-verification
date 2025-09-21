package infrastructure

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

// BlacklistChecker handles blacklist checking for known bad actors
type BlacklistChecker struct {
	// Blacklist databases
	businessBlacklist map[string]*BlacklistEntry
	domainBlacklist   map[string]*BlacklistEntry

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// BlacklistEntry represents a blacklist entry
type BlacklistEntry struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // business, domain, ip
	Value     string    `json:"value"`
	Reason    string    `json:"reason"`
	RiskLevel string    `json:"risk_level"` // low, medium, high, critical
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBlacklistChecker creates a new blacklist checker
func NewBlacklistChecker(logger *log.Logger) *BlacklistChecker {
	if logger == nil {
		logger = log.Default()
	}

	return &BlacklistChecker{
		businessBlacklist: make(map[string]*BlacklistEntry),
		domainBlacklist:   make(map[string]*BlacklistEntry),
		logger:            logger,
	}
}

// Initialize initializes the blacklist checker with blacklist databases
func (bc *BlacklistChecker) Initialize(ctx context.Context) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.logger.Printf("🚫 Initializing Blacklist Checker")

	// Load business blacklist
	if err := bc.loadBusinessBlacklist(); err != nil {
		return fmt.Errorf("failed to load business blacklist: %w", err)
	}

	// Load domain blacklist
	if err := bc.loadDomainBlacklist(); err != nil {
		return fmt.Errorf("failed to load domain blacklist: %w", err)
	}

	bc.logger.Printf("✅ Blacklist Checker initialized with %d business entries and %d domain entries",
		len(bc.businessBlacklist), len(bc.domainBlacklist))

	return nil
}

// CheckBlacklist checks if a business name or domain is blacklisted
func (bc *BlacklistChecker) CheckBlacklist(ctx context.Context, businessName, websiteURL string) ([]DetectedRisk, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var detectedRisks []DetectedRisk

	// Check business name against business blacklist
	if businessName != "" {
		businessRisks := bc.checkBusinessBlacklist(businessName)
		detectedRisks = append(detectedRisks, businessRisks...)
	}

	// Check domain against domain blacklist
	if websiteURL != "" {
		domainRisks := bc.checkDomainBlacklist(websiteURL)
		detectedRisks = append(detectedRisks, domainRisks...)
	}

	return detectedRisks, nil
}

// IsBlacklisted checks if a business name or domain is blacklisted
func (bc *BlacklistChecker) IsBlacklisted(businessName, websiteURL string) (bool, *BlacklistEntry) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Check business name
	if businessName != "" {
		if entry, exists := bc.businessBlacklist[strings.ToLower(businessName)]; exists {
			return true, entry
		}
	}

	// Check domain
	if websiteURL != "" {
		domain := bc.extractDomain(websiteURL)
		if domain != "" {
			if entry, exists := bc.domainBlacklist[strings.ToLower(domain)]; exists {
				return true, entry
			}
		}
	}

	return false, nil
}

// HealthCheck performs a health check on the blacklist checker
func (bc *BlacklistChecker) HealthCheck(ctx context.Context) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Check if blacklists are loaded
	if len(bc.businessBlacklist) == 0 && len(bc.domainBlacklist) == 0 {
		return fmt.Errorf("blacklists not loaded")
	}

	return nil
}

// checkBusinessBlacklist checks business name against business blacklist
func (bc *BlacklistChecker) checkBusinessBlacklist(businessName string) []DetectedRisk {
	var detectedRisks []DetectedRisk

	// Normalize business name for comparison
	normalizedName := strings.ToLower(strings.TrimSpace(businessName))

	// Check exact match
	if entry, exists := bc.businessBlacklist[normalizedName]; exists {
		risk := DetectedRisk{
			Category:    "blacklist",
			Severity:    entry.RiskLevel,
			Confidence:  1.0, // High confidence for exact matches
			Keywords:    []string{entry.Value},
			Description: fmt.Sprintf("Business blacklisted: %s (Reason: %s)", entry.Value, entry.Reason),
		}
		detectedRisks = append(detectedRisks, risk)
		return detectedRisks
	}

	// Check partial matches
	for blacklistedName, entry := range bc.businessBlacklist {
		if strings.Contains(normalizedName, blacklistedName) || strings.Contains(blacklistedName, normalizedName) {
			risk := DetectedRisk{
				Category:    "blacklist",
				Severity:    entry.RiskLevel,
				Confidence:  0.8, // Lower confidence for partial matches
				Keywords:    []string{entry.Value},
				Description: fmt.Sprintf("Business similar to blacklisted entity: %s (Reason: %s)", entry.Value, entry.Reason),
			}
			detectedRisks = append(detectedRisks, risk)
		}
	}

	return detectedRisks
}

// checkDomainBlacklist checks domain against domain blacklist
func (bc *BlacklistChecker) checkDomainBlacklist(websiteURL string) []DetectedRisk {
	var detectedRisks []DetectedRisk

	domain := bc.extractDomain(websiteURL)
	if domain == "" {
		return detectedRisks
	}

	// Normalize domain for comparison
	normalizedDomain := strings.ToLower(domain)

	// Check exact match
	if entry, exists := bc.domainBlacklist[normalizedDomain]; exists {
		risk := DetectedRisk{
			Category:    "blacklist",
			Severity:    entry.RiskLevel,
			Confidence:  1.0, // High confidence for exact matches
			Keywords:    []string{entry.Value},
			Description: fmt.Sprintf("Domain blacklisted: %s (Reason: %s)", entry.Value, entry.Reason),
		}
		detectedRisks = append(detectedRisks, risk)
		return detectedRisks
	}

	// Check subdomain matches
	for blacklistedDomain, entry := range bc.domainBlacklist {
		if strings.HasSuffix(normalizedDomain, "."+blacklistedDomain) {
			risk := DetectedRisk{
				Category:    "blacklist",
				Severity:    entry.RiskLevel,
				Confidence:  0.9, // High confidence for subdomain matches
				Keywords:    []string{entry.Value},
				Description: fmt.Sprintf("Subdomain of blacklisted domain: %s (Reason: %s)", entry.Value, entry.Reason),
			}
			detectedRisks = append(detectedRisks, risk)
		}
	}

	return detectedRisks
}

// extractDomain extracts domain from URL
func (bc *BlacklistChecker) extractDomain(websiteURL string) string {
	// Add protocol if missing
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
	}

	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return ""
	}

	return parsedURL.Hostname()
}

// loadBusinessBlacklist loads business blacklist from the database
func (bc *BlacklistChecker) loadBusinessBlacklist() error {
	// This would typically load from the Supabase database
	// For now, we'll use a sample set of blacklisted businesses

	bc.businessBlacklist = map[string]*BlacklistEntry{
		"fake business inc": {
			ID:        "bl_001",
			Type:      "business",
			Value:     "Fake Business Inc",
			Reason:    "Fraudulent business entity",
			RiskLevel: "critical",
			Source:    "manual_review",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		"scam company": {
			ID:        "bl_002",
			Type:      "business",
			Value:     "Scam Company",
			Reason:    "Known scam operation",
			RiskLevel: "critical",
			Source:    "user_report",
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
		"money launderer": {
			ID:        "bl_003",
			Type:      "business",
			Value:     "Money Launderer",
			Reason:    "Money laundering activities",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -3, 0),
			UpdatedAt: time.Now(),
		},
		"drug dealer": {
			ID:        "bl_004",
			Type:      "business",
			Value:     "Drug Dealer",
			Reason:    "Illegal drug sales",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		"weapon seller": {
			ID:        "bl_005",
			Type:      "business",
			Value:     "Weapon Seller",
			Reason:    "Illegal weapon sales",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
	}

	return nil
}

// loadDomainBlacklist loads domain blacklist from the database
func (bc *BlacklistChecker) loadDomainBlacklist() error {
	// This would typically load from the Supabase database
	// For now, we'll use a sample set of blacklisted domains

	bc.domainBlacklist = map[string]*BlacklistEntry{
		"scam-site.com": {
			ID:        "bl_dom_001",
			Type:      "domain",
			Value:     "scam-site.com",
			Reason:    "Known scam website",
			RiskLevel: "critical",
			Source:    "security_research",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		"fake-business.net": {
			ID:        "bl_dom_002",
			Type:      "domain",
			Value:     "fake-business.net",
			Reason:    "Fraudulent business website",
			RiskLevel: "critical",
			Source:    "user_report",
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
		"money-laundering.org": {
			ID:        "bl_dom_003",
			Type:      "domain",
			Value:     "money-laundering.org",
			Reason:    "Money laundering operations",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -3, 0),
			UpdatedAt: time.Now(),
		},
		"illegal-drugs.info": {
			ID:        "bl_dom_004",
			Type:      "domain",
			Value:     "illegal-drugs.info",
			Reason:    "Illegal drug sales",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		"weapon-sales.biz": {
			ID:        "bl_dom_005",
			Type:      "domain",
			Value:     "weapon-sales.biz",
			Reason:    "Illegal weapon sales",
			RiskLevel: "critical",
			Source:    "law_enforcement",
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
		"adult-entertainment.xxx": {
			ID:        "bl_dom_006",
			Type:      "domain",
			Value:     "adult-entertainment.xxx",
			Reason:    "Prohibited adult content",
			RiskLevel: "high",
			Source:    "content_policy",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		"gambling-site.com": {
			ID:        "bl_dom_007",
			Type:      "domain",
			Value:     "gambling-site.com",
			Reason:    "Prohibited gambling activities",
			RiskLevel: "high",
			Source:    "content_policy",
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
		"crypto-scam.io": {
			ID:        "bl_dom_008",
			Type:      "domain",
			Value:     "crypto-scam.io",
			Reason:    "Cryptocurrency scam",
			RiskLevel: "critical",
			Source:    "security_research",
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
	}

	return nil
}
