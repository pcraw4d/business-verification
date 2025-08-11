package encryption

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.KeySize != 32 {
		t.Errorf("Expected key size 32, got %d", config.KeySize)
	}
	if config.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm AES-256-GCM, got %s", config.Algorithm)
	}
	if config.SaltRounds != 12 {
		t.Errorf("Expected salt rounds 12, got %d", config.SaltRounds)
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Test different key sizes
	key16, err := GenerateKey(16)
	if err != nil {
		t.Fatalf("Failed to generate 16-byte key: %v", err)
	}
	if len(key16) != 16 {
		t.Errorf("Expected key length 16, got %d", len(key16))
	}
}

func TestNewEncryptor(t *testing.T) {
	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key, nil)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}
	if encryptor == nil {
		t.Fatal("Encryptor should not be nil")
	}

	// Test with custom config
	config := &Config{
		KeySize:     32,
		Algorithm:   "AES-256-GCM",
		SaltRounds:  10,
		KeyEncoding: "hex",
	}
	encryptor, err = NewEncryptor(key, config)
	if err != nil {
		t.Fatalf("Failed to create encryptor with custom config: %v", err)
	}

	// Test with wrong key size
	shortKey := []byte("short")
	_, err = NewEncryptor(shortKey, nil)
	if err == nil {
		t.Error("Expected error for short key")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key, nil)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	plaintext := []byte("Hello, World!")
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("Ciphertext should not be empty")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match original. Expected %s, got %s", plaintext, decrypted)
	}
}

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password, 0)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == password {
		t.Error("Hash should not equal original password")
	}

	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password, 0)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("Password should match hash")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Error("Wrong password should not match hash")
	}
}

func TestHashData(t *testing.T) {
	data := []byte("test data")
	hash := HashData(data)

	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}

	// Hash should be consistent
	hash2 := HashData(data)
	if hash != hash2 {
		t.Error("Hash should be consistent for same data")
	}

	// Different data should produce different hashes
	data2 := []byte("different data")
	hash3 := HashData(data2)
	if hash == hash3 {
		t.Error("Different data should produce different hashes")
	}
}

func TestEncodeDecodeKey(t *testing.T) {
	key := []byte("testkey123456789012345678901234")

	// Test hex encoding
	hexEncoded := EncodeKey(key, "hex")
	decoded, err := DecodeKey(hexEncoded, "hex")
	if err != nil {
		t.Fatalf("Failed to decode hex key: %v", err)
	}
	if string(decoded) != string(key) {
		t.Error("Decoded key should match original")
	}

	// Test base64 encoding
	base64Encoded := EncodeKey(key, "base64")
	decoded, err = DecodeKey(base64Encoded, "base64")
	if err != nil {
		t.Fatalf("Failed to decode base64 key: %v", err)
	}
	if string(decoded) != string(key) {
		t.Error("Decoded key should match original")
	}

	// Test default encoding (hex)
	defaultEncoded := EncodeKey(key, "unknown")
	decoded, err = DecodeKey(defaultEncoded, "unknown")
	if err != nil {
		t.Fatalf("Failed to decode default key: %v", err)
	}
	if string(decoded) != string(key) {
		t.Error("Decoded key should match original")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	bytes, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	if len(bytes) != 16 {
		t.Errorf("Expected 16 bytes, got %d", len(bytes))
	}

	// Generate another set to ensure randomness
	bytes2, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("Failed to generate second set of random bytes: %v", err)
	}
	if string(bytes) == string(bytes2) {
		t.Error("Random bytes should be different")
	}
}

func TestGenerateRandomString(t *testing.T) {
	str, err := GenerateRandomString(10)
	if err != nil {
		t.Fatalf("Failed to generate random string: %v", err)
	}
	if len(str) != 10 {
		t.Errorf("Expected string length 10, got %d", len(str))
	}

	// Generate another string to ensure randomness
	str2, err := GenerateRandomString(10)
	if err != nil {
		t.Fatalf("Failed to generate second random string: %v", err)
	}
	if str == str2 {
		t.Error("Random strings should be different")
	}
}

func TestEncryptorWithCustomConfig(t *testing.T) {
	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	config := &Config{
		KeySize:     32,
		Algorithm:   "AES-256-GCM",
		SaltRounds:  8,
		KeyEncoding: "base64",
	}

	encryptor, err := NewEncryptor(key, config)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	plaintext := []byte("Test with custom config")
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Error("Decrypted text doesn't match original")
	}
}
