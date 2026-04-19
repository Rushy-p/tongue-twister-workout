package infrastructure

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FileStorage provides JSON-based file persistence
type FileStorage struct {
	dataDir string
}

// NewFileStorage creates a new FileStorage with the given data directory
func NewFileStorage(dataDir string) *FileStorage {
	if dataDir == "" {
		dataDir = "./data"
	}
	return &FileStorage{dataDir: dataDir}
}

// EnsureDir creates the data directory if it does not exist
func (s *FileStorage) EnsureDir() error {
	return os.MkdirAll(s.dataDir, 0755)
}

// SaveJSON marshals v to JSON and writes it to dataDir/filename
func (s *FileStorage) SaveJSON(filename string, v interface{}) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dataDir, filename), data, 0644)
}

// LoadJSON reads dataDir/filename and unmarshals JSON into v
func (s *FileStorage) LoadJSON(filename string, v interface{}) error {
	data, err := os.ReadFile(filepath.Join(s.dataDir, filename))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Exists returns true if dataDir/filename exists
func (s *FileStorage) Exists(filename string) bool {
	_, err := os.Stat(filepath.Join(s.dataDir, filename))
	return err == nil
}

// Delete removes dataDir/filename
func (s *FileStorage) Delete(filename string) error {
	path := filepath.Join(s.dataDir, filename)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return errors.New("file not found")
	}
	return os.Remove(path)
}

// List returns filenames in dataDir that start with prefix
func (s *FileStorage) List(prefix string) ([]string, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	var result []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			result = append(result, e.Name())
		}
	}
	return result, nil
}
