package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// Config holds encryption configuration
type Config struct {
	KeySize     int    `json:"key_size"`
	Algorithm   string `json:"algorithm"`
	SaltRounds  int    `json:"salt_rounds"`
	KeyEncoding string `json:"key_encoding"` // hex, base64
}

// DefaultConfig returns default encryption configuration
func DefaultConfig() *Config {
	return &Config{
		KeySize:     32, // 256 bits
		Algorithm:   "AES-256-GCM",
		SaltRounds:  12,
		KeyEncoding: "hex",
	}
}

// Encryptor provides encryption and decryption capabilities
type Encryptor struct {
	config *Config
	key    []byte
}

// NewEncryptor creates a new encryptor with the given key and configuration
func NewEncryptor(key []byte, config *Config) (*Encryptor, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if len(key) != config.KeySize {
		return nil, fmt.Errorf("key size must be %d bytes, got %d", config.KeySize, len(key))
	}

	return &Encryptor{
		config: config,
		key:    key,
	}, nil
}

// GenerateKey generates a random encryption key
func GenerateKey(size int) ([]byte, error) {
	key := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts data using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string, cost int) (string, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// CheckPassword checks if a password matches a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashData creates a SHA-256 hash of data
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// EncodeKey encodes a key to the specified format
func EncodeKey(key []byte, encoding string) string {
	switch encoding {
	case "hex":
		return hex.EncodeToString(key)
	case "base64":
		return base64.StdEncoding.EncodeToString(key)
	default:
		return hex.EncodeToString(key)
	}
}

// DecodeKey decodes a key from the specified format
func DecodeKey(encodedKey, encoding string) ([]byte, error) {
	switch encoding {
	case "hex":
		return hex.DecodeString(encodedKey)
	case "base64":
		return base64.StdEncoding.DecodeString(encodedKey)
	default:
		return hex.DecodeString(encodedKey)
	}
}

// GenerateRandomBytes generates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
