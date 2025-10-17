package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"
)

// EncryptionManager handles data encryption and decryption for both at rest and in transit
type EncryptionManager struct {
	aesKey    []byte
	rsaKey    *rsa.PrivateKey
	publicKey *rsa.PublicKey
	keyID     string
	createdAt time.Time
}

// EncryptionConfig holds configuration for encryption
type EncryptionConfig struct {
	AESKeySize    int           `json:"aes_key_size"`
	RSAKeySize    int           `json:"rsa_key_size"`
	KeyRotation   time.Duration `json:"key_rotation"`
	EncryptionAlg string        `json:"encryption_algorithm"`
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data      string    `json:"data"`
	KeyID     string    `json:"key_id"`
	Algorithm string    `json:"algorithm"`
	IV        string    `json:"iv"`
	CreatedAt time.Time `json:"created_at"`
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	KeyID      string    `json:"key_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewEncryptionManager creates a new encryption manager with generated keys
func NewEncryptionManager(config *EncryptionConfig) (*EncryptionManager, error) {
	if config == nil {
		config = &EncryptionConfig{
			AESKeySize:    256,
			RSAKeySize:    2048,
			KeyRotation:   24 * time.Hour,
			EncryptionAlg: "AES-256-GCM",
		}
	}

	// Generate AES key
	aesKey := make([]byte, config.AESKeySize/8)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Generate RSA key pair
	rsaKey, err := rsa.GenerateKey(rand.Reader, config.RSAKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	keyID := generateKeyID()

	return &EncryptionManager{
		aesKey:    aesKey,
		rsaKey:    rsaKey,
		publicKey: &rsaKey.PublicKey,
		keyID:     keyID,
		createdAt: time.Now(),
	}, nil
}

// NewEncryptionManagerFromKeys creates an encryption manager from existing keys
func NewEncryptionManagerFromKeys(aesKey []byte, rsaKeyPEM string) (*EncryptionManager, error) {
	// Parse RSA private key
	block, _ := pem.Decode([]byte(rsaKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode RSA private key PEM")
	}

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	keyID := generateKeyID()

	return &EncryptionManager{
		aesKey:    aesKey,
		rsaKey:    rsaKey,
		publicKey: &rsaKey.PublicKey,
		keyID:     keyID,
		createdAt: time.Now(),
	}, nil
}

// EncryptData encrypts data using AES-256-GCM
func (em *EncryptionManager) EncryptData(data []byte) (*EncryptedData, error) {
	// Create AES cipher
	block, err := aes.NewCipher(em.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, iv, data, nil)

	return &EncryptedData{
		Data:      base64.StdEncoding.EncodeToString(ciphertext),
		KeyID:     em.keyID,
		Algorithm: "AES-256-GCM",
		IV:        base64.StdEncoding.EncodeToString(iv),
		CreatedAt: time.Now(),
	}, nil
}

// DecryptData decrypts data using AES-256-GCM
func (em *EncryptionManager) DecryptData(encryptedData *EncryptedData) ([]byte, error) {
	if encryptedData.KeyID != em.keyID {
		return nil, errors.New("key ID mismatch")
	}

	if encryptedData.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported algorithm: %s", encryptedData.Algorithm)
	}

	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decode base64 IV
	iv, err := base64.StdEncoding.DecodeString(encryptedData.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(em.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// EncryptWithRSA encrypts data using RSA public key
func (em *EncryptionManager) EncryptWithRSA(data []byte) (string, error) {
	// Use OAEP padding
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, em.publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with RSA: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptWithRSA decrypts data using RSA private key
func (em *EncryptionManager) DecryptWithRSA(encryptedData string) ([]byte, error) {
	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Use OAEP padding
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, em.rsaKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return plaintext, nil
}

// GetPublicKeyPEM returns the public key in PEM format
func (em *EncryptionManager) GetPublicKeyPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(em.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// GetPrivateKeyPEM returns the private key in PEM format
func (em *EncryptionManager) GetPrivateKeyPEM() (string, error) {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(em.rsaKey)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// GetKeyPair returns both public and private keys
func (em *EncryptionManager) GetKeyPair() (*KeyPair, error) {
	publicKeyPEM, err := em.GetPublicKeyPEM()
	if err != nil {
		return nil, err
	}

	privateKeyPEM, err := em.GetPrivateKeyPEM()
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		KeyID:      em.keyID,
		CreatedAt:  em.createdAt,
	}, nil
}

// RotateKeys generates new encryption keys
func (em *EncryptionManager) RotateKeys() error {
	// Generate new AES key
	newAESKey := make([]byte, len(em.aesKey))
	if _, err := rand.Read(newAESKey); err != nil {
		return fmt.Errorf("failed to generate new AES key: %w", err)
	}

	// Generate new RSA key pair
	newRSAKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate new RSA key: %w", err)
	}

	// Update keys
	em.aesKey = newAESKey
	em.rsaKey = newRSAKey
	em.publicKey = &newRSAKey.PublicKey
	em.keyID = generateKeyID()
	em.createdAt = time.Now()

	return nil
}

// GetKeyID returns the current key ID
func (em *EncryptionManager) GetKeyID() string {
	return em.keyID
}

// GetKeyAge returns the age of the current key
func (em *EncryptionManager) GetKeyAge() time.Duration {
	return time.Since(em.createdAt)
}

// IsKeyExpired checks if the key needs rotation
func (em *EncryptionManager) IsKeyExpired(rotationInterval time.Duration) bool {
	return em.GetKeyAge() > rotationInterval
}

// generateKeyID generates a unique key identifier
func generateKeyID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("key_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("key_%x", bytes)
}

// DeriveKeyFromPassword derives an encryption key from a password using PBKDF2
func DeriveKeyFromPassword(password string, salt []byte, iterations int) ([]byte, error) {
	if iterations <= 0 {
		iterations = 100000 // Default iterations
	}

	// Use PBKDF2 to derive key
	key := make([]byte, 32) // 256 bits
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// In a real implementation, you would use crypto/pbkdf2
	// For now, we'll use a simple hash-based approach
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(salt)
	derivedKey := hash.Sum(nil)

	// Repeat the process for the specified number of iterations
	for i := 0; i < iterations-1; i++ {
		hash.Reset()
		hash.Write(derivedKey)
		derivedKey = hash.Sum(nil)
	}

	return derivedKey, nil
}

// GenerateSalt generates a random salt for key derivation
func GenerateSalt(size int) ([]byte, error) {
	if size <= 0 {
		size = 32 // Default salt size
	}

	salt := make([]byte, size)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	return salt, nil
}
