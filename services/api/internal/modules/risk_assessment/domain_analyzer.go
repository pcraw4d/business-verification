package risk_assessment

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DomainAnalyzer provides domain analysis capabilities
type DomainAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
}

// DomainAnalysisResult contains comprehensive domain analysis results
type DomainAnalysisResult struct {
	DomainName      string         `json:"domain_name"`
	WHOISInfo       *WHOISInfo     `json:"whois_info,omitempty"`
	DomainAge       *DomainAge     `json:"domain_age,omitempty"`
	RegistrarInfo   *RegistrarInfo `json:"registrar_info,omitempty"`
	DNSInfo         *DNSInfo       `json:"dns_info,omitempty"`
	OverallScore    float64        `json:"overall_score"`
	RiskFactors     []RiskFactor   `json:"risk_factors"`
	Recommendations []string       `json:"recommendations"`
	LastUpdated     time.Time      `json:"last_updated"`
}

// WHOISInfo contains WHOIS data for a domain
type WHOISInfo struct {
	DomainName     string      `json:"domain_name"`
	Registrar      string      `json:"registrar"`
	CreationDate   *time.Time  `json:"creation_date,omitempty"`
	ExpirationDate *time.Time  `json:"expiration_date,omitempty"`
	UpdatedDate    *time.Time  `json:"updated_date,omitempty"`
	Status         []string    `json:"status"`
	NameServers    []string    `json:"name_servers"`
	Registrant     *Registrant `json:"registrant,omitempty"`
	AdminContact   *Contact    `json:"admin_contact,omitempty"`
	TechContact    *Contact    `json:"tech_contact,omitempty"`
	DNSSEC         bool        `json:"dnssec"`
	RawData        string      `json:"raw_data,omitempty"`
}

// Registrant contains registrant information
type Registrant struct {
	Organization string `json:"organization,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
}

// Contact contains contact information
type Contact struct {
	Organization string `json:"organization,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
}

// DomainAge contains domain age analysis
type DomainAge struct {
	AgeInDays      int       `json:"age_in_days"`
	AgeInYears     float64   `json:"age_in_years"`
	CreationDate   time.Time `json:"creation_date"`
	IsNewDomain    bool      `json:"is_new_domain"`
	IsExpiringSoon bool      `json:"is_expiring_soon"`
	DaysToExpiry   int       `json:"days_to_expiry"`
	AgeScore       float64   `json:"age_score"`
}

// RegistrarInfo contains registrar analysis
type RegistrarInfo struct {
	RegistrarName       string  `json:"registrar_name"`
	RegistrarCountry    string  `json:"registrar_country"`
	RegistrarReputation float64 `json:"registrar_reputation"`
	IsReputable         bool    `json:"is_reputable"`
	RegistrarScore      float64 `json:"registrar_score"`
}

// DNSInfo contains DNS analysis results
type DNSInfo struct {
	ARecords      []string `json:"a_records"`
	AAAARecords   []string `json:"aaaa_records"`
	MXRecords     []string `json:"mx_records"`
	NSRecords     []string `json:"ns_records"`
	TXTRecords    []string `json:"txt_records"`
	DNSSECEnabled bool     `json:"dnssec_enabled"`
	DNSScore      float64  `json:"dns_score"`
}

// NewDomainAnalyzer creates a new domain analyzer
func NewDomainAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *DomainAnalyzer {
	return &DomainAnalyzer{
		config: config,
		logger: logger,
	}
}

// AnalyzeDomain performs comprehensive domain analysis
func (da *DomainAnalyzer) AnalyzeDomain(ctx context.Context, domainName string) (*DomainAnalysisResult, error) {
	da.logger.Info("Starting domain analysis", zap.String("domain", domainName))

	result := &DomainAnalysisResult{
		DomainName:  domainName,
		LastUpdated: time.Now(),
	}

	// Extract domain name from URL if needed
	cleanDomain := da.extractDomainName(domainName)

	// Perform WHOIS analysis if enabled
	if da.config.WHOISLookupEnabled {
		whoisInfo, err := da.analyzeWHOISData(ctx, cleanDomain)
		if err != nil {
			da.logger.Warn("WHOIS analysis failed", zap.String("domain", cleanDomain), zap.Error(err))
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Category:    "domain",
				Factor:      "whois_data_retrieval",
				Description: fmt.Sprintf("WHOIS data retrieval failed: %v", err),
				Severity:    RiskLevelMedium,
				Score:       0.5,
				Evidence:    err.Error(),
				Impact:      "Cannot verify domain registration details",
			})
		} else {
			result.WHOISInfo = whoisInfo
		}
	}

	// Analyze domain age if WHOIS data is available
	if result.WHOISInfo != nil && result.WHOISInfo.CreationDate != nil {
		domainAge, err := da.analyzeDomainAge(*result.WHOISInfo.CreationDate, result.WHOISInfo.ExpirationDate)
		if err != nil {
			da.logger.Warn("Domain age analysis failed", zap.String("domain", cleanDomain), zap.Error(err))
		} else {
			result.DomainAge = domainAge
		}
	}

	// Analyze registrar information if available
	if result.WHOISInfo != nil && result.WHOISInfo.Registrar != "" {
		registrarInfo := da.analyzeRegistrarInfo(result.WHOISInfo.Registrar)
		result.RegistrarInfo = registrarInfo
	}

	// Analyze DNS information
	dnsInfo, err := da.analyzeDNSInfo(ctx, cleanDomain)
	if err != nil {
		da.logger.Warn("DNS analysis failed", zap.String("domain", cleanDomain), zap.Error(err))
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Category:    "domain",
			Factor:      "dns_analysis",
			Description: fmt.Sprintf("DNS analysis failed: %v", err),
			Severity:    RiskLevelMedium,
			Score:       0.5,
			Evidence:    err.Error(),
			Impact:      "Cannot verify DNS configuration",
		})
	} else {
		result.DNSInfo = dnsInfo
	}

	// Calculate overall score
	result.OverallScore = da.calculateOverallScore(result)

	// Generate recommendations
	result.Recommendations = da.generateRecommendations(result)

	da.logger.Info("Domain analysis completed",
		zap.String("domain", cleanDomain),
		zap.Float64("score", result.OverallScore))

	return result, nil
}

// analyzeWHOISData retrieves and analyzes WHOIS data for a domain
func (da *DomainAnalyzer) analyzeWHOISData(ctx context.Context, domainName string) (*WHOISInfo, error) {
	da.logger.Debug("Retrieving WHOIS data", zap.String("domain", domainName))

	// In a real implementation, this would use a WHOIS client library
	// For now, we'll simulate WHOIS data retrieval
	whoisInfo := &WHOISInfo{
		DomainName: domainName,
		Registrar:  "Example Registrar, Inc.",
		Status:     []string{"clientTransferProhibited", "clientUpdateProhibited"},
		NameServers: []string{
			"ns1.example.com",
			"ns2.example.com",
		},
		DNSSEC: true,
	}

	// Simulate creation date (1-5 years ago)
	creationDate := time.Now().AddDate(-2, -6, -15) // 2.5 years ago
	whoisInfo.CreationDate = &creationDate

	// Simulate expiration date (1-10 years from now)
	expirationDate := time.Now().AddDate(7, 0, 0) // 7 years from now
	whoisInfo.ExpirationDate = &expirationDate

	// Simulate last update
	updatedDate := time.Now().AddDate(0, -3, -10) // 3 months ago
	whoisInfo.UpdatedDate = &updatedDate

	// Add registrant information
	whoisInfo.Registrant = &Registrant{
		Organization: "Example Corporation",
		Name:         "John Doe",
		Email:        "admin@example.com",
		Phone:        "+1.5551234567",
		Address:      "123 Main Street",
		City:         "Anytown",
		State:        "CA",
		Country:      "US",
		PostalCode:   "12345",
	}

	// Add admin contact
	whoisInfo.AdminContact = &Contact{
		Organization: "Example Corporation",
		Name:         "Admin Contact",
		Email:        "admin@example.com",
		Phone:        "+1.5551234567",
	}

	// Add tech contact
	whoisInfo.TechContact = &Contact{
		Organization: "Example Corporation",
		Name:         "Tech Contact",
		Email:        "tech@example.com",
		Phone:        "+1.5551234568",
	}

	// Add raw WHOIS data
	whoisInfo.RawData = da.generateRawWHOISData(whoisInfo)

	return whoisInfo, nil
}

// analyzeDomainAge calculates domain age and related metrics
func (da *DomainAnalyzer) analyzeDomainAge(creationDate time.Time, expirationDate *time.Time) (*DomainAge, error) {
	now := time.Now()
	ageInDays := int(now.Sub(creationDate).Hours() / 24)
	ageInYears := float64(ageInDays) / 365.25

	domainAge := &DomainAge{
		AgeInDays:    ageInDays,
		AgeInYears:   ageInYears,
		CreationDate: creationDate,
		IsNewDomain:  ageInDays < 30, // Less than 30 days old
	}

	// Check if domain is expiring soon
	if expirationDate != nil {
		daysToExpiry := int(expirationDate.Sub(now).Hours() / 24)
		domainAge.DaysToExpiry = daysToExpiry
		domainAge.IsExpiringSoon = daysToExpiry < 30 // Less than 30 days to expiry
	}

	// Calculate age score (0-1, higher is better)
	domainAge.AgeScore = da.calculateAgeScore(domainAge)

	return domainAge, nil
}

// analyzeRegistrarInfo analyzes registrar reputation and reliability
func (da *DomainAnalyzer) analyzeRegistrarInfo(registrarName string) *RegistrarInfo {
	// In a real implementation, this would query a registrar reputation database
	// For now, we'll use a simple heuristic based on registrar name patterns

	registrarInfo := &RegistrarInfo{
		RegistrarName: registrarName,
	}

	// Simple reputation scoring based on registrar name patterns
	reputablePatterns := []string{
		"godaddy", "namecheap", "google", "cloudflare", "name.com",
		"enom", "tucows", "network solutions", "register.com",
	}

	registrarLower := strings.ToLower(registrarName)
	for _, pattern := range reputablePatterns {
		if strings.Contains(registrarLower, pattern) {
			registrarInfo.IsReputable = true
			registrarInfo.RegistrarReputation = 0.8
			break
		}
	}

	// Default reputation for unknown registrars
	if !registrarInfo.IsReputable {
		registrarInfo.RegistrarReputation = 0.5
	}

	// Calculate registrar score
	registrarInfo.RegistrarScore = registrarInfo.RegistrarReputation

	return registrarInfo
}

// analyzeDNSInfo performs DNS analysis for the domain
func (da *DomainAnalyzer) analyzeDNSInfo(ctx context.Context, domainName string) (*DNSInfo, error) {
	dnsInfo := &DNSInfo{}

	// Resolve A records
	aRecords, err := net.LookupHost(domainName)
	if err == nil {
		dnsInfo.ARecords = aRecords
	}

	// Resolve AAAA records (IPv6)
	// Note: net.LookupHost returns both A and AAAA records
	// In a real implementation, you might want to separate them

	// Resolve MX records
	mxRecords, err := net.LookupMX(domainName)
	if err == nil {
		for _, mx := range mxRecords {
			dnsInfo.MXRecords = append(dnsInfo.MXRecords, mx.Host)
		}
	}

	// Resolve NS records
	nsRecords, err := net.LookupNS(domainName)
	if err == nil {
		for _, ns := range nsRecords {
			dnsInfo.NSRecords = append(dnsInfo.NSRecords, ns.Host)
		}
	}

	// Resolve TXT records
	txtRecords, err := net.LookupTXT(domainName)
	if err == nil {
		dnsInfo.TXTRecords = txtRecords
	}

	// Check for DNSSEC (simplified check)
	dnsInfo.DNSSECEnabled = da.checkDNSSEC(domainName)

	// Calculate DNS score
	dnsInfo.DNSScore = da.calculateDNSScore(dnsInfo)

	return dnsInfo, nil
}

// extractDomainName extracts the domain name from a URL or domain string
func (da *DomainAnalyzer) extractDomainName(input string) string {
	// Remove protocol if present
	domain := input
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}

	// Remove path and query parameters
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove www. prefix
	domain = strings.TrimPrefix(domain, "www.")

	return strings.ToLower(domain)
}

// calculateAgeScore calculates a score based on domain age
func (da *DomainAnalyzer) calculateAgeScore(domainAge *DomainAge) float64 {
	// Base score starts at 0.5 for new domains
	score := 0.5

	// Older domains get higher scores (up to 1.0)
	if domainAge.AgeInYears > 1 {
		score += 0.3
	}
	if domainAge.AgeInYears > 3 {
		score += 0.2
	}

	// Penalty for very new domains
	if domainAge.IsNewDomain {
		score -= 0.2
	}

	// Penalty for expiring domains
	if domainAge.IsExpiringSoon {
		score -= 0.3
	}

	return da.max(0.0, da.min(1.0, score))
}

// calculateDNSScore calculates a score based on DNS configuration
func (da *DomainAnalyzer) calculateDNSScore(dnsInfo *DNSInfo) float64 {
	score := 0.5 // Base score

	// Bonus for having A records
	if len(dnsInfo.ARecords) > 0 {
		score += 0.2
	}

	// Bonus for having MX records
	if len(dnsInfo.MXRecords) > 0 {
		score += 0.1
	}

	// Bonus for having NS records
	if len(dnsInfo.NSRecords) > 0 {
		score += 0.1
	}

	// Bonus for DNSSEC
	if dnsInfo.DNSSECEnabled {
		score += 0.1
	}

	return da.max(0.0, da.min(1.0, score))
}

// checkDNSSEC performs a simplified DNSSEC check
func (da *DomainAnalyzer) checkDNSSEC(domainName string) bool {
	// In a real implementation, this would perform actual DNSSEC validation
	// For now, we'll simulate DNSSEC being enabled for most domains
	return true
}

// calculateOverallScore calculates the overall domain analysis score
func (da *DomainAnalyzer) calculateOverallScore(result *DomainAnalysisResult) float64 {
	score := 0.5 // Base score

	// WHOIS score (30% weight)
	if result.WHOISInfo != nil {
		whoisScore := 0.8 // Base WHOIS score
		if result.WHOISInfo.DNSSEC {
			whoisScore += 0.1
		}
		if len(result.WHOISInfo.Status) > 0 {
			whoisScore += 0.1
		}
		score += whoisScore * 0.3
	}

	// Domain age score (25% weight)
	if result.DomainAge != nil {
		score += result.DomainAge.AgeScore * 0.25
	}

	// Registrar score (20% weight)
	if result.RegistrarInfo != nil {
		score += result.RegistrarInfo.RegistrarScore * 0.2
	}

	// DNS score (25% weight)
	if result.DNSInfo != nil {
		score += result.DNSInfo.DNSScore * 0.25
	}

	return da.max(0.0, da.min(1.0, score))
}

// generateRecommendations generates recommendations based on analysis results
func (da *DomainAnalyzer) generateRecommendations(result *DomainAnalysisResult) []string {
	var recommendations []string

	// Domain age recommendations
	if result.DomainAge != nil {
		if result.DomainAge.IsNewDomain {
			recommendations = append(recommendations, "Domain is very new (< 30 days). Consider additional verification.")
		}
		if result.DomainAge.IsExpiringSoon {
			recommendations = append(recommendations, "Domain expires soon. Verify renewal status.")
		}
	}

	// Registrar recommendations
	if result.RegistrarInfo != nil && !result.RegistrarInfo.IsReputable {
		recommendations = append(recommendations, "Domain uses less reputable registrar. Consider transfer to established registrar.")
	}

	// DNS recommendations
	if result.DNSInfo != nil {
		if !result.DNSInfo.DNSSECEnabled {
			recommendations = append(recommendations, "Enable DNSSEC for enhanced security.")
		}
		if len(result.DNSInfo.MXRecords) == 0 {
			recommendations = append(recommendations, "No MX records found. Verify email configuration.")
		}
	}

	// WHOIS recommendations
	if result.WHOISInfo != nil {
		if result.WHOISInfo.Registrant != nil && result.WHOISInfo.Registrant.Email == "" {
			recommendations = append(recommendations, "Registrant email not found in WHOIS data.")
		}
	}

	return recommendations
}

// generateRawWHOISData generates simulated raw WHOIS data
func (da *DomainAnalyzer) generateRawWHOISData(whoisInfo *WHOISInfo) string {
	return fmt.Sprintf(`Domain Name: %s
Registrar: %s
Creation Date: %s
Expiration Date: %s
Updated Date: %s
Status: %s
Name Servers: %s
DNSSEC: %t
Registrant Organization: %s
Registrant Name: %s
Registrant Email: %s`,
		whoisInfo.DomainName,
		whoisInfo.Registrar,
		whoisInfo.CreationDate.Format("2006-01-02"),
		whoisInfo.ExpirationDate.Format("2006-01-02"),
		whoisInfo.UpdatedDate.Format("2006-01-02"),
		strings.Join(whoisInfo.Status, ", "),
		strings.Join(whoisInfo.NameServers, ", "),
		whoisInfo.DNSSEC,
		whoisInfo.Registrant.Organization,
		whoisInfo.Registrant.Name,
		whoisInfo.Registrant.Email)
}

// Helper functions
func (da *DomainAnalyzer) max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (da *DomainAnalyzer) min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
