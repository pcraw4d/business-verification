package risk_assessment

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FreeDomainAnalyzer provides free domain analysis capabilities
type FreeDomainAnalyzer struct {
	config *RiskAssessmentConfig
	logger *zap.Logger
	client *http.Client
}

// FreeWHOISResponse represents response from free WHOIS API
type FreeWHOISResponse struct {
	DomainName     string    `json:"domain_name"`
	Registrar      string    `json:"registrar"`
	CreationDate   time.Time `json:"creation_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	UpdatedDate    time.Time `json:"updated_date"`
	Status         []string  `json:"status"`
	NameServers    []string  `json:"name_servers"`
	Registrant     struct {
		Organization string `json:"organization"`
		Name         string `json:"name"`
		Email        string `json:"email"`
		Country      string `json:"country"`
	} `json:"registrant"`
	RawData string `json:"raw_data"`
}

// NewFreeDomainAnalyzer creates a new free domain analyzer
func NewFreeDomainAnalyzer(config *RiskAssessmentConfig, logger *zap.Logger) *FreeDomainAnalyzer {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	return &FreeDomainAnalyzer{
		config: config,
		logger: logger,
		client: client,
	}
}

// AnalyzeDomainFree performs comprehensive domain analysis using only free services
func (fda *FreeDomainAnalyzer) AnalyzeDomainFree(ctx context.Context, domainName string) (*DomainAnalysisResult, error) {
	fda.logger.Info("Starting free domain analysis", zap.String("domain", domainName))

	result := &DomainAnalysisResult{
		DomainName:  domainName,
		LastUpdated: time.Now(),
	}

	// Extract clean domain name
	cleanDomain := fda.extractDomainName(domainName)

	// Update result with clean domain name
	result.DomainName = cleanDomain

	// Perform free WHOIS analysis
	whoisInfo, err := fda.performFreeWHOISLookup(ctx, cleanDomain)
	if err != nil {
		fda.logger.Warn("Free WHOIS lookup failed", zap.String("domain", cleanDomain), zap.Error(err))
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Category:    "domain",
			Factor:      "free_whois_lookup",
			Description: fmt.Sprintf("Free WHOIS lookup failed: %v", err),
			Severity:    RiskLevelMedium,
			Score:       0.5,
			Evidence:    err.Error(),
			Impact:      "Cannot verify domain registration details",
		})
	} else {
		result.WHOISInfo = whoisInfo
	}

	// Calculate domain age from WHOIS data
	if result.WHOISInfo != nil && result.WHOISInfo.CreationDate != nil {
		domainAge, err := fda.calculateDomainAge(*result.WHOISInfo.CreationDate, result.WHOISInfo.ExpirationDate)
		if err != nil {
			fda.logger.Warn("Domain age calculation failed", zap.String("domain", cleanDomain), zap.Error(err))
		} else {
			result.DomainAge = domainAge
		}
	}

	// Analyze registrar information
	if result.WHOISInfo != nil && result.WHOISInfo.Registrar != "" {
		registrarInfo := fda.analyzeRegistrarInfo(result.WHOISInfo.Registrar)
		result.RegistrarInfo = registrarInfo
	}

	// Perform free SSL certificate analysis
	sslInfo, err := fda.performFreeSSLAnalysis(ctx, cleanDomain)
	if err != nil {
		fda.logger.Warn("Free SSL analysis failed", zap.String("domain", cleanDomain), zap.Error(err))
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Category:    "domain",
			Factor:      "free_ssl_analysis",
			Description: fmt.Sprintf("Free SSL analysis failed: %v", err),
			Severity:    RiskLevelMedium,
			Score:       0.5,
			Evidence:    err.Error(),
			Impact:      "Cannot verify SSL certificate security",
		})
	} else {
		// Convert SSL info to domain analysis format
		result.DNSInfo = &DNSInfo{
			DNSScore: sslInfo.CertificateScore,
		}
	}

	// Perform comprehensive DNS analysis
	dnsInfo, err := fda.performFreeDNSAnalysis(ctx, cleanDomain)
	if err != nil {
		fda.logger.Warn("Free DNS analysis failed", zap.String("domain", cleanDomain), zap.Error(err))
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Category:    "domain",
			Factor:      "free_dns_analysis",
			Description: fmt.Sprintf("Free DNS analysis failed: %v", err),
			Severity:    RiskLevelMedium,
			Score:       0.5,
			Evidence:    err.Error(),
			Impact:      "Cannot verify DNS configuration",
		})
	} else {
		if result.DNSInfo == nil {
			result.DNSInfo = dnsInfo
		} else {
			// Merge SSL and DNS information
			result.DNSInfo.ARecords = dnsInfo.ARecords
			result.DNSInfo.AAAARecords = dnsInfo.AAAARecords
			result.DNSInfo.MXRecords = dnsInfo.MXRecords
			result.DNSInfo.NSRecords = dnsInfo.NSRecords
			result.DNSInfo.TXTRecords = dnsInfo.TXTRecords
			result.DNSInfo.DNSSECEnabled = dnsInfo.DNSSECEnabled
			result.DNSInfo.DNSScore = (result.DNSInfo.DNSScore + dnsInfo.DNSScore) / 2
		}
	}

	// Calculate overall score
	result.OverallScore = fda.calculateOverallScore(result)

	// Generate recommendations
	result.Recommendations = fda.generateRecommendations(result)

	fda.logger.Info("Free domain analysis completed",
		zap.String("domain", cleanDomain),
		zap.Float64("score", result.OverallScore))

	return result, nil
}

// performFreeWHOISLookup performs WHOIS lookup using free services
func (fda *FreeDomainAnalyzer) performFreeWHOISLookup(ctx context.Context, domainName string) (*WHOISInfo, error) {
	fda.logger.Debug("Performing free WHOIS lookup", zap.String("domain", domainName))

	// Use free WHOIS API (rate limited but free)
	url := fmt.Sprintf("https://whoisjson.com/api/v1/whois?domain=%s", domainName)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create WHOIS request: %w", err)
	}

	req.Header.Set("User-Agent", "KYB-Platform/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := fda.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WHOIS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback to direct WHOIS lookup
		return fda.performDirectWHOISLookup(ctx, domainName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read WHOIS response: %w", err)
	}

	var whoisResp FreeWHOISResponse
	if err := json.Unmarshal(body, &whoisResp); err != nil {
		// Fallback to direct WHOIS lookup
		return fda.performDirectWHOISLookup(ctx, domainName)
	}

	// Convert to internal format
	whoisInfo := &WHOISInfo{
		DomainName:     whoisResp.DomainName,
		Registrar:      whoisResp.Registrar,
		CreationDate:   &whoisResp.CreationDate,
		ExpirationDate: &whoisResp.ExpirationDate,
		UpdatedDate:    &whoisResp.UpdatedDate,
		Status:         whoisResp.Status,
		NameServers:    whoisResp.NameServers,
		RawData:        whoisResp.RawData,
	}

	// Add registrant information
	if whoisResp.Registrant.Organization != "" {
		whoisInfo.Registrant = &Registrant{
			Organization: whoisResp.Registrant.Organization,
			Name:         whoisResp.Registrant.Name,
			Email:        whoisResp.Registrant.Email,
			Country:      whoisResp.Registrant.Country,
		}
	}

	return whoisInfo, nil
}

// performDirectWHOISLookup performs direct WHOIS lookup as fallback
func (fda *FreeDomainAnalyzer) performDirectWHOISLookup(ctx context.Context, domainName string) (*WHOISInfo, error) {
	fda.logger.Debug("Performing direct WHOIS lookup", zap.String("domain", domainName))

	// Connect to WHOIS server
	conn, err := net.DialTimeout("tcp", "whois.verisign-grs.com:43", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WHOIS server: %w", err)
	}
	defer conn.Close()

	// Send WHOIS query
	query := domainName + "\r\n"
	_, err = conn.Write([]byte(query))
	if err != nil {
		return nil, fmt.Errorf("failed to send WHOIS query: %w", err)
	}

	// Read response
	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read WHOIS response: %w", err)
	}

	// Parse WHOIS response
	whoisInfo := fda.parseWHOISResponse(string(response), domainName)
	return whoisInfo, nil
}

// parseWHOISResponse parses raw WHOIS response
func (fda *FreeDomainAnalyzer) parseWHOISResponse(rawData, domainName string) *WHOISInfo {
	whoisInfo := &WHOISInfo{
		DomainName: domainName,
		RawData:    rawData,
	}

	lines := strings.Split(rawData, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "registrar":
			whoisInfo.Registrar = value
		case "creation date", "created":
			if date, err := fda.parseDate(value); err == nil {
				whoisInfo.CreationDate = &date
			}
		case "expiration date", "expires", "expiry date":
			if date, err := fda.parseDate(value); err == nil {
				whoisInfo.ExpirationDate = &date
			}
		case "updated date", "last updated":
			if date, err := fda.parseDate(value); err == nil {
				whoisInfo.UpdatedDate = &date
			}
		case "name server", "nameserver":
			whoisInfo.NameServers = append(whoisInfo.NameServers, value)
		case "status":
			whoisInfo.Status = append(whoisInfo.Status, value)
		}
	}

	return whoisInfo
}

// parseDate parses various date formats from WHOIS data
func (fda *FreeDomainAnalyzer) parseDate(dateStr string) (time.Time, error) {
	// Common WHOIS date formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-Jan-2006",
		"2006-01-02T15:04:05.000Z",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// calculateDomainAge calculates domain age from creation date
func (fda *FreeDomainAnalyzer) calculateDomainAge(creationDate time.Time, expirationDate *time.Time) (*DomainAge, error) {
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
	domainAge.AgeScore = fda.calculateAgeScore(domainAge)

	return domainAge, nil
}

// performFreeSSLAnalysis performs SSL certificate analysis using free methods
func (fda *FreeDomainAnalyzer) performFreeSSLAnalysis(ctx context.Context, domainName string) (*SSLInfo, error) {
	fda.logger.Debug("Performing free SSL analysis", zap.String("domain", domainName))

	// Create TLS connection
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		domainName+":443",
		&tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to establish TLS connection: %w", err)
	}
	defer conn.Close()

	// Get certificate
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	cert := state.PeerCertificates[0]
	now := time.Now()

	sslInfo := &SSLInfo{
		Valid:               true,
		Issuer:              cert.Issuer.CommonName,
		Subject:             cert.Subject.CommonName,
		ValidFrom:           cert.NotBefore,
		ValidTo:             cert.NotAfter,
		DaysUntilExpiration: int(cert.NotAfter.Sub(now).Hours() / 24),
		SignatureAlgorithm:  cert.SignatureAlgorithm.String(),
		TrustedByBrowser:    true,
		WildcardCertificate: strings.Contains(cert.Subject.CommonName, "*"),
		ExtendedValidation:  len(cert.Subject.Organization) > 0,
		Issues:              make([]string, 0),
	}

	// Check certificate validity
	if now.After(cert.NotAfter) {
		sslInfo.Valid = false
		sslInfo.Issues = append(sslInfo.Issues, "Certificate expired")
	} else if now.Before(cert.NotBefore) {
		sslInfo.Valid = false
		sslInfo.Issues = append(sslInfo.Issues, "Certificate not yet valid")
	}

	// Check expiration warning
	if sslInfo.DaysUntilExpiration <= 30 {
		sslInfo.Issues = append(sslInfo.Issues, "Certificate expires soon")
	}

	// Check key size
	if publicKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		sslInfo.KeySize = publicKey.N.BitLen()
		if sslInfo.KeySize < 2048 {
			sslInfo.Issues = append(sslInfo.Issues, "Weak key size")
		}
	}

	// Build certificate chain
	for _, cert := range state.PeerCertificates {
		sslInfo.CertificateChain = append(sslInfo.CertificateChain, cert.Subject.CommonName)
	}

	// Calculate certificate score
	sslInfo.CertificateScore = fda.calculateCertificateScore(sslInfo)

	return sslInfo, nil
}

// performFreeDNSAnalysis performs comprehensive DNS analysis
func (fda *FreeDomainAnalyzer) performFreeDNSAnalysis(ctx context.Context, domainName string) (*DNSInfo, error) {
	fda.logger.Debug("Performing free DNS analysis", zap.String("domain", domainName))

	dnsInfo := &DNSInfo{}

	// Resolve A records
	aRecords, err := net.LookupHost(domainName)
	if err == nil {
		dnsInfo.ARecords = aRecords
	}

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
	dnsInfo.DNSSECEnabled = fda.checkDNSSEC(domainName)

	// Calculate DNS score
	dnsInfo.DNSScore = fda.calculateDNSScore(dnsInfo)

	return dnsInfo, nil
}

// Helper methods (reused from existing domain analyzer)
func (fda *FreeDomainAnalyzer) extractDomainName(input string) string {
	domain := input
	if strings.HasPrefix(domain, "http://") {
		domain = strings.TrimPrefix(domain, "http://")
	} else if strings.HasPrefix(domain, "https://") {
		domain = strings.TrimPrefix(domain, "https://")
	}

	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove www. prefix (case-insensitive)
	domainLower := strings.ToLower(domain)
	if strings.HasPrefix(domainLower, "www.") {
		domain = domain[4:] // Remove "www." (4 characters)
	}
	return strings.ToLower(domain)
}

func (fda *FreeDomainAnalyzer) calculateAgeScore(domainAge *DomainAge) float64 {
	score := 0.5

	if domainAge.AgeInYears > 1 {
		score += 0.3
	}
	if domainAge.AgeInYears > 3 {
		score += 0.2
	}

	if domainAge.IsNewDomain {
		score -= 0.2
	}

	if domainAge.IsExpiringSoon {
		score -= 0.3
	}

	return fda.max(0.0, fda.min(1.0, score))
}

func (fda *FreeDomainAnalyzer) calculateCertificateScore(sslInfo *SSLInfo) float64 {
	score := 0.5

	if sslInfo.Valid {
		score += 0.3
	}

	if sslInfo.DaysUntilExpiration > 30 {
		score += 0.1
	}

	if sslInfo.KeySize >= 2048 {
		score += 0.1
	}

	if sslInfo.ExtendedValidation {
		score += 0.1
	}

	return fda.max(0.0, fda.min(1.0, score))
}

func (fda *FreeDomainAnalyzer) calculateDNSScore(dnsInfo *DNSInfo) float64 {
	score := 0.5

	if len(dnsInfo.ARecords) > 0 {
		score += 0.2
	}

	if len(dnsInfo.MXRecords) > 0 {
		score += 0.1
	}

	if len(dnsInfo.NSRecords) > 0 {
		score += 0.1
	}

	if dnsInfo.DNSSECEnabled {
		score += 0.1
	}

	return fda.max(0.0, fda.min(1.0, score))
}

func (fda *FreeDomainAnalyzer) checkDNSSEC(domainName string) bool {
	// Simplified DNSSEC check - in production, use proper DNSSEC validation
	return true
}

func (fda *FreeDomainAnalyzer) analyzeRegistrarInfo(registrarName string) *RegistrarInfo {
	registrarInfo := &RegistrarInfo{
		RegistrarName: registrarName,
	}

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

	if !registrarInfo.IsReputable {
		registrarInfo.RegistrarReputation = 0.5
	}

	registrarInfo.RegistrarScore = registrarInfo.RegistrarReputation
	return registrarInfo
}

func (fda *FreeDomainAnalyzer) calculateOverallScore(result *DomainAnalysisResult) float64 {
	score := 0.5

	if result.WHOISInfo != nil {
		whoisScore := 0.8
		if result.WHOISInfo.DNSSEC {
			whoisScore += 0.1
		}
		if len(result.WHOISInfo.Status) > 0 {
			whoisScore += 0.1
		}
		score += whoisScore * 0.3
	}

	if result.DomainAge != nil {
		score += result.DomainAge.AgeScore * 0.25
	}

	if result.RegistrarInfo != nil {
		score += result.RegistrarInfo.RegistrarScore * 0.2
	}

	if result.DNSInfo != nil {
		score += result.DNSInfo.DNSScore * 0.25
	}

	return fda.max(0.0, fda.min(1.0, score))
}

func (fda *FreeDomainAnalyzer) generateRecommendations(result *DomainAnalysisResult) []string {
	var recommendations []string

	if result.DomainAge != nil {
		if result.DomainAge.IsNewDomain {
			recommendations = append(recommendations, "Domain is very new (< 30 days). Consider additional verification.")
		}
		if result.DomainAge.IsExpiringSoon {
			recommendations = append(recommendations, "Domain expires soon. Verify renewal status.")
		}
	}

	if result.RegistrarInfo != nil && !result.RegistrarInfo.IsReputable {
		recommendations = append(recommendations, "Domain uses less reputable registrar. Consider transfer to established registrar.")
	}

	if result.DNSInfo != nil {
		if !result.DNSInfo.DNSSECEnabled {
			recommendations = append(recommendations, "Enable DNSSEC for enhanced security.")
		}
		if len(result.DNSInfo.MXRecords) == 0 {
			recommendations = append(recommendations, "No MX records found. Verify email configuration.")
		}
	}

	return recommendations
}

func (fda *FreeDomainAnalyzer) max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (fda *FreeDomainAnalyzer) min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
