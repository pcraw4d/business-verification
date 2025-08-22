package classification

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIndustryCodes(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "industry_codes_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test NAICS file
	naicsFile := filepath.Join(tempDir, "NAICS-2022-Codes_industries.csv")
	naicsData := `2022 NAICS Code,2022 NAICS Title,
541511,Custom Computer Programming Services,
541512,Computer Systems Design Services,
541519,Other Computer Related Services,
522110,Commercial Banking,
621111,Offices of Physicians (except Mental Health Specialists),
`
	err = os.WriteFile(naicsFile, []byte(naicsData), 0644)
	if err != nil {
		t.Fatalf("Failed to write NAICS test file: %v", err)
	}

	// Create test MCC file
	mccFile := filepath.Join(tempDir, "mcc_codes.csv")
	mccData := `MCC,Description
5045,"Computers, Computer Peripheral Equipment, Software"
5411,"Grocery Stores, Supermarkets"
5812,"Eating Places and Restaurants"
`
	err = os.WriteFile(mccFile, []byte(mccData), 0644)
	if err != nil {
		t.Fatalf("Failed to write MCC test file: %v", err)
	}

	// Create test SIC file
	sicFile := filepath.Join(tempDir, "sic-codes.csv")
	sicData := `Division,Major Group,Industry Group,SIC,Description
C,35,357,3571,Electronic Computers
C,35,357,3572,Computer Storage Devices
C,35,357,3575,Computer Terminals
`
	err = os.WriteFile(sicFile, []byte(sicData), 0644)
	if err != nil {
		t.Fatalf("Failed to write SIC test file: %v", err)
	}

	// Load industry codes
	data, err := LoadIndustryCodes(tempDir)
	if err != nil {
		t.Fatalf("Failed to load industry codes: %v", err)
	}

	// Test NAICS data
	if data.GetNAICSCount() == 0 {
		t.Error("Expected NAICS codes to be loaded")
	}

	naicsName := data.GetNAICSName("541511")
	if naicsName != "Custom Computer Programming Services" {
		t.Errorf("Expected NAICS name 'Custom Computer Programming Services', got %s", naicsName)
	}

	// Test MCC data
	if data.GetMCCCount() == 0 {
		t.Error("Expected MCC codes to be loaded")
	}

	mccDesc := data.GetMCCDescription("5045")
	if mccDesc != "Computers, Computer Peripheral Equipment, Software" {
		t.Errorf("Expected MCC description 'Computers, Computer Peripheral Equipment, Software', got %s", mccDesc)
	}

	// Test SIC data
	if data.GetSICCount() == 0 {
		t.Error("Expected SIC codes to be loaded")
	}

	sicDesc := data.GetSICDescription("C-35-357-3571")
	if sicDesc != "Electronic Computers" {
		t.Errorf("Expected SIC description 'Electronic Computers', got %s", sicDesc)
	}
}

func TestSearchNAICSByKeyword(t *testing.T) {
	data := &IndustryCodeData{
		NAICS: map[string]string{
			"541511": "Custom Computer Programming Services",
			"541512": "Computer Systems Design Services",
			"522110": "Commercial Banking",
			"621111": "Offices of Physicians (except Mental Health Specialists)",
		},
	}

	// Test search for "computer"
	results := data.SearchNAICSByKeyword("computer")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'computer', got %d", len(results))
	}

	// Test search for "banking"
	results = data.SearchNAICSByKeyword("banking")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'banking', got %d", len(results))
	}

	// Test search for non-existent keyword
	results = data.SearchNAICSByKeyword("nonexistent")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestSearchMCCByKeyword(t *testing.T) {
	data := &IndustryCodeData{
		MCC: map[string]string{
			"5045": "Computers, Computer Peripheral Equipment, Software",
			"5411": "Grocery Stores, Supermarkets",
			"5812": "Eating Places and Restaurants",
		},
	}

	// Test search for "computer"
	results := data.SearchMCCByKeyword("computer")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'computer', got %d", len(results))
	}

	// Test search for "store"
	results = data.SearchMCCByKeyword("store")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'store', got %d", len(results))
	}

	// Test search for "grocery"
	results = data.SearchMCCByKeyword("grocery")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'grocery', got %d", len(results))
	}
}

func TestSearchSICByKeyword(t *testing.T) {
	data := &IndustryCodeData{
		SIC: map[string]string{
			"C-35-357-3571": "Electronic Computers",
			"C-35-357-3572": "Computer Storage Devices",
			"C-35-357-3575": "Computer Terminals",
		},
	}

	// Test search for "computer"
	results := data.SearchSICByKeyword("computer")
	if len(results) != 3 {
		t.Errorf("Expected 3 results for 'computer', got %d", len(results))
	}

	// Test search for "electronic"
	results = data.SearchSICByKeyword("electronic")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'electronic', got %d", len(results))
	}
}

func TestGetIndustryNameMethods(t *testing.T) {
	data := &IndustryCodeData{
		NAICS: map[string]string{
			"541511": "Custom Computer Programming Services",
		},
		MCC: map[string]string{
			"5045": "Computers, Computer Peripheral Equipment, Software",
		},
		SIC: map[string]string{
			"C-35-357-3571": "Electronic Computers",
		},
	}

	// Test NAICS name
	name := data.GetNAICSName("541511")
	if name != "Custom Computer Programming Services" {
		t.Errorf("Expected NAICS name 'Custom Computer Programming Services', got %s", name)
	}

	// Test unknown NAICS code
	name = data.GetNAICSName("999999")
	if name != "Unknown NAICS Industry" {
		t.Errorf("Expected 'Unknown NAICS Industry', got %s", name)
	}

	// Test MCC description
	desc := data.GetMCCDescription("5045")
	if desc != "Computers, Computer Peripheral Equipment, Software" {
		t.Errorf("Expected MCC description 'Computers, Computer Peripheral Equipment, Software', got %s", desc)
	}

	// Test unknown MCC code
	desc = data.GetMCCDescription("9999")
	if desc != "Unknown MCC Industry" {
		t.Errorf("Expected 'Unknown MCC Industry', got %s", desc)
	}

	// Test SIC description
	desc = data.GetSICDescription("C-35-357-3571")
	if desc != "Electronic Computers" {
		t.Errorf("Expected SIC description 'Electronic Computers', got %s", desc)
	}

	// Test unknown SIC code
	desc = data.GetSICDescription("unknown")
	if desc != "Unknown SIC Industry" {
		t.Errorf("Expected 'Unknown SIC Industry', got %s", desc)
	}
}

func TestLoadIndustryCodesWithMissingFiles(t *testing.T) {
	// Create a temporary directory without any CSV files
	tempDir, err := os.MkdirTemp("", "industry_codes_test_empty")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try to load industry codes from empty directory
	_, err = LoadIndustryCodes(tempDir)
	if err == nil {
		t.Error("Expected error when loading from directory without CSV files")
	}
}
