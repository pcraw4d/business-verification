package test

import (
	"context"
)

// MockKeywordRepository implements KeywordRepository for testing
type MockKeywordRepository struct {
	keywords            map[int][]*IndustryKeyword
	classificationCodes map[int][]*ClassificationCode
	codesByType         map[string][]*ClassificationCode
}

// NewMockKeywordRepository creates a new mock repository with sample data
func NewMockKeywordRepository() *MockKeywordRepository {
	mock := &MockKeywordRepository{
		keywords:            make(map[int][]*IndustryKeyword),
		classificationCodes: make(map[int][]*ClassificationCode),
		codesByType:         make(map[string][]*ClassificationCode),
	}

	// Initialize with sample data
	mock.initializeSampleData()

	return mock
}

// GetKeywordsByIndustry returns keywords for a specific industry
func (m *MockKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error) {
	keywords, exists := m.keywords[industryID]
	if !exists {
		return []*IndustryKeyword{}, nil
	}
	return keywords, nil
}

// GetClassificationCodesByIndustry returns classification codes for a specific industry
func (m *MockKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error) {
	codes, exists := m.classificationCodes[industryID]
	if !exists {
		return []*ClassificationCode{}, nil
	}
	return codes, nil
}

// GetClassificationCodesByType returns classification codes by type
func (m *MockKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	codes, exists := m.codesByType[codeType]
	if !exists {
		return []*ClassificationCode{}, nil
	}
	return codes, nil
}

// initializeSampleData initializes the mock repository with sample data
func (m *MockKeywordRepository) initializeSampleData() {
	// Technology industry (ID: 1)
	m.keywords[1] = []*IndustryKeyword{
		{ID: 1, IndustryID: 1, Keyword: "software", Weight: 1.0, IsActive: true},
		{ID: 2, IndustryID: 1, Keyword: "technology", Weight: 1.0, IsActive: true},
		{ID: 3, IndustryID: 1, Keyword: "cloud", Weight: 0.9, IsActive: true},
		{ID: 4, IndustryID: 1, Keyword: "development", Weight: 0.8, IsActive: true},
		{ID: 5, IndustryID: 1, Keyword: "AI", Weight: 0.9, IsActive: true},
		{ID: 6, IndustryID: 1, Keyword: "artificial intelligence", Weight: 0.9, IsActive: true},
		{ID: 7, IndustryID: 1, Keyword: "machine learning", Weight: 0.8, IsActive: true},
		{ID: 8, IndustryID: 1, Keyword: "platform", Weight: 0.7, IsActive: true},
		{ID: 9, IndustryID: 1, Keyword: "digital", Weight: 0.6, IsActive: true},
		{ID: 10, IndustryID: 1, Keyword: "tech", Weight: 0.5, IsActive: true},
	}

	m.classificationCodes[1] = []*ClassificationCode{
		{ID: 1, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores", IsActive: true},
		{ID: 2, IndustryID: 1, CodeType: "MCC", Code: "7372", Description: "Computer Programming Services", IsActive: true},
		{ID: 3, IndustryID: 1, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software", IsActive: true},
		{ID: 4, IndustryID: 1, CodeType: "SIC", Code: "7371", Description: "Computer Programming Services", IsActive: true},
		{ID: 5, IndustryID: 1, CodeType: "NAICS", Code: "541511", Description: "Custom Computer Programming Services", IsActive: true},
		{ID: 6, IndustryID: 1, CodeType: "NAICS", Code: "511210", Description: "Software Publishers", IsActive: true},
	}

	// Healthcare industry (ID: 2)
	m.keywords[2] = []*IndustryKeyword{
		{ID: 11, IndustryID: 2, Keyword: "healthcare", Weight: 1.0, IsActive: true},
		{ID: 12, IndustryID: 2, Keyword: "medical", Weight: 1.0, IsActive: true},
		{ID: 13, IndustryID: 2, Keyword: "clinic", Weight: 0.9, IsActive: true},
		{ID: 14, IndustryID: 2, Keyword: "hospital", Weight: 0.9, IsActive: true},
		{ID: 15, IndustryID: 2, Keyword: "patient", Weight: 0.8, IsActive: true},
		{ID: 16, IndustryID: 2, Keyword: "pharmacy", Weight: 0.7, IsActive: true},
		{ID: 17, IndustryID: 2, Keyword: "pharmaceutical", Weight: 0.9, IsActive: true},
		{ID: 18, IndustryID: 2, Keyword: "drug", Weight: 0.8, IsActive: true},
		{ID: 19, IndustryID: 2, Keyword: "research", Weight: 0.7, IsActive: true},
		{ID: 20, IndustryID: 2, Keyword: "devices", Weight: 0.8, IsActive: true},
	}

	m.classificationCodes[2] = []*ClassificationCode{
		{ID: 7, IndustryID: 2, CodeType: "MCC", Code: "8011", Description: "Doctors", IsActive: true},
		{ID: 8, IndustryID: 2, CodeType: "MCC", Code: "8021", Description: "Dentists", IsActive: true},
		{ID: 9, IndustryID: 2, CodeType: "SIC", Code: "8011", Description: "Offices and Clinics of Doctors of Medicine", IsActive: true},
		{ID: 10, IndustryID: 2, CodeType: "SIC", Code: "8062", Description: "General Medical and Surgical Hospitals", IsActive: true},
		{ID: 11, IndustryID: 2, CodeType: "NAICS", Code: "621111", Description: "Offices of Physicians", IsActive: true},
		{ID: 12, IndustryID: 2, CodeType: "NAICS", Code: "622110", Description: "General Medical and Surgical Hospitals", IsActive: true},
	}

	// Finance industry (ID: 3)
	m.keywords[3] = []*IndustryKeyword{
		{ID: 21, IndustryID: 3, Keyword: "banking", Weight: 1.0, IsActive: true},
		{ID: 22, IndustryID: 3, Keyword: "finance", Weight: 1.0, IsActive: true},
		{ID: 23, IndustryID: 3, Keyword: "loans", Weight: 0.9, IsActive: true},
		{ID: 24, IndustryID: 3, Keyword: "investment", Weight: 0.8, IsActive: true},
		{ID: 25, IndustryID: 3, Keyword: "credit", Weight: 0.8, IsActive: true},
		{ID: 26, IndustryID: 3, Keyword: "insurance", Weight: 0.9, IsActive: true},
		{ID: 27, IndustryID: 3, Keyword: "fintech", Weight: 0.9, IsActive: true},
		{ID: 28, IndustryID: 3, Keyword: "digital banking", Weight: 0.8, IsActive: true},
		{ID: 29, IndustryID: 3, Keyword: "payment", Weight: 0.7, IsActive: true},
		{ID: 30, IndustryID: 3, Keyword: "deposits", Weight: 0.7, IsActive: true},
	}

	m.classificationCodes[3] = []*ClassificationCode{
		{ID: 13, IndustryID: 3, CodeType: "MCC", Code: "6011", Description: "Automated Teller Machine Services", IsActive: true},
		{ID: 14, IndustryID: 3, CodeType: "MCC", Code: "6012", Description: "Financial Institutions", IsActive: true},
		{ID: 15, IndustryID: 3, CodeType: "SIC", Code: "6021", Description: "National Commercial Banks", IsActive: true},
		{ID: 16, IndustryID: 3, CodeType: "SIC", Code: "6022", Description: "State Commercial Banks", IsActive: true},
		{ID: 17, IndustryID: 3, CodeType: "NAICS", Code: "522110", Description: "Commercial Banking", IsActive: true},
		{ID: 18, IndustryID: 3, CodeType: "NAICS", Code: "523110", Description: "Investment Banking and Securities Dealing", IsActive: true},
	}

	// Retail industry (ID: 4)
	m.keywords[4] = []*IndustryKeyword{
		{ID: 31, IndustryID: 4, Keyword: "retail", Weight: 1.0, IsActive: true},
		{ID: 32, IndustryID: 4, Keyword: "shopping", Weight: 0.9, IsActive: true},
		{ID: 33, IndustryID: 4, Keyword: "consumer", Weight: 0.8, IsActive: true},
		{ID: 34, IndustryID: 4, Keyword: "goods", Weight: 0.7, IsActive: true},
		{ID: 35, IndustryID: 4, Keyword: "e-commerce", Weight: 0.9, IsActive: true},
		{ID: 36, IndustryID: 4, Keyword: "marketplace", Weight: 0.8, IsActive: true},
		{ID: 37, IndustryID: 4, Keyword: "online", Weight: 0.7, IsActive: true},
		{ID: 38, IndustryID: 4, Keyword: "store", Weight: 0.6, IsActive: true},
		{ID: 39, IndustryID: 4, Keyword: "shop", Weight: 0.5, IsActive: true},
		{ID: 40, IndustryID: 4, Keyword: "merchandise", Weight: 0.5, IsActive: true},
	}

	m.classificationCodes[4] = []*ClassificationCode{
		{ID: 19, IndustryID: 4, CodeType: "MCC", Code: "5310", Description: "Department Stores", IsActive: true},
		{ID: 20, IndustryID: 4, CodeType: "MCC", Code: "5732", Description: "Electronics Stores", IsActive: true},
		{ID: 21, IndustryID: 4, CodeType: "SIC", Code: "5311", Description: "Department Stores", IsActive: true},
		{ID: 22, IndustryID: 4, CodeType: "SIC", Code: "5731", Description: "Radio, Television, and Consumer Electronics Stores", IsActive: true},
		{ID: 23, IndustryID: 4, CodeType: "NAICS", Code: "454110", Description: "Electronic Shopping and Mail-Order Houses", IsActive: true},
		{ID: 24, IndustryID: 4, CodeType: "NAICS", Code: "452111", Description: "Department Stores", IsActive: true},
	}

	// Manufacturing industry (ID: 5)
	m.keywords[5] = []*IndustryKeyword{
		{ID: 41, IndustryID: 5, Keyword: "manufacturing", Weight: 1.0, IsActive: true},
		{ID: 42, IndustryID: 5, Keyword: "factory", Weight: 0.9, IsActive: true},
		{ID: 43, IndustryID: 5, Keyword: "production", Weight: 0.8, IsActive: true},
		{ID: 44, IndustryID: 5, Keyword: "industrial", Weight: 0.7, IsActive: true},
		{ID: 45, IndustryID: 5, Keyword: "equipment", Weight: 0.8, IsActive: true},
		{ID: 46, IndustryID: 5, Keyword: "machinery", Weight: 0.8, IsActive: true},
		{ID: 47, IndustryID: 5, Keyword: "automation", Weight: 0.7, IsActive: true},
		{ID: 48, IndustryID: 5, Keyword: "food processing", Weight: 0.9, IsActive: true},
		{ID: 49, IndustryID: 5, Keyword: "packaged foods", Weight: 0.8, IsActive: true},
		{ID: 50, IndustryID: 5, Keyword: "beverages", Weight: 0.7, IsActive: true},
	}

	m.classificationCodes[5] = []*ClassificationCode{
		{ID: 25, IndustryID: 5, CodeType: "MCC", Code: "5085", Description: "Industrial Supplies", IsActive: true},
		{ID: 26, IndustryID: 5, CodeType: "MCC", Code: "5047", Description: "Medical Equipment", IsActive: true},
		{ID: 27, IndustryID: 5, CodeType: "SIC", Code: "3531", Description: "Construction Machinery and Equipment", IsActive: true},
		{ID: 28, IndustryID: 5, CodeType: "SIC", Code: "3532", Description: "Mining Machinery and Equipment", IsActive: true},
		{ID: 29, IndustryID: 5, CodeType: "NAICS", Code: "333120", Description: "Construction Machinery Manufacturing", IsActive: true},
		{ID: 30, IndustryID: 5, CodeType: "NAICS", Code: "332996", Description: "Miscellaneous Fabricated Metal Product Manufacturing", IsActive: true},
	}

	// Professional Services industry (ID: 6)
	m.keywords[6] = []*IndustryKeyword{
		{ID: 51, IndustryID: 6, Keyword: "consulting", Weight: 1.0, IsActive: true},
		{ID: 52, IndustryID: 6, Keyword: "management", Weight: 0.9, IsActive: true},
		{ID: 53, IndustryID: 6, Keyword: "strategy", Weight: 0.8, IsActive: true},
		{ID: 54, IndustryID: 6, Keyword: "business", Weight: 0.7, IsActive: true},
		{ID: 55, IndustryID: 6, Keyword: "legal", Weight: 1.0, IsActive: true},
		{ID: 56, IndustryID: 6, Keyword: "law", Weight: 0.9, IsActive: true},
		{ID: 57, IndustryID: 6, Keyword: "litigation", Weight: 0.8, IsActive: true},
		{ID: 58, IndustryID: 6, Keyword: "corporate", Weight: 0.7, IsActive: true},
		{ID: 59, IndustryID: 6, Keyword: "services", Weight: 0.6, IsActive: true},
		{ID: 60, IndustryID: 6, Keyword: "professional", Weight: 0.6, IsActive: true},
	}

	m.classificationCodes[6] = []*ClassificationCode{
		{ID: 31, IndustryID: 6, CodeType: "MCC", Code: "7392", Description: "Management Consulting", IsActive: true},
		{ID: 32, IndustryID: 6, CodeType: "MCC", Code: "8999", Description: "Professional Services", IsActive: true},
		{ID: 33, IndustryID: 6, CodeType: "SIC", Code: "8742", Description: "Management Consulting Services", IsActive: true},
		{ID: 34, IndustryID: 6, CodeType: "SIC", Code: "8741", Description: "Management Services", IsActive: true},
		{ID: 35, IndustryID: 6, CodeType: "NAICS", Code: "541611", Description: "Administrative Management and General Management Consulting Services", IsActive: true},
		{ID: 36, IndustryID: 6, CodeType: "NAICS", Code: "541612", Description: "Human Resources Consulting Services", IsActive: true},
	}

	// Real Estate industry (ID: 7)
	m.keywords[7] = []*IndustryKeyword{
		{ID: 61, IndustryID: 7, Keyword: "real estate", Weight: 1.0, IsActive: true},
		{ID: 62, IndustryID: 7, Keyword: "property", Weight: 0.9, IsActive: true},
		{ID: 63, IndustryID: 7, Keyword: "residential", Weight: 0.8, IsActive: true},
		{ID: 64, IndustryID: 7, Keyword: "commercial", Weight: 0.8, IsActive: true},
		{ID: 65, IndustryID: 7, Keyword: "sales", Weight: 0.7, IsActive: true},
		{ID: 66, IndustryID: 7, Keyword: "leasing", Weight: 0.7, IsActive: true},
		{ID: 67, IndustryID: 7, Keyword: "property management", Weight: 0.8, IsActive: true},
	}

	m.classificationCodes[7] = []*ClassificationCode{
		{ID: 37, IndustryID: 7, CodeType: "MCC", Code: "6513", Description: "Real Estate Agents and Managers", IsActive: true},
		{ID: 38, IndustryID: 7, CodeType: "MCC", Code: "6514", Description: "Real Estate Services", IsActive: true},
		{ID: 39, IndustryID: 7, CodeType: "SIC", Code: "6531", Description: "Real Estate Agents and Managers", IsActive: true},
		{ID: 40, IndustryID: 7, CodeType: "SIC", Code: "6512", Description: "Operators of Nonresidential Buildings", IsActive: true},
		{ID: 41, IndustryID: 7, CodeType: "NAICS", Code: "531210", Description: "Offices of Real Estate Agents and Brokers", IsActive: true},
		{ID: 42, IndustryID: 7, CodeType: "NAICS", Code: "531312", Description: "Nonresidential Property Managers", IsActive: true},
	}

	// Education industry (ID: 8)
	m.keywords[8] = []*IndustryKeyword{
		{ID: 68, IndustryID: 8, Keyword: "education", Weight: 1.0, IsActive: true},
		{ID: 69, IndustryID: 8, Keyword: "e-learning", Weight: 0.9, IsActive: true},
		{ID: 70, IndustryID: 8, Keyword: "online learning", Weight: 0.8, IsActive: true},
		{ID: 71, IndustryID: 8, Keyword: "learning", Weight: 0.7, IsActive: true},
		{ID: 72, IndustryID: 8, Keyword: "educational software", Weight: 0.8, IsActive: true},
		{ID: 73, IndustryID: 8, Keyword: "digital learning", Weight: 0.7, IsActive: true},
	}

	m.classificationCodes[8] = []*ClassificationCode{
		{ID: 43, IndustryID: 8, CodeType: "MCC", Code: "7372", Description: "Computer Programming Services", IsActive: true},
		{ID: 44, IndustryID: 8, CodeType: "MCC", Code: "5999", Description: "Miscellaneous and Specialty Retail Stores", IsActive: true},
		{ID: 45, IndustryID: 8, CodeType: "SIC", Code: "7372", Description: "Prepackaged Software", IsActive: true},
		{ID: 46, IndustryID: 8, CodeType: "SIC", Code: "8299", Description: "Schools and Educational Services", IsActive: true},
		{ID: 47, IndustryID: 8, CodeType: "NAICS", Code: "611710", Description: "Educational Support Services", IsActive: true},
		{ID: 48, IndustryID: 8, CodeType: "NAICS", Code: "518210", Description: "Data Processing, Hosting, and Related Services", IsActive: true},
	}

	// Energy industry (ID: 9)
	m.keywords[9] = []*IndustryKeyword{
		{ID: 74, IndustryID: 9, Keyword: "renewable energy", Weight: 1.0, IsActive: true},
		{ID: 75, IndustryID: 9, Keyword: "solar", Weight: 0.9, IsActive: true},
		{ID: 76, IndustryID: 9, Keyword: "wind power", Weight: 0.9, IsActive: true},
		{ID: 77, IndustryID: 9, Keyword: "clean technology", Weight: 0.8, IsActive: true},
		{ID: 78, IndustryID: 9, Keyword: "sustainable energy", Weight: 0.8, IsActive: true},
		{ID: 79, IndustryID: 9, Keyword: "energy", Weight: 0.7, IsActive: true},
	}

	m.classificationCodes[9] = []*ClassificationCode{
		{ID: 49, IndustryID: 9, CodeType: "MCC", Code: "4900", Description: "Utilities", IsActive: true},
		{ID: 50, IndustryID: 9, CodeType: "MCC", Code: "5999", Description: "Miscellaneous and Specialty Retail Stores", IsActive: true},
		{ID: 51, IndustryID: 9, CodeType: "SIC", Code: "4911", Description: "Electric Services", IsActive: true},
		{ID: 52, IndustryID: 9, CodeType: "SIC", Code: "4953", Description: "Refuse Systems", IsActive: true},
		{ID: 53, IndustryID: 9, CodeType: "NAICS", Code: "221111", Description: "Hydroelectric Power Generation", IsActive: true},
		{ID: 54, IndustryID: 9, CodeType: "NAICS", Code: "221115", Description: "Wind Electric Power Generation", IsActive: true},
	}

	// General Business industry (ID: 10)
	m.keywords[10] = []*IndustryKeyword{
		{ID: 80, IndustryID: 10, Keyword: "business", Weight: 0.5, IsActive: true},
		{ID: 81, IndustryID: 10, Keyword: "services", Weight: 0.4, IsActive: true},
		{ID: 82, IndustryID: 10, Keyword: "consulting", Weight: 0.3, IsActive: true},
		{ID: 83, IndustryID: 10, Keyword: "solutions", Weight: 0.3, IsActive: true},
	}

	m.classificationCodes[10] = []*ClassificationCode{
		{ID: 55, IndustryID: 10, CodeType: "MCC", Code: "8999", Description: "Professional Services", IsActive: true},
		{ID: 56, IndustryID: 10, CodeType: "SIC", Code: "7389", Description: "Business Services", IsActive: true},
		{ID: 57, IndustryID: 10, CodeType: "NAICS", Code: "541990", Description: "All Other Professional, Scientific, and Technical Services", IsActive: true},
	}

	// Initialize codes by type
	m.initializeCodesByType()
}

// initializeCodesByType initializes the codes by type map
func (m *MockKeywordRepository) initializeCodesByType() {
	// Collect all codes by type
	mccCodes := []*ClassificationCode{}
	sicCodes := []*ClassificationCode{}
	naicsCodes := []*ClassificationCode{}

	for _, codes := range m.classificationCodes {
		for _, code := range codes {
			switch code.CodeType {
			case "MCC":
				mccCodes = append(mccCodes, code)
			case "SIC":
				sicCodes = append(sicCodes, code)
			case "NAICS":
				naicsCodes = append(naicsCodes, code)
			}
		}
	}

	m.codesByType["MCC"] = mccCodes
	m.codesByType["SIC"] = sicCodes
	m.codesByType["NAICS"] = naicsCodes
}

// AddKeyword adds a keyword to the mock repository
func (m *MockKeywordRepository) AddKeyword(industryID int, keyword *IndustryKeyword) {
	if m.keywords[industryID] == nil {
		m.keywords[industryID] = []*IndustryKeyword{}
	}
	m.keywords[industryID] = append(m.keywords[industryID], keyword)
}

// AddClassificationCode adds a classification code to the mock repository
func (m *MockKeywordRepository) AddClassificationCode(industryID int, code *ClassificationCode) {
	if m.classificationCodes[industryID] == nil {
		m.classificationCodes[industryID] = []*ClassificationCode{}
	}
	m.classificationCodes[industryID] = append(m.classificationCodes[industryID], code)

	// Also add to codes by type
	if m.codesByType[code.CodeType] == nil {
		m.codesByType[code.CodeType] = []*ClassificationCode{}
	}
	m.codesByType[code.CodeType] = append(m.codesByType[code.CodeType], code)
}

// GetStatistics returns statistics about the mock repository
func (m *MockKeywordRepository) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count keywords by industry
	keywordCounts := make(map[int]int)
	for industryID, keywords := range m.keywords {
		keywordCounts[industryID] = len(keywords)
	}
	stats["keywords_by_industry"] = keywordCounts

	// Count codes by industry
	codeCounts := make(map[int]int)
	for industryID, codes := range m.classificationCodes {
		codeCounts[industryID] = len(codes)
	}
	stats["codes_by_industry"] = codeCounts

	// Count codes by type
	codeTypeCounts := make(map[string]int)
	for codeType, codes := range m.codesByType {
		codeTypeCounts[codeType] = len(codes)
	}
	stats["codes_by_type"] = codeTypeCounts

	// Total counts
	totalKeywords := 0
	for _, count := range keywordCounts {
		totalKeywords += count
	}
	stats["total_keywords"] = totalKeywords

	totalCodes := 0
	for _, count := range codeCounts {
		totalCodes += count
	}
	stats["total_codes"] = totalCodes

	stats["total_industries"] = len(m.keywords)

	return stats
}
