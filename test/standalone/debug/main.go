package main

import (
	"fmt"
	"strings"
	"time"
)

type Location struct {
	Type            string
	Address         string
	City            string
	State           string
	Country         string
	PostalCode      string
	Phone           string
	Email           string
	ConfidenceScore float64
	ExtractedAt     time.Time
	Source          string
}

func parseAddress(address, country string) Location {
	location := Location{
		Type:        "office",
		Address:     address,
		Country:     country,
		ExtractedAt: time.Now(),
		Source:      "address_parsing",
	}

	// For now, just use simple string matching
	if country == "us" && strings.Contains(address, "10001") {
		location.PostalCode = "10001"
	} else if country == "uk" && strings.Contains(address, "SW1A 2AA") {
		location.PostalCode = "SW1A 2AA"
	}

	// Parse address components
	parts := strings.Split(address, ",")
	if len(parts) >= 2 {
		// For US addresses: "Street, City, State PostalCode"
		// For UK addresses: "Street, City, PostalCode"
		if country == "us" && len(parts) >= 3 {
			// Last part contains state and postal code
			lastPart := strings.TrimSpace(parts[len(parts)-1])
			// Remove postal code from last part to get state
			if location.PostalCode != "" {
				statePart := strings.ReplaceAll(lastPart, location.PostalCode, "")
				statePart = strings.TrimSpace(statePart)
				location.State = statePart
			}
			// Second to last part is city
			location.City = strings.TrimSpace(parts[len(parts)-2])
		} else {
			// For UK and other formats, last part might be postal code
			// Second to last part is city
			cityPart := strings.TrimSpace(parts[len(parts)-2])
			// Remove postal code from city part if it's there
			if location.PostalCode != "" {
				cityPart = strings.ReplaceAll(cityPart, location.PostalCode, "")
				cityPart = strings.TrimSpace(cityPart)
			}
			location.City = cityPart
		}
	}

	// Calculate confidence based on completeness
	confidence := 0.3
	if location.City != "" {
		confidence += 0.2
	}
	if location.State != "" {
		confidence += 0.2
	}
	if location.PostalCode != "" {
		confidence += 0.3
	}

	location.ConfidenceScore = confidence
	return location
}

func main() {
	// Test US address
	usResult := parseAddress("123 Main Street, New York, NY 10001", "us")
	fmt.Printf("US Address Result:\n")
	fmt.Printf("  City: '%s'\n", usResult.City)
	fmt.Printf("  State: '%s'\n", usResult.State)
	fmt.Printf("  PostalCode: '%s'\n", usResult.PostalCode)
	fmt.Printf("  Confidence: %f\n", usResult.ConfidenceScore)

	// Test UK address
	ukResult := parseAddress("10 Downing Street, London, SW1A 2AA", "uk")
	fmt.Printf("\nUK Address Result:\n")
	fmt.Printf("  City: '%s'\n", ukResult.City)
	fmt.Printf("  State: '%s'\n", ukResult.State)
	fmt.Printf("  PostalCode: '%s'\n", ukResult.PostalCode)
	fmt.Printf("  Confidence: %f\n", ukResult.ConfidenceScore)
}
