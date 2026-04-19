package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

// StoredData is the top-level structure written to disk (Req 14.1)
type StoredData struct {
	Sessions         []domain.PracticeSession                        `json:"sessions"`
	ProgressRecords  []domain.ProgressRecord                         `json:"progress_records"`
	StreakRecords     map[string]domain.StreakRecord                  `json:"streak_records"`
	CategoryProgress map[string]map[string]domain.CategoryProgress   `json:"category_progress"`
	Achievements     map[string][]domain.Achievement                 `json:"achievements"`
	Preferences      map[string]domain.UserPreferences               `json:"preferences"`
	Version          string                                          `json:"version"`
	SavedAt          time.Time                                       `json:"saved_at"`
}

// FileStorage handles JSON-based persistence for all user data (Req 14.1)
type FileStorage struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStorage creates a new FileStorage, creating the data directory if needed
func NewFileStorage(dataDir string) (*FileStorage, error) {
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	return &FileStorage{dataDir: dataDir}, nil
}

// DataDir returns the data directory path
func (fs *FileStorage) DataDir() string {
	return fs.dataDir
}

// dataFilePath returns the path to the main data file
func (fs *FileStorage) dataFilePath() string {
	return filepath.Join(fs.dataDir, "userdata.json")
}

// Load reads the stored data from disk. Returns empty StoredData if file doesn't exist.
func (fs *FileStorage) Load() (*StoredData, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.dataFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return fs.emptyStoredData(), nil
		}
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	var stored StoredData
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	// Ensure maps are initialised
	fs.initStoredData(&stored)
	return &stored, nil
}

// Save writes the stored data to disk using an atomic write (temp file + rename)
func (fs *FileStorage) Save(data *StoredData) error {
	// Auto-backup before saving (Req 14.4 recovery point)
	_ = fs.backupInternal()

	fs.mu.Lock()
	defer fs.mu.Unlock()

	data.SavedAt = time.Now()
	if data.Version == "" {
		data.Version = "1.0"
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := fs.dataFilePath() + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, fs.dataFilePath()); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

// emptyStoredData returns an initialised empty StoredData
func (fs *FileStorage) emptyStoredData() *StoredData {
	return &StoredData{
		Sessions:         []domain.PracticeSession{},
		ProgressRecords:  []domain.ProgressRecord{},
		StreakRecords:     make(map[string]domain.StreakRecord),
		CategoryProgress: make(map[string]map[string]domain.CategoryProgress),
		Achievements:     make(map[string][]domain.Achievement),
		Preferences:      make(map[string]domain.UserPreferences),
		Version:          "1.0",
	}
}

// initStoredData ensures all maps in StoredData are non-nil
func (fs *FileStorage) initStoredData(d *StoredData) {
	if d.StreakRecords == nil {
		d.StreakRecords = make(map[string]domain.StreakRecord)
	}
	if d.CategoryProgress == nil {
		d.CategoryProgress = make(map[string]map[string]domain.CategoryProgress)
	}
	if d.Achievements == nil {
		d.Achievements = make(map[string][]domain.Achievement)
	}
	if d.Preferences == nil {
		d.Preferences = make(map[string]domain.UserPreferences)
	}
	if d.Sessions == nil {
		d.Sessions = []domain.PracticeSession{}
	}
	if d.ProgressRecords == nil {
		d.ProgressRecords = []domain.ProgressRecord{}
	}
}

// backupInternal creates a backup without acquiring the write lock (called from Save)
func (fs *FileStorage) backupInternal() error {
	src := fs.dataFilePath()
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // nothing to back up yet
	}

	backupDir := filepath.Join(fs.dataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102150405")
	dst := filepath.Join(backupDir, fmt.Sprintf("userdata_%s.json", timestamp))

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return err
	}

	// Keep only the last 5 backups
	return fs.pruneBackups(backupDir, 5)
}

// pruneBackups removes old backups keeping only the most recent `keep` files
func (fs *FileStorage) pruneBackups(backupDir string, keep int) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	// Collect backup files (sorted by name = sorted by timestamp)
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}

	// Sort ascending (oldest first) — lexicographic on timestamp filenames works
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i] > files[j] {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Delete oldest files beyond the keep limit
	for len(files) > keep {
		_ = os.Remove(filepath.Join(backupDir, files[0]))
		files = files[1:]
	}
	return nil
}

// Backup creates a timestamped backup of the current data file (Req 14.4)
func (fs *FileStorage) Backup() (string, error) {
	src := fs.dataFilePath()
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return "", fmt.Errorf("no data file to back up")
	}

	backupDir := filepath.Join(fs.dataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	dst := filepath.Join(backupDir, fmt.Sprintf("userdata_%s.json", timestamp))

	fs.mu.RLock()
	data, err := os.ReadFile(src)
	fs.mu.RUnlock()
	if err != nil {
		return "", fmt.Errorf("failed to read data file: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	_ = fs.pruneBackups(backupDir, 5)
	return dst, nil
}

// ListBackups returns all available backup file paths
func (fs *FileStorage) ListBackups() ([]string, error) {
	backupDir := filepath.Join(fs.dataDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var result []string
	for _, e := range entries {
		if !e.IsDir() {
			result = append(result, filepath.Join(backupDir, e.Name()))
		}
	}
	return result, nil
}

// RestoreFromBackup restores data from a specific backup file
func (fs *FileStorage) RestoreFromBackup(backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate the backup is parseable
	var stored StoredData
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("backup file is invalid JSON: %w", err)
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	tmpPath := fs.dataFilePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write restore temp file: %w", err)
	}
	if err := os.Rename(tmpPath, fs.dataFilePath()); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to restore data file: %w", err)
	}
	return nil
}

// ValidateData checks data integrity of the stored file
func (fs *FileStorage) ValidateData() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.dataFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no data yet is valid
		}
		return fmt.Errorf("failed to read data file: %w", err)
	}

	var stored StoredData
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("data file contains invalid JSON: %w", err)
	}

	if stored.Version == "" {
		return fmt.Errorf("data file missing version field")
	}

	return nil
}

// DeleteUserData removes all data for a user (Req 14.4)
func (fs *FileStorage) DeleteUserData(userID string) error {
	stored, err := fs.Load()
	if err != nil {
		return err
	}

	// Filter sessions
	filtered := []domain.PracticeSession{}
	for _, s := range stored.Sessions {
		if s.UserID != userID {
			filtered = append(filtered, s)
		}
	}
	stored.Sessions = filtered

	// Filter progress records
	filteredP := []domain.ProgressRecord{}
	for _, p := range stored.ProgressRecords {
		if p.UserID != userID {
			filteredP = append(filteredP, p)
		}
	}
	stored.ProgressRecords = filteredP

	delete(stored.StreakRecords, userID)
	delete(stored.CategoryProgress, userID)
	delete(stored.Achievements, userID)
	delete(stored.Preferences, userID)

	return fs.Save(stored)
}

// ExportUserData exports all data for a user in the specified format (Req 9.7, 14.5)
func (fs *FileStorage) ExportUserData(userID string, format domain.ExportFormat) (string, error) {
	stored, err := fs.Load()
	if err != nil {
		return "", err
	}

	// Collect user-specific data
	var sessions []domain.PracticeSession
	for _, s := range stored.Sessions {
		if s.UserID == userID {
			sessions = append(sessions, s)
		}
	}

	var progressRecords []domain.ProgressRecord
	for _, p := range stored.ProgressRecords {
		if p.UserID == userID {
			progressRecords = append(progressRecords, p)
		}
	}

	streak, _ := stored.StreakRecords[userID]
	catProgress, _ := stored.CategoryProgress[userID]
	achievements, _ := stored.Achievements[userID]
	prefs, _ := stored.Preferences[userID]

	switch format {
	case domain.ExportFormatJSON:
		export := map[string]interface{}{
			"user_id":          userID,
			"sessions":         sessions,
			"progress_records": progressRecords,
			"streak":           streak,
			"category_progress": catProgress,
			"achievements":     achievements,
			"preferences":      prefs,
			"exported_at":      time.Now(),
		}
		out, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal export: %w", err)
		}
		return string(out), nil

	case domain.ExportFormatCSV:
		return buildCSVExport(userID, sessions, progressRecords, streak, catProgress, achievements, prefs), nil

	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// buildCSVExport builds a multi-section CSV export for a user
func buildCSVExport(
	userID string,
	sessions []domain.PracticeSession,
	progressRecords []domain.ProgressRecord,
	streak domain.StreakRecord,
	catProgress map[string]domain.CategoryProgress,
	achievements []domain.Achievement,
	prefs domain.UserPreferences,
) string {
	csv := "# User Data Export\n"
	csv += fmt.Sprintf("user_id,%s\n", userID)
	csv += fmt.Sprintf("exported_at,%s\n", time.Now().Format(time.RFC3339))

	csv += "\n# Sessions\n"
	csv += "session_id,user_id,start_time,status,exercise_count,total_duration\n"
	for _, s := range sessions {
		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d\n",
			s.ID, s.UserID, s.StartTime.Format(time.RFC3339),
			s.Status, len(s.Exercises), s.TotalDuration)
	}

	csv += "\n# Progress Records\n"
	csv += "id,user_id,date,category,exercise_count,duration,completed\n"
	for _, p := range progressRecords {
		csv += fmt.Sprintf("%s,%s,%s,%s,%d,%d,%t\n",
			p.ID, p.UserID, p.Date.Format("2006-01-02"),
			p.Category, p.ExerciseCount, p.Duration, p.Completed)
	}

	csv += "\n# Streak\n"
	csv += "current_streak,longest_streak,last_activity\n"
	csv += fmt.Sprintf("%d,%d,%s\n",
		streak.CurrentStreak, streak.LongestStreak,
		streak.LastActivityDate.Format("2006-01-02"))

	csv += "\n# Category Progress\n"
	csv += "category,total_exercises,completed_exercises,total_time\n"
	for cat, cp := range catProgress {
		csv += fmt.Sprintf("%s,%d,%d,%d\n",
			cat, cp.TotalExercises, cp.CompletedExercises, cp.TotalTime)
	}

	csv += "\n# Achievements\n"
	csv += "id,name,type,progress,target,unlocked\n"
	for _, a := range achievements {
		csv += fmt.Sprintf("%s,%s,%s,%d,%d,%t\n",
			a.ID, a.Name, a.AchievementType, a.Progress, a.Target, a.IsUnlocked())
	}

	csv += "\n# Preferences\n"
	csv += exportPreferencesToCSV(prefs)

	return csv
}
