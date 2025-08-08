package classification

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IndustryCodeData holds all the industry code mappings
type IndustryCodeData struct {
	NAICS map[string]string // NAICS code -> title
	MCC   map[string]string // MCC code -> description
	SIC   map[string]string // SIC code -> description
}

// LoadIndustryCodes loads all industry code data from CSV files
func LoadIndustryCodes(dataPath string) (*IndustryCodeData, error) {
	data := &IndustryCodeData{
		NAICS: make(map[string]string),
		MCC:   make(map[string]string),
		SIC:   make(map[string]string),
	}

	// Load NAICS codes
	if err := loadNAICSCodes(dataPath, data); err != nil {
		return nil, fmt.Errorf("failed to load NAICS codes: %w", err)
	}

	// Load MCC codes
	if err := loadMCCCodes(dataPath, data); err != nil {
		return nil, fmt.Errorf("failed to load MCC codes: %w", err)
	}

	// Load SIC codes
	if err := loadSICCodes(dataPath, data); err != nil {
		return nil, fmt.Errorf("failed to load SIC codes: %w", err)
	}

	return data, nil
}

// loadNAICSCodes loads NAICS codes from CSV file
func loadNAICSCodes(dataPath string, data *IndustryCodeData) error {
	filePath := filepath.Join(dataPath, "NAICS-2022-Codes_industries.csv")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open NAICS file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read NAICS CSV: %w", err)
	}

	// Skip header and empty rows
	for i, record := range records {
		if i == 0 || len(record) < 2 || record[0] == "" {
			continue
		}

		code := strings.TrimSpace(record[0])
		title := strings.TrimSpace(record[1])

		if code != "" && title != "" {
			data.NAICS[code] = title
		}
	}

	return nil
}

// loadMCCCodes loads MCC codes from CSV file
func loadMCCCodes(dataPath string, data *IndustryCodeData) error {
	filePath := filepath.Join(dataPath, "mcc_codes.csv")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open MCC file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read MCC CSV: %w", err)
	}

	// Skip header
	for i, record := range records {
		if i == 0 || len(record) < 2 {
			continue
		}

		code := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])

		if code != "" && description != "" {
			data.MCC[code] = description
		}
	}

	return nil
}

// loadSICCodes loads SIC codes from CSV file
func loadSICCodes(dataPath string, data *IndustryCodeData) error {
	filePath := filepath.Join(dataPath, "sic-codes.csv")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open SIC file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read SIC CSV: %w", err)
	}

	// Skip header
	for i, record := range records {
		if i == 0 || len(record) < 5 {
			continue
		}

		division := strings.TrimSpace(record[0])
		majorGroup := strings.TrimSpace(record[1])
		industryGroup := strings.TrimSpace(record[2])
		sicCode := strings.TrimSpace(record[3])
		description := strings.TrimSpace(record[4])

		if sicCode != "" && description != "" {
			// Create a composite key for better organization
			key := fmt.Sprintf("%s-%s-%s-%s", division, majorGroup, industryGroup, sicCode)
			data.SIC[key] = description
		}
	}

	return nil
}

// GetNAICSName returns the NAICS title for a given code
func (d *IndustryCodeData) GetNAICSName(code string) string {
	if name, exists := d.NAICS[code]; exists {
		return name
	}
	return "Unknown NAICS Industry"
}

// GetMCCDescription returns the MCC description for a given code
func (d *IndustryCodeData) GetMCCDescription(code string) string {
	if description, exists := d.MCC[code]; exists {
		return description
	}
	return "Unknown MCC Industry"
}

// GetSICDescription returns the SIC description for a given code
func (d *IndustryCodeData) GetSICDescription(code string) string {
	if description, exists := d.SIC[code]; exists {
		return description
	}
	return "Unknown SIC Industry"
}

// SearchNAICSByKeyword searches NAICS codes by keyword in title
func (d *IndustryCodeData) SearchNAICSByKeyword(keyword string) []string {
	var results []string
	keyword = strings.ToLower(keyword)

	for code, title := range d.NAICS {
		if strings.Contains(strings.ToLower(title), keyword) {
			results = append(results, code)
		}
	}

	return results
}

// SearchNAICSByFuzzy returns NAICS codes whose titles are similar to the query above the threshold
// threshold in [0,1]; typical values: 0.72-0.85
func (d *IndustryCodeData) SearchNAICSByFuzzy(query string, threshold float64) []string {
	var results []string
	if query == "" {
		return results
	}
	for code, title := range d.NAICS {
		if tokenMaxSimilarity(query, title) >= threshold {
			results = append(results, code)
		}
	}
	return results
}

// SearchMCCByKeyword searches MCC codes by keyword in description
func (d *IndustryCodeData) SearchMCCByKeyword(keyword string) []string {
	var results []string
	keyword = strings.ToLower(keyword)

	for code, description := range d.MCC {
		if strings.Contains(strings.ToLower(description), keyword) {
			results = append(results, code)
		}
	}

	return results
}

// SearchMCCByFuzzy returns MCC codes whose descriptions are similar to the query above the threshold
func (d *IndustryCodeData) SearchMCCByFuzzy(query string, threshold float64) []string {
	var results []string
	if query == "" {
		return results
	}
	for code, desc := range d.MCC {
		if tokenMaxSimilarity(query, desc) >= threshold {
			results = append(results, code)
		}
	}
	return results
}

// SearchSICByKeyword searches SIC codes by keyword in description
func (d *IndustryCodeData) SearchSICByKeyword(keyword string) []string {
	var results []string
	keyword = strings.ToLower(keyword)

	for code, description := range d.SIC {
		if strings.Contains(strings.ToLower(description), keyword) {
			results = append(results, code)
		}
	}

	return results
}

// SearchSICByFuzzy returns SIC codes whose descriptions are similar to the query above the threshold
func (d *IndustryCodeData) SearchSICByFuzzy(query string, threshold float64) []string {
	var results []string
	if query == "" {
		return results
	}
	for code, desc := range d.SIC {
		if tokenMaxSimilarity(query, desc) >= threshold {
			results = append(results, code)
		}
	}
	return results
}

// GetNAICSCount returns the total number of NAICS codes loaded
func (d *IndustryCodeData) GetNAICSCount() int {
	return len(d.NAICS)
}

// GetMCCCount returns the total number of MCC codes loaded
func (d *IndustryCodeData) GetMCCCount() int {
	return len(d.MCC)
}

// GetSICCount returns the total number of SIC codes loaded
func (d *IndustryCodeData) GetSICCount() int {
	return len(d.SIC)
}
