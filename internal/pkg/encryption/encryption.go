package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptionService defines the interface for data encryption
type EncryptionService interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	GenerateKey() ([]byte, error)
}

// NoOpEncryptionService implements EncryptionService without any encryption (local-only mode)
type NoOpEncryptionService struct{}

// Encrypt returns data unchanged
func (n *NoOpEncryptionService) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

// Decrypt returns data unchanged
func (n *NoOpEncryptionService) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}

// GenerateKey returns 32 random bytes
func (n *NoOpEncryptionService) GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// AESEncryptionService implements AES-256-GCM encryption
type AESEncryptionService struct {
	key []byte
}

// NewAESEncryptionService creates a new AESEncryptionService; key must be exactly 32 bytes
func NewAESEncryptionService(key []byte) (*AESEncryptionService, error) {
	if len(key) != 32 {
		return nil, errors.New("AES-256 key must be exactly 32 bytes")
	}
	return &AESEncryptionService{key: key}, nil
}

// Encrypt encrypts data using AES-256-GCM; the nonce is prepended to the ciphertext
func (s *AESEncryptionService) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts AES-256-GCM data; expects the nonce in the first 12 bytes
func (s *AESEncryptionService) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateKey generates a 32-byte random key suitable for AES-256
func (s *AESEncryptionService) GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
