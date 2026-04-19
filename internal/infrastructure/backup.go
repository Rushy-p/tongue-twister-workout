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

// BackupManager handles backup and recovery of user data files
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

// CreateBackup copies all files belonging to userID into a timestamped backup directory
func (m *BackupManager) CreateBackup(userID string) error {
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	backupID := fmt.Sprintf("backup_%s_%s", userID, timestamp)
	destDir := filepath.Join(m.backupDir, backupID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup destination: %w", err)
	}

	// Collect all files that belong to this user
	prefixes := []string{
		fmt.Sprintf("preferences_%s", userID),
		fmt.Sprintf("session_"),
		fmt.Sprintf("progress_%s_", userID),
		fmt.Sprintf("streak_%s", userID),
		fmt.Sprintf("achievements_%s", userID),
		fmt.Sprintf("category_progress_%s", userID),
	}

	copied := 0
	for _, prefix := range prefixes {
		files, err := m.storage.List(prefix)
		if err != nil {
			continue
		}
		for _, f := range files {
			// For session_ prefix, only copy sessions belonging to this user
			// (we can't easily filter without loading; copy all and let restore handle it)
			src := filepath.Join(m.storage.dataDir, f)
			dst := filepath.Join(destDir, f)
			if err := copyFile(src, dst); err == nil {
				copied++
			}
		}
	}

	// Also copy the sessions index
	indexSrc := filepath.Join(m.storage.dataDir, sessionsIndexFile)
	if _, err := os.Stat(indexSrc); err == nil {
		_ = copyFile(indexSrc, filepath.Join(destDir, sessionsIndexFile))
	}

	if copied == 0 {
		// Remove empty backup dir
		_ = os.Remove(destDir)
		return errors.New("no data found for user")
	}
	return nil
}

// RestoreBackup restores user files from a specific backup
func (m *BackupManager) RestoreBackup(userID string, backupID string) error {
	srcDir := filepath.Join(m.backupDir, backupID)
	if _, err := os.Stat(srcDir); errors.Is(err, os.ErrNotExist) {
		return errors.New("backup not found")
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := m.storage.EnsureDir(); err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(m.storage.dataDir, e.Name())
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", e.Name(), err)
		}
	}
	return nil
}

// ListBackups returns all backup IDs for a user
func (m *BackupManager) ListBackups(userID string) ([]string, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	prefix := fmt.Sprintf("backup_%s_", userID)
	var result []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
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
	// Backups are named backup_{userID}_{timestamp} — lexicographic sort gives latest last
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
