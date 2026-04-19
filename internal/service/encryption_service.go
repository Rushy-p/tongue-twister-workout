package service

import "encoding/json"

// EncryptionService defines the interface for data encryption operations (Req 14.2)
type EncryptionService interface {
	// Encrypt encrypts plaintext data and returns base64-encoded ciphertext
	Encrypt(plaintext []byte) ([]byte, error)
	// Decrypt decrypts base64-encoded ciphertext and returns plaintext
	Decrypt(ciphertext []byte) ([]byte, error)
	// GenerateKey generates a new random encryption key
	GenerateKey() ([]byte, error)
	// KeyID returns the identifier for the current encryption key
	KeyID() string
}

// EncryptedPayload wraps encrypted data with metadata for transmission
type EncryptedPayload struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	Data      string `json:"data"`       // base64-encoded ciphertext
	Nonce     string `json:"nonce"`      // base64-encoded nonce/IV
	CreatedAt string `json:"created_at"`
}

// PrepareForSync encrypts data for cloud sync transmission (Req 14.2)
func PrepareForSync(enc EncryptionService, data []byte) (*EncryptedPayload, error) {
	encrypted, err := enc.Encrypt(data)
	if err != nil {
		return nil, err
	}
	// The AES implementation returns a JSON-encoded EncryptedPayload as bytes.
	var payload EncryptedPayload
	if err := json.Unmarshal(encrypted, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// DecryptFromSync decrypts a received sync payload
func DecryptFromSync(enc EncryptionService, payload *EncryptedPayload) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return enc.Decrypt(raw)
}
