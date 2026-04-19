package infrastructure

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupManager handles backup and recovery of user data files.
// Note: FileStorage now has built-in Backup/Restore methods.
// BackupManager provides per-user backup operations on top of that.
type BackupManager struct {
	storage   *FileStorage
	backupDir string
}

// NewBackupManager creates a new BackupManager
func NewBackupManager(storage *FileStorage) *BackupManager {
	return &BackupManager{
		storage:   storage,
		backupDir: filepath.Join(storage.dataDir, "backups"),
	}
}

// CreateBackup copies the current data file into a user-tagged timestamped backup
func (m *BackupManager) CreateBackup(userID string) error {
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	src := filepath.Join(m.storage.dataDir, "userdata.json")
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return errors.New("no data file to back up")
	}

	timestamp := time.Now().Format("20060102150405")
	dst := filepath.Join(m.backupDir, fmt.Sprintf("userdata_%s_%s.json", userID, timestamp))

	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	return nil
}

// RestoreBackup restores data from a specific backup ID
func (m *BackupManager) RestoreBackup(userID string, backupID string) error {
	srcPath := filepath.Join(m.backupDir, backupID)
	if _, err := os.Stat(srcPath); errors.Is(err, os.ErrNotExist) {
		return errors.New("backup not found")
	}
	return m.storage.RestoreFromBackup(srcPath)
}

// ListBackups returns all backup IDs for a user
func (m *BackupManager) ListBackups(userID string) ([]string, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	prefix := fmt.Sprintf("userdata_%s_", userID)
	var result []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

// GetLatestBackup returns the most recent backup ID for a user
func (m *BackupManager) GetLatestBackup(userID string) (string, error) {
	backups, err := m.ListBackups(userID)
	if err != nil {
		return "", err
	}
	if len(backups) == 0 {
		return "", errors.New("no backups found for user")
	}
	latest := backups[0]
	for _, b := range backups[1:] {
		if b > latest {
			latest = b
		}
	}
	return latest, nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
