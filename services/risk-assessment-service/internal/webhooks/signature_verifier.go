package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookSignatureVerifier implements WebhookSignatureVerifier
type DefaultWebhookSignatureVerifier struct {
	tolerance time.Duration
}

// NewDefaultWebhookSignatureVerifier creates a new default webhook signature verifier
func NewDefaultWebhookSignatureVerifier(tolerance time.Duration) *DefaultWebhookSignatureVerifier {
	return &DefaultWebhookSignatureVerifier{
		tolerance: tolerance,
	}
}

// GenerateSignature generates a webhook signature
func (sv *DefaultWebhookSignatureVerifier) GenerateSignature(payload []byte, secret string, timestamp string) (string, error) {
	// Create the message to sign: timestamp.payload
	message := fmt.Sprintf("%s.%s", timestamp, string(payload))

	// Create HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature, nil
}

// VerifySignature verifies a webhook signature
func (sv *DefaultWebhookSignatureVerifier) VerifySignature(payload []byte, signature string, secret string, timestamp string) (bool, error) {
	// Generate expected signature
	expectedSignature, err := sv.GenerateSignature(payload, secret, timestamp)
	if err != nil {
		return false, fmt.Errorf("failed to generate expected signature: %w", err)
	}

	// Compare signatures using constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}

// ValidateTimestamp validates a webhook timestamp
func (sv *DefaultWebhookSignatureVerifier) ValidateTimestamp(timestamp string, tolerance time.Duration) (bool, error) {
	// Parse timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Convert to time
	timestampTime := time.Unix(ts, 0)
	now := time.Now()

	// Check if timestamp is within tolerance
	diff := now.Sub(timestampTime)
	if diff < 0 {
		diff = -diff
	}

	return diff <= tolerance, nil
}

// GenerateWebhookSignature generates a complete webhook signature with timestamp
func (sv *DefaultWebhookSignatureVerifier) GenerateWebhookSignature(payload []byte, secret string) (*WebhookSignature, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	signature, err := sv.GenerateSignature(payload, secret, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	return &WebhookSignature{
		Algorithm: "sha256",
		Signature: signature,
		Timestamp: timestamp,
		Nonce:     generateNonce(),
	}, nil
}

// VerifyWebhookSignature verifies a complete webhook signature
func (sv *DefaultWebhookSignatureVerifier) VerifyWebhookSignature(payload []byte, signature *WebhookSignature, secret string) (bool, error) {
	// Validate timestamp
	validTimestamp, err := sv.ValidateTimestamp(signature.Timestamp, sv.tolerance)
	if err != nil {
		return false, fmt.Errorf("failed to validate timestamp: %w", err)
	}

	if !validTimestamp {
		return false, fmt.Errorf("timestamp is outside tolerance window")
	}

	// Verify signature
	validSignature, err := sv.VerifySignature(payload, signature.Signature, secret, signature.Timestamp)
	if err != nil {
		return false, fmt.Errorf("failed to verify signature: %w", err)
	}

	return validSignature, nil
}

// ParseSignatureHeader parses a webhook signature from HTTP headers
func (sv *DefaultWebhookSignatureVerifier) ParseSignatureHeader(header string) (*WebhookSignature, error) {
	// Expected format: "sha256=signature,t=timestamp,v1=version"
	parts := strings.Split(header, ",")

	signature := &WebhookSignature{
		Algorithm: "sha256",
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		kv := strings.SplitN(part, "=", 2)

		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "sha256":
			signature.Signature = value
		case "t":
			signature.Timestamp = value
		case "v1":
			// Version field, can be used for future compatibility
		}
	}

	// Validate required fields
	if signature.Signature == "" {
		return nil, fmt.Errorf("missing signature in header")
	}

	if signature.Timestamp == "" {
		return nil, fmt.Errorf("missing timestamp in header")
	}

	return signature, nil
}

// GenerateSignatureHeader generates a webhook signature header
func (sv *DefaultWebhookSignatureVerifier) GenerateSignatureHeader(payload []byte, secret string) (string, error) {
	signature, err := sv.GenerateWebhookSignature(payload, secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate webhook signature: %w", err)
	}

	// Format: "sha256=signature,t=timestamp,v1=1"
	header := fmt.Sprintf("sha256=%s,t=%s,v1=1", signature.Signature, signature.Timestamp)

	return header, nil
}

// ValidateWebhookRequest validates a complete webhook request
func (sv *DefaultWebhookSignatureVerifier) ValidateWebhookRequest(payload []byte, signatureHeader string, secret string) (bool, error) {
	// Parse signature from header
	signature, err := sv.ParseSignatureHeader(signatureHeader)
	if err != nil {
		return false, fmt.Errorf("failed to parse signature header: %w", err)
	}

	// Verify signature
	valid, err := sv.VerifyWebhookSignature(payload, signature, secret)
	if err != nil {
		return false, fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	return valid, nil
}

// GetSignatureInfo extracts signature information from a webhook request
func (sv *DefaultWebhookSignatureVerifier) GetSignatureInfo(signatureHeader string) (*SignatureInfo, error) {
	signature, err := sv.ParseSignatureHeader(signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signature header: %w", err)
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(signature.Timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	info := &SignatureInfo{
		Algorithm: signature.Algorithm,
		Signature: signature.Signature,
		Timestamp: time.Unix(timestamp, 0),
		Nonce:     signature.Nonce,
		Age:       time.Since(time.Unix(timestamp, 0)),
	}

	return info, nil
}

// Additional data structures for signature verification

// SignatureInfo represents information about a webhook signature
type SignatureInfo struct {
	Algorithm string        `json:"algorithm"`
	Signature string        `json:"signature"`
	Timestamp time.Time     `json:"timestamp"`
	Nonce     string        `json:"nonce"`
	Age       time.Duration `json:"age"`
}

// SignatureValidationResult represents the result of signature validation
type SignatureValidationResult struct {
	Valid           bool          `json:"valid"`
	Error           string        `json:"error,omitempty"`
	Algorithm       string        `json:"algorithm"`
	Timestamp       time.Time     `json:"timestamp"`
	Age             time.Duration `json:"age"`
	WithinTolerance bool          `json:"within_tolerance"`
}

// Helper functions

// generateNonce generates a random nonce for webhook signatures
func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// IsValidSignatureAlgorithm checks if the signature algorithm is supported
func IsValidSignatureAlgorithm(algorithm string) bool {
	supportedAlgorithms := []string{"sha256", "sha1"}

	for _, supported := range supportedAlgorithms {
		if algorithm == supported {
			return true
		}
	}

	return false
}

// GetSupportedAlgorithms returns the list of supported signature algorithms
func GetSupportedAlgorithms() []string {
	return []string{"sha256", "sha1"}
}

// DefaultSignatureTolerance returns the default signature tolerance
func DefaultSignatureTolerance() time.Duration {
	return 5 * time.Minute
}

// ValidateSignatureTolerance validates that the signature tolerance is reasonable
func ValidateSignatureTolerance(tolerance time.Duration) error {
	if tolerance < 0 {
		return fmt.Errorf("signature tolerance cannot be negative")
	}

	if tolerance > 1*time.Hour {
		return fmt.Errorf("signature tolerance cannot exceed 1 hour")
	}

	return nil
}
