package infrastructure

import (
	"encoding/json"
	"fmt"

	"speech-practice-app/internal/pkg/encryption"
)

// EncryptedFileStorage wraps FileStorage with encryption for sync transmission
type EncryptedFileStorage struct {
	storage *FileStorage
	enc     encryption.EncryptionService
}

// NewEncryptedFileStorage creates a new EncryptedFileStorage
func NewEncryptedFileStorage(storage *FileStorage, enc encryption.EncryptionService) *EncryptedFileStorage {
	return &EncryptedFileStorage{storage: storage, enc: enc}
}

// SaveEncrypted marshals v to JSON, encrypts it, and writes to dataDir/filename
func (s *EncryptedFileStorage) SaveEncrypted(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	encrypted, err := s.enc.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}
	// Store raw bytes via a byte-slice wrapper
	return s.storage.SaveJSON(filename+".enc", encrypted)
}

// LoadEncrypted reads dataDir/filename, decrypts it, and unmarshals JSON into v
func (s *EncryptedFileStorage) LoadEncrypted(filename string, v interface{}) error {
	var encrypted []byte
	if err := s.storage.LoadJSON(filename+".enc", &encrypted); err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}
	decrypted, err := s.enc.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	return json.Unmarshal(decrypted, v)
}
