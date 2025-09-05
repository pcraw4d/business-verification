package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	postalCode := "k1a0a6"
	cleaned := strings.ToUpper(regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(postalCode, ""))
	fmt.Printf("Original: %s\n", postalCode)
	fmt.Printf("Cleaned: %s\n", cleaned)
	fmt.Printf("Length: %d\n", len(cleaned))
	if len(cleaned) == 6 {
		result := cleaned[:3] + " " + cleaned[3:]
		fmt.Printf("Result: %s\n", result)
	}
}
