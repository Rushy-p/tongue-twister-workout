package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// SaveEncrypted marshals v to JSON, encrypts it, and writes to dataDir/filename.enc
func (s *EncryptedFileStorage) SaveEncrypted(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	encrypted, err := s.enc.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}
	encData, err := json.Marshal(encrypted)
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted data: %w", err)
	}
	path := filepath.Join(s.storage.dataDir, filename+".enc")
	if err := os.MkdirAll(s.storage.dataDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, encData, 0644)
}

// LoadEncrypted reads dataDir/filename.enc, decrypts it, and unmarshals JSON into v
func (s *EncryptedFileStorage) LoadEncrypted(filename string, v interface{}) error {
	path := filepath.Join(s.storage.dataDir, filename+".enc")
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}
	var encrypted []byte
	if err := json.Unmarshal(raw, &encrypted); err != nil {
		return fmt.Errorf("failed to unmarshal encrypted wrapper: %w", err)
	}
	decrypted, err := s.enc.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}
	return json.Unmarshal(decrypted, v)
}
