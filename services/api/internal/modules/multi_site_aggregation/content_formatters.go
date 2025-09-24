package multi_site_aggregation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BasicContentFormatter implements ContentFormatter interface with basic formatting capabilities
type BasicContentFormatter struct {
	region string
	rules  *LocalizationRules
}

// FormatDate formats a date string according to the target format
func (f *BasicContentFormatter) FormatDate(date string, targetFormat string) (string, error) {
	if date == "" {
		return date, nil
	}

	// Try to parse common date formats
	var parsedDate time.Time
	var err error

	// Common date formats to try
	dateFormats := []string{
		"2006-01-02",          // ISO format
		"01/02/2006",          // US format
		"02/01/2006",          // EU format
		"2006/01/02",          // Alternative ISO
		"02.01.2006",          // German format
		"2006-01-02 15:04:05", // ISO with time
		"January 2, 2006",     // Long format
		"Jan 2, 2006",         // Medium format
	}

	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, date)
		if err == nil {
			break
		}
	}

	if err != nil {
		return date, fmt.Errorf("unable to parse date: %s", date)
	}

	// Format according to target format
	switch targetFormat {
	case "MM/dd/yyyy":
		return parsedDate.Format("01/02/2006"), nil
	case "dd/MM/yyyy":
		return parsedDate.Format("02/01/2006"), nil
	case "dd.MM.yyyy":
		return parsedDate.Format("02.01.2006"), nil
	case "yyyy-MM-dd":
		return parsedDate.Format("2006-01-02"), nil
	default:
		// If target format is not recognized, try to use it as Go time format
		return parsedDate.Format(targetFormat), nil
	}
}

// FormatTime formats a time string according to the target format
func (f *BasicContentFormatter) FormatTime(timeStr string, targetFormat string) (string, error) {
	if timeStr == "" {
		return timeStr, nil
	}

	// Try to parse common time formats
	var parsedTime time.Time
	var err error

	timeFormats := []string{
		"15:04",      // 24-hour format
		"3:04 PM",    // 12-hour format
		"3:04 pm",    // 12-hour format lowercase
		"15:04:05",   // 24-hour with seconds
		"3:04:05 PM", // 12-hour with seconds
	}

	for _, format := range timeFormats {
		parsedTime, err = time.Parse(format, timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return timeStr, fmt.Errorf("unable to parse time: %s", timeStr)
	}

	// Format according to target format
	switch targetFormat {
	case "HH:mm":
		return parsedTime.Format("15:04"), nil
	case "h:mm a":
		return parsedTime.Format("3:04 pm"), nil
	case "h:mm A":
		return parsedTime.Format("3:04 PM"), nil
	case "HH:mm:ss":
		return parsedTime.Format("15:04:05"), nil
	default:
		return parsedTime.Format(targetFormat), nil
	}
}

// FormatNumber formats a number string according to the target format
func (f *BasicContentFormatter) FormatNumber(number string, targetFormat string) (string, error) {
	if number == "" {
		return number, nil
	}

	// Clean the number string - remove existing separators
	cleaned := regexp.MustCompile(`[,.\s]`).ReplaceAllString(number, "")

	// Check if it's actually a number
	if !regexp.MustCompile(`^\d+$`).MatchString(cleaned) {
		return number, fmt.Errorf("not a valid number: %s", number)
	}

	// Parse as integer (for simplicity)
	value, err := strconv.Atoi(cleaned)
	if err != nil {
		return number, fmt.Errorf("unable to parse number: %s", number)
	}

	// Format according to target format
	switch targetFormat {
	case "1,234.56":
		return f.formatNumberUS(value), nil
	case "1.234,56":
		return f.formatNumberEU(value), nil
	case "1 234,56":
		return f.formatNumberFR(value), nil
	default:
		return f.formatNumberUS(value), nil // Default to US format
	}
}

// FormatCurrency formats a currency string according to the target format
func (f *BasicContentFormatter) FormatCurrency(amount string, currency string, targetFormat string) (string, error) {
	if amount == "" {
		return amount, nil
	}

	// Extract numeric value from currency string
	numericPattern := regexp.MustCompile(`[\d.,]+`)
	numericPart := numericPattern.FindString(amount)

	if numericPart == "" {
		return amount, fmt.Errorf("no numeric value found in currency: %s", amount)
	}

	// Clean and parse the numeric value
	cleaned := strings.ReplaceAll(strings.ReplaceAll(numericPart, ",", ""), " ", "")
	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return amount, fmt.Errorf("unable to parse currency amount: %s", amount)
	}

	// Detect or use provided currency
	if currency == "" {
		currency = f.detectCurrency(amount)
	}

	// Format according to target format
	return f.formatCurrencyValue(value, currency, targetFormat), nil
}

// FormatAddress formats an address according to the target format
func (f *BasicContentFormatter) FormatAddress(address *LocalizedAddress, targetFormat string) (string, error) {
	if address == nil {
		return "", fmt.Errorf("address is nil")
	}

	// Parse target format and build formatted address
	formatted := targetFormat

	// Replace placeholders with actual values
	replacements := map[string]string{
		"{street}":      address.Street,
		"{city}":        address.City,
		"{state}":       address.State,
		"{postal_code}": address.PostalCode,
		"{country}":     address.Country,
	}

	for placeholder, value := range replacements {
		if value != "" {
			formatted = strings.Replace(formatted, placeholder, value, -1)
		} else {
			// Remove placeholder and any surrounding punctuation
			formatted = strings.Replace(formatted, placeholder+", ", "", -1)
			formatted = strings.Replace(formatted, ", "+placeholder, "", -1)
			formatted = strings.Replace(formatted, placeholder, "", -1)
		}
	}

	// Clean up any double commas or extra spaces
	formatted = regexp.MustCompile(`,\s*,`).ReplaceAllString(formatted, ",")
	formatted = regexp.MustCompile(`\s+`).ReplaceAllString(formatted, " ")
	formatted = strings.TrimSpace(formatted)
	formatted = strings.Trim(formatted, ",")

	return formatted, nil
}

// FormatPhone formats a phone number according to the target format
func (f *BasicContentFormatter) FormatPhone(phone string, targetFormat string) (string, error) {
	if phone == "" {
		return phone, nil
	}

	// Extract digits only
	digits := regexp.MustCompile(`\d`).FindAllString(phone, -1)
	if len(digits) == 0 {
		return phone, fmt.Errorf("no digits found in phone number: %s", phone)
	}

	digitString := strings.Join(digits, "")

	// Format according to target format
	switch targetFormat {
	case "(123) 456-7890":
		return f.formatPhoneUS(digitString), nil
	case "+44 20 1234 5678":
		return f.formatPhoneGB(digitString), nil
	case "+49 30 12345678":
		return f.formatPhoneDE(digitString), nil
	case "+33 1 23 45 67 89":
		return f.formatPhoneFR(digitString), nil
	default:
		return f.formatPhoneUS(digitString), nil // Default to US format
	}
}

// Helper methods for number formatting

func (f *BasicContentFormatter) formatNumberUS(value int) string {
	// Format with comma thousand separators
	str := strconv.Itoa(value)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func (f *BasicContentFormatter) formatNumberEU(value int) string {
	// Format with dot thousand separators
	str := strconv.Itoa(value)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func (f *BasicContentFormatter) formatNumberFR(value int) string {
	// Format with space thousand separators
	str := strconv.Itoa(value)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(" ")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func (f *BasicContentFormatter) detectCurrency(amount string) string {
	// Simple currency detection based on symbols
	if strings.Contains(amount, "$") {
		if strings.Contains(amount, "C$") {
			return "CAD"
		}
		if strings.Contains(amount, "A$") {
			return "AUD"
		}
		return "USD"
	}
	if strings.Contains(amount, "€") {
		return "EUR"
	}
	if strings.Contains(amount, "£") {
		return "GBP"
	}
	if strings.Contains(amount, "¥") {
		return "JPY"
	}

	// Default based on region
	switch f.region {
	case "US":
		return "USD"
	case "CA":
		return "CAD"
	case "GB":
		return "GBP"
	case "EU", "DE", "FR", "IT", "ES":
		return "EUR"
	case "AU":
		return "AUD"
	case "JP":
		return "JPY"
	default:
		return "USD"
	}
}

func (f *BasicContentFormatter) formatCurrencyValue(value float64, currency string, targetFormat string) string {
	// Get currency symbol
	symbol := f.getCurrencySymbol(currency)

	// Format number part
	var numberPart string
	switch targetFormat {
	case "$1,234.56":
		numberPart = fmt.Sprintf("%.2f", value)
		numberPart = f.addThousandsSeparators(numberPart, ",", ".")
		return symbol + numberPart
	case "1.234,56 €":
		numberPart = fmt.Sprintf("%.2f", value)
		numberPart = f.addThousandsSeparators(numberPart, ".", ",")
		return numberPart + " " + symbol
	case "£1,234.56":
		numberPart = fmt.Sprintf("%.2f", value)
		numberPart = f.addThousandsSeparators(numberPart, ",", ".")
		return symbol + numberPart
	default:
		numberPart = fmt.Sprintf("%.2f", value)
		numberPart = f.addThousandsSeparators(numberPart, ",", ".")
		return symbol + numberPart
	}
}

func (f *BasicContentFormatter) getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$",
		"CAD": "C$",
		"AUD": "A$",
		"EUR": "€",
		"GBP": "£",
		"JPY": "¥",
	}

	if symbol, exists := symbols[currency]; exists {
		return symbol
	}
	return currency
}

func (f *BasicContentFormatter) addThousandsSeparators(numberStr, thousandsSep, decimalSep string) string {
	parts := strings.Split(numberStr, ".")
	integerPart := parts[0]
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}

	// Add thousands separators to integer part
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(thousandsSep)
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}

	// Combine with decimal part
	if decimalPart != "" {
		return integerPart + decimalSep + decimalPart
	}
	return integerPart
}

// Helper methods for phone formatting

func (f *BasicContentFormatter) formatPhoneUS(digits string) string {
	// Remove country code if present
	if len(digits) == 11 && digits[0] == '1' {
		digits = digits[1:]
	}

	if len(digits) != 10 {
		return digits // Return as-is if not 10 digits
	}

	return fmt.Sprintf("(%s) %s-%s", digits[0:3], digits[3:6], digits[6:10])
}

func (f *BasicContentFormatter) formatPhoneGB(digits string) string {
	// Add +44 country code and format
	if len(digits) >= 10 {
		// Remove leading 0 if present and add +44
		if digits[0] == '0' {
			digits = digits[1:]
		}
		return fmt.Sprintf("+44 %s %s %s", digits[0:2], digits[2:6], digits[6:])
	}
	return digits
}

func (f *BasicContentFormatter) formatPhoneDE(digits string) string {
	// Add +49 country code and format
	if len(digits) >= 10 {
		// Remove leading 0 if present and add +49
		if digits[0] == '0' {
			digits = digits[1:]
		}
		return fmt.Sprintf("+49 %s %s", digits[0:2], digits[2:])
	}
	return digits
}

func (f *BasicContentFormatter) formatPhoneFR(digits string) string {
	// Add +33 country code and format
	if len(digits) >= 10 {
		// Remove leading 0 if present and add +33
		if digits[0] == '0' {
			digits = digits[1:]
		}
		if len(digits) >= 9 {
			return fmt.Sprintf("+33 %s %s %s %s %s",
				digits[0:1],
				digits[1:3],
				digits[3:5],
				digits[5:7],
				digits[7:9])
		}
	}
	return digits
}

// BasicContentValidator implements ContentValidator interface with basic validation capabilities
type BasicContentValidator struct {
	region string
	rules  *LocalizationRules
}

// ValidateField validates a field value against validation rules
func (v *BasicContentValidator) ValidateField(field string, value interface{}, rules []ValidationRule) []LocalizationError {
	var errors []LocalizationError

	valueStr := fmt.Sprintf("%v", value)

	for _, rule := range rules {
		switch rule.Type {
		case "required":
			if valueStr == "" || valueStr == "<nil>" {
				errors = append(errors, LocalizationError{
					Field:      field,
					ErrorType:  "required",
					Message:    fmt.Sprintf("Field '%s' is required", field),
					Severity:   "high",
					Suggestion: fmt.Sprintf("Provide a value for field '%s'", field),
				})
			}
		case "format":
			if rule.Pattern != "" {
				matched, err := regexp.MatchString(rule.Pattern, valueStr)
				if err != nil || !matched {
					errors = append(errors, LocalizationError{
						Field:      field,
						ErrorType:  "format",
						Message:    fmt.Sprintf("Field '%s' does not match required format", field),
						Severity:   "medium",
						Suggestion: fmt.Sprintf("Format field '%s' according to pattern: %s", field, rule.Pattern),
					})
				}
			}
		case "length":
			if rule.MinLength > 0 && len(valueStr) < rule.MinLength {
				errors = append(errors, LocalizationError{
					Field:      field,
					ErrorType:  "length",
					Message:    fmt.Sprintf("Field '%s' is too short (minimum %d characters)", field, rule.MinLength),
					Severity:   "medium",
					Suggestion: fmt.Sprintf("Provide at least %d characters for field '%s'", rule.MinLength, field),
				})
			}
			if rule.MaxLength > 0 && len(valueStr) > rule.MaxLength {
				errors = append(errors, LocalizationError{
					Field:      field,
					ErrorType:  "length",
					Message:    fmt.Sprintf("Field '%s' is too long (maximum %d characters)", field, rule.MaxLength),
					Severity:   "medium",
					Suggestion: fmt.Sprintf("Limit field '%s' to %d characters", field, rule.MaxLength),
				})
			}
		}
	}

	return errors
}

// ValidateContact validates localized contact information
func (v *BasicContentValidator) ValidateContact(contact *LocalizedContactInfo) []LocalizationError {
	var errors []LocalizationError

	// Validate phone numbers
	for i, phone := range contact.PhoneNumbers {
		if phone.Number == "" {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("phone_numbers[%d].number", i),
				ErrorType:  "required",
				Message:    "Phone number is required",
				Severity:   "high",
				Suggestion: "Provide a valid phone number",
			})
		} else if !v.isValidPhoneFormat(phone.Number, v.region) {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("phone_numbers[%d].number", i),
				ErrorType:  "format",
				Message:    "Phone number format is invalid for region",
				Severity:   "medium",
				Suggestion: fmt.Sprintf("Format phone number according to %s standards", v.region),
			})
		}
	}

	// Validate email addresses
	for i, email := range contact.EmailAddresses {
		if email.Email == "" {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("email_addresses[%d].email", i),
				ErrorType:  "required",
				Message:    "Email address is required",
				Severity:   "high",
				Suggestion: "Provide a valid email address",
			})
		} else if !v.isValidEmailFormat(email.Email) {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("email_addresses[%d].email", i),
				ErrorType:  "format",
				Message:    "Email address format is invalid",
				Severity:   "medium",
				Suggestion: "Provide a valid email address format",
			})
		}
	}

	// Validate addresses
	for i, address := range contact.Addresses {
		if address.FullAddress == "" && address.FormattedAddress == "" {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("addresses[%d]", i),
				ErrorType:  "required",
				Message:    "Address is required",
				Severity:   "high",
				Suggestion: "Provide a complete address",
			})
		}

		// Validate postal code format for region
		if address.PostalCode != "" && !v.isValidPostalCodeFormat(address.PostalCode, v.region) {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("addresses[%d].postal_code", i),
				ErrorType:  "format",
				Message:    "Postal code format is invalid for region",
				Severity:   "medium",
				Suggestion: fmt.Sprintf("Use valid postal code format for %s", v.region),
			})
		}
	}

	return errors
}

// ValidateBusinessHours validates localized business hours
func (v *BasicContentValidator) ValidateBusinessHours(hours *LocalizedBusinessHours) []LocalizationError {
	var errors []LocalizationError

	// Validate that at least some hours are provided
	if len(hours.RegularHours) == 0 {
		errors = append(errors, LocalizationError{
			Field:      "business_hours",
			ErrorType:  "required",
			Message:    "Business hours are required",
			Severity:   "medium",
			Suggestion: "Provide business hours for at least some days of the week",
		})
		return errors
	}

	// Validate each day's hours
	validDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for day, dayHours := range hours.RegularHours {
		// Check if day is valid
		if !v.isValidDayName(day, validDays) {
			errors = append(errors, LocalizationError{
				Field:      fmt.Sprintf("business_hours.%s", day),
				ErrorType:  "format",
				Message:    fmt.Sprintf("Invalid day name: %s", day),
				Severity:   "medium",
				Suggestion: "Use standard day names (monday, tuesday, etc.)",
			})
		}

		// Validate time format if hours are provided
		if dayHours.IsOpen {
			if dayHours.OpenTime == "" || dayHours.CloseTime == "" {
				errors = append(errors, LocalizationError{
					Field:      fmt.Sprintf("business_hours.%s", day),
					ErrorType:  "required",
					Message:    "Open and close times are required for open days",
					Severity:   "medium",
					Suggestion: "Provide both open and close times",
				})
			} else if !v.isValidTimeFormat(dayHours.OpenTime) || !v.isValidTimeFormat(dayHours.CloseTime) {
				errors = append(errors, LocalizationError{
					Field:      fmt.Sprintf("business_hours.%s", day),
					ErrorType:  "format",
					Message:    "Invalid time format",
					Severity:   "medium",
					Suggestion: "Use valid time format (HH:mm or h:mm a)",
				})
			}
		}
	}

	// Validate timezone
	if hours.Timezone == "" {
		errors = append(errors, LocalizationError{
			Field:      "business_hours.timezone",
			ErrorType:  "required",
			Message:    "Timezone is required for business hours",
			Severity:   "medium",
			Suggestion: "Specify the timezone for business hours",
		})
	}

	return errors
}

// ValidateCompliance validates localized compliance information
func (v *BasicContentValidator) ValidateCompliance(compliance *LocalizedComplianceInfo) []LocalizationError {
	var errors []LocalizationError

	// Validate that required regulations are addressed for the region
	requiredRegulations := v.getRequiredRegulationsForRegion(v.region)
	for _, required := range requiredRegulations {
		found := false
		for _, regulation := range compliance.LocalRegulations {
			if regulation.Name == required {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, LocalizationError{
				Field:      "compliance.regulations",
				ErrorType:  "required",
				Message:    fmt.Sprintf("Required regulation '%s' is missing", required),
				Severity:   "high",
				Suggestion: fmt.Sprintf("Add compliance information for %s", required),
			})
		}
	}

	// Validate privacy policies
	if len(compliance.PrivacyPolicies) == 0 && v.isPrivacyPolicyRequired(v.region) {
		errors = append(errors, LocalizationError{
			Field:      "compliance.privacy_policy",
			ErrorType:  "required",
			Message:    "Privacy policy is required for this region",
			Severity:   "high",
			Suggestion: "Provide a privacy policy compliant with regional regulations",
		})
	}

	// Validate terms of service
	if len(compliance.TermsOfService) == 0 && v.isTermsOfServiceRequired(v.region) {
		errors = append(errors, LocalizationError{
			Field:      "compliance.terms_of_service",
			ErrorType:  "required",
			Message:    "Terms of service are required for this region",
			Severity:   "high",
			Suggestion: "Provide terms of service compliant with regional regulations",
		})
	}

	return errors
}

// Helper validation methods

func (v *BasicContentValidator) isValidPhoneFormat(phone, region string) bool {
	// Define phone patterns for different regions
	patterns := map[string]string{
		"US": `^\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}$`,
		"CA": `^\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}$`,
		"GB": `^\+?44[-.\s]?[0-9]{2,4}[-.\s]?[0-9]{3,4}[-.\s]?[0-9]{3,4}$`,
		"DE": `^\+?49[-.\s]?[0-9]{2,4}[-.\s]?[0-9]{3,8}$`,
		"FR": `^\+?33[-.\s]?[0-9][-.\s]?[0-9]{2}[-.\s]?[0-9]{2}[-.\s]?[0-9]{2}[-.\s]?[0-9]{2}$`,
	}

	pattern, exists := patterns[region]
	if !exists {
		pattern = patterns["US"] // Default to US format
	}

	matched, err := regexp.MatchString(pattern, phone)
	return err == nil && matched
}

func (v *BasicContentValidator) isValidEmailFormat(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	return err == nil && matched
}

func (v *BasicContentValidator) isValidPostalCodeFormat(postalCode, region string) bool {
	patterns := map[string]string{
		"US": `^\d{5}(-\d{4})?$`,
		"CA": `^[A-Z]\d[A-Z] \d[A-Z]\d$`,
		"GB": `^[A-Z]{1,2}\d[A-Z\d]? \d[A-Z]{2}$`,
		"DE": `^\d{5}$`,
		"FR": `^\d{5}$`,
	}

	pattern, exists := patterns[region]
	if !exists {
		return true // If no pattern defined, assume valid
	}

	matched, err := regexp.MatchString(pattern, postalCode)
	return err == nil && matched
}

func (v *BasicContentValidator) isValidDayName(day string, validDays []string) bool {
	dayLower := strings.ToLower(day)
	for _, validDay := range validDays {
		if dayLower == validDay {
			return true
		}
	}
	return false
}

func (v *BasicContentValidator) isValidTimeFormat(timeStr string) bool {
	timeFormats := []string{
		"15:04",      // 24-hour format
		"3:04 PM",    // 12-hour format
		"3:04 pm",    // 12-hour format lowercase
		"15:04:05",   // 24-hour with seconds
		"3:04:05 PM", // 12-hour with seconds
	}

	for _, format := range timeFormats {
		if _, err := time.Parse(format, timeStr); err == nil {
			return true
		}
	}
	return false
}

func (v *BasicContentValidator) getRequiredRegulationsForRegion(region string) []string {
	regulations := map[string][]string{
		"EU": {"GDPR"},
		"CA": {"PIPEDA"},
		"US": {"CCPA"}, // California-specific, would need more granular region handling
		"GB": {"UK GDPR"},
	}

	if reqs, exists := regulations[region]; exists {
		return reqs
	}
	return []string{} // No specific requirements
}

func (v *BasicContentValidator) isPrivacyPolicyRequired(region string) bool {
	// Regions that require privacy policies
	requiredRegions := []string{"EU", "CA", "US", "GB"}
	for _, req := range requiredRegions {
		if region == req {
			return true
		}
	}
	return false
}

func (v *BasicContentValidator) isTermsOfServiceRequired(region string) bool {
	// Regions that require terms of service
	requiredRegions := []string{"EU", "US", "GB"}
	for _, req := range requiredRegions {
		if region == req {
			return true
		}
	}
	return false
}
