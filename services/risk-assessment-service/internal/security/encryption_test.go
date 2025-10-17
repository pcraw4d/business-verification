package security

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptionManager(t *testing.T) {
	tests := []struct {
		name   string
		config *EncryptionConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &EncryptionConfig{
				AESKeySize:    256,
				RSAKeySize:    2048,
				KeyRotation:   24 * time.Hour,
				EncryptionAlg: "AES-256-GCM",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em, err := NewEncryptionManager(tt.config)
			require.NoError(t, err)
			assert.NotNil(t, em)
			assert.NotEmpty(t, em.GetKeyID())
			assert.Equal(t, time.Duration(0), em.GetKeyAge())
		})
	}
}

func TestNewEncryptionManagerFromKeys(t *testing.T) {
	// Create a temporary encryption manager to get keys
	tempEM, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	privateKeyPEM, err := tempEM.GetPrivateKeyPEM()
	require.NoError(t, err)

	aesKey := make([]byte, 32)
	copy(aesKey, tempEM.aesKey)

	t.Run("valid keys", func(t *testing.T) {
		em, err := NewEncryptionManagerFromKeys(aesKey, privateKeyPEM)
		require.NoError(t, err)
		assert.NotNil(t, em)
		assert.NotEmpty(t, em.GetKeyID())
	})

	t.Run("invalid private key", func(t *testing.T) {
		em, err := NewEncryptionManagerFromKeys(aesKey, "invalid key")
		assert.Error(t, err)
		assert.Nil(t, em)
	})
}

func TestEncryptionManager_EncryptData(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple text",
			data: []byte("Hello, World!"),
		},
		{
			name: "empty data",
			data: []byte(""),
		},
		{
			name: "large data",
			data: make([]byte, 1024*1024), // 1MB
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := em.EncryptData(tt.data)
			require.NoError(t, err)
			assert.NotNil(t, encrypted)
			assert.Equal(t, em.GetKeyID(), encrypted.KeyID)
			assert.Equal(t, "AES-256-GCM", encrypted.Algorithm)
			assert.NotEmpty(t, encrypted.Data)
			assert.NotEmpty(t, encrypted.IV)
			assert.False(t, encrypted.CreatedAt.IsZero())
		})
	}
}

func TestEncryptionManager_DecryptData(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	originalData := []byte("Hello, World!")
	encrypted, err := em.EncryptData(originalData)
	require.NoError(t, err)

	t.Run("successful decryption", func(t *testing.T) {
		decrypted, err := em.DecryptData(encrypted)
		require.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("key ID mismatch", func(t *testing.T) {
		encrypted.KeyID = "different_key_id"
		_, err := em.DecryptData(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key ID mismatch")
	})

	t.Run("unsupported algorithm", func(t *testing.T) {
		encrypted.Algorithm = "AES-256-CBC"
		_, err := em.DecryptData(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported algorithm")
	})

	t.Run("invalid base64 data", func(t *testing.T) {
		encrypted.Data = "invalid base64"
		_, err := em.DecryptData(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode ciphertext")
	})

	t.Run("invalid base64 IV", func(t *testing.T) {
		encrypted.IV = "invalid base64"
		_, err := em.DecryptData(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode IV")
	})
}

func TestEncryptionManager_EncryptWithRSA(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple text",
			data: []byte("Hello, World!"),
		},
		{
			name: "empty data",
			data: []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := em.EncryptWithRSA(tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
		})
	}
}

func TestEncryptionManager_DecryptWithRSA(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	originalData := []byte("Hello, World!")
	encrypted, err := em.EncryptWithRSA(originalData)
	require.NoError(t, err)

	t.Run("successful decryption", func(t *testing.T) {
		decrypted, err := em.DecryptWithRSA(encrypted)
		require.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("invalid base64 data", func(t *testing.T) {
		_, err := em.DecryptWithRSA("invalid base64")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode ciphertext")
	})
}

func TestEncryptionManager_GetPublicKeyPEM(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	publicKeyPEM, err := em.GetPublicKeyPEM()
	require.NoError(t, err)
	assert.NotEmpty(t, publicKeyPEM)
	assert.Contains(t, publicKeyPEM, "BEGIN PUBLIC KEY")
	assert.Contains(t, publicKeyPEM, "END PUBLIC KEY")
}

func TestEncryptionManager_GetPrivateKeyPEM(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	privateKeyPEM, err := em.GetPrivateKeyPEM()
	require.NoError(t, err)
	assert.NotEmpty(t, privateKeyPEM)
	assert.Contains(t, privateKeyPEM, "BEGIN RSA PRIVATE KEY")
	assert.Contains(t, privateKeyPEM, "END RSA PRIVATE KEY")
}

func TestEncryptionManager_GetKeyPair(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	keyPair, err := em.GetKeyPair()
	require.NoError(t, err)
	assert.NotNil(t, keyPair)
	assert.Equal(t, em.GetKeyID(), keyPair.KeyID)
	assert.NotEmpty(t, keyPair.PrivateKey)
	assert.NotEmpty(t, keyPair.PublicKey)
	assert.False(t, keyPair.CreatedAt.IsZero())
}

func TestEncryptionManager_RotateKeys(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	originalKeyID := em.GetKeyID()
	originalAESKey := make([]byte, len(em.aesKey))
	copy(originalAESKey, em.aesKey)

	err = em.RotateKeys()
	require.NoError(t, err)

	assert.NotEqual(t, originalKeyID, em.GetKeyID())
	assert.NotEqual(t, originalAESKey, em.aesKey)
	assert.Equal(t, time.Duration(0), em.GetKeyAge())
}

func TestEncryptionManager_GetKeyAge(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	// Key should be very new
	age := em.GetKeyAge()
	assert.True(t, age < time.Second)

	// Wait a bit and check age again
	time.Sleep(10 * time.Millisecond)
	age = em.GetKeyAge()
	assert.True(t, age >= 10*time.Millisecond)
}

func TestEncryptionManager_IsKeyExpired(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	t.Run("key not expired", func(t *testing.T) {
		expired := em.IsKeyExpired(24 * time.Hour)
		assert.False(t, expired)
	})

	t.Run("key expired", func(t *testing.T) {
		expired := em.IsKeyExpired(1 * time.Nanosecond)
		assert.True(t, expired)
	})
}

func TestDeriveKeyFromPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		salt        []byte
		iterations  int
		expectError bool
	}{
		{
			name:        "valid parameters",
			password:    "testpassword",
			salt:        []byte("testsalt"),
			iterations:  1000,
			expectError: false,
		},
		{
			name:        "zero iterations",
			password:    "testpassword",
			salt:        []byte("testsalt"),
			iterations:  0,
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			salt:        []byte("testsalt"),
			iterations:  1000,
			expectError: false,
		},
		{
			name:        "empty salt",
			password:    "testpassword",
			salt:        []byte(""),
			iterations:  1000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := DeriveKeyFromPassword(tt.password, tt.salt, tt.iterations)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, key)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				assert.Len(t, key, 32) // 256 bits
			}
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "default size",
			size: 0,
		},
		{
			name: "custom size",
			size: 16,
		},
		{
			name: "large size",
			size: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, err := GenerateSalt(tt.size)
			require.NoError(t, err)
			assert.NotNil(t, salt)

			expectedSize := tt.size
			if expectedSize <= 0 {
				expectedSize = 32 // Default size
			}
			assert.Len(t, salt, expectedSize)
		})
	}
}

func TestEncryptionManager_EndToEnd(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	// Test data
	originalData := []byte("This is sensitive business data that needs to be encrypted")

	// Encrypt data
	encrypted, err := em.EncryptData(originalData)
	require.NoError(t, err)

	// Decrypt data
	decrypted, err := em.DecryptData(encrypted)
	require.NoError(t, err)

	// Verify data integrity
	assert.Equal(t, originalData, decrypted)

	// Test RSA encryption/decryption
	rsaEncrypted, err := em.EncryptWithRSA(originalData)
	require.NoError(t, err)

	rsaDecrypted, err := em.DecryptWithRSA(rsaEncrypted)
	require.NoError(t, err)

	assert.Equal(t, originalData, rsaDecrypted)
}

func TestEncryptionManager_KeyRotation(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	// Encrypt data with original key
	originalData := []byte("Test data")
	encrypted, err := em.EncryptData(originalData)
	require.NoError(t, err)

	// Rotate keys
	err = em.RotateKeys()
	require.NoError(t, err)

	// Old encrypted data should still be decryptable (if we had key history)
	// In a real implementation, you would maintain key history
	decrypted, err := em.DecryptData(encrypted)
	require.NoError(t, err)
	assert.Equal(t, originalData, decrypted)

	// New data should be encrypted with new key
	newEncrypted, err := em.EncryptData(originalData)
	require.NoError(t, err)
	assert.NotEqual(t, encrypted.Data, newEncrypted.Data)
}

func TestEncryptionManager_ConcurrentAccess(t *testing.T) {
	em, err := NewEncryptionManager(nil)
	require.NoError(t, err)

	// Test concurrent encryption/decryption
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			data := []byte(fmt.Sprintf("Test data %d", i))

			// Encrypt
			encrypted, err := em.EncryptData(data)
			require.NoError(t, err)

			// Decrypt
			decrypted, err := em.DecryptData(encrypted)
			require.NoError(t, err)

			assert.Equal(t, data, decrypted)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
