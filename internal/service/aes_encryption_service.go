package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// AESEncryptionService implements EncryptionService using AES-256-GCM (Req 14.2)
type AESEncryptionService struct {
	key   []byte // 32-byte AES-256 key
	keyID string
}

// NewAESEncryptionService creates a new AESEncryptionService.
// key must be exactly 32 bytes for AES-256; returns an error if the length is wrong.
func NewAESEncryptionService(key []byte) (*AESEncryptionService, error) {
	if len(key) != 32 {
		return nil, errors.New("AES-256 key must be exactly 32 bytes")
	}
	// Derive a stable key ID from the key material (first 8 hex chars of SHA-256)
	sum := sha256.Sum256(key)
	keyID := hex.EncodeToString(sum[:])[:8]
	return &AESEncryptionService{key: key, keyID: keyID}, nil
}

// NewAESEncryptionServiceFromEnv reads the key from the ENCRYPTION_KEY environment
// variable (hex-encoded 32 bytes). If the variable is not set, a random key is
// generated (suitable for local / single-session use).
func NewAESEncryptionServiceFromEnv() (*AESEncryptionService, error) {
	raw := os.Getenv("ENCRYPTION_KEY")
	if raw == "" {
		key := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, fmt.Errorf("failed to generate random key: %w", err)
		}
		return NewAESEncryptionService(key)
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY is not valid hex: %w", err)
	}
	return NewAESEncryptionService(key)
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce and returns
// a JSON-encoded EncryptedPayload as bytes.
func (s *AESEncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
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

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	payload := EncryptedPayload{
		KeyID:     s.keyID,
		Algorithm: "AES-256-GCM",
		Data:      base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	out, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return out, nil
}

// Decrypt parses a JSON-encoded EncryptedPayload and returns the original plaintext.
func (s *AESEncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	var payload EncryptedPayload
	if err := json.Unmarshal(ciphertext, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse encrypted payload: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}
	data, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}

// GenerateKey generates 32 random bytes suitable for AES-256.
func (s *AESEncryptionService) GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// KeyID returns the identifier for the current encryption key.
func (s *AESEncryptionService) KeyID() string {
	return s.keyID
}
