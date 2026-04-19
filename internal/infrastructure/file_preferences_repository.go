package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"speech-practice-app/internal/domain"
)

// FilePreferencesRepository persists user preferences to disk
type FilePreferencesRepository struct {
	storage *FileStorage
}

// NewFilePreferencesRepository creates a new FilePreferencesRepository
func NewFilePreferencesRepository(storage *FileStorage) *FilePreferencesRepository {
	return &FilePreferencesRepository{storage: storage}
}

func (r *FilePreferencesRepository) filename(userID string) string {
	return fmt.Sprintf("preferences_%s.json", userID)
}

// Get returns preferences for a user, or default preferences if none exist
func (r *FilePreferencesRepository) Get(userID string) (*domain.UserPreferences, error) {
	var prefs domain.UserPreferences
	err := r.storage.LoadJSON(r.filename(userID), &prefs)
	if err != nil {
		return domain.NewUserPreferences(userID), nil
	}
	return &prefs, nil
}

// Save stores user preferences to disk
func (r *FilePreferencesRepository) Save(preferences *domain.UserPreferences) error {
	if preferences.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	preferences.UpdatedAt = time.Now()
	return r.storage.SaveJSON(r.filename(preferences.UserID), preferences)
}

// Delete removes preferences for a user
func (r *FilePreferencesRepository) Delete(userID string) error {
	if !r.storage.Exists(r.filename(userID)) {
		return errors.New("preferences not found")
	}
	return r.storage.Delete(r.filename(userID))
}

// Export exports user preferences in the specified format
func (r *FilePreferencesRepository) Export(userID string, format domain.ExportFormat) (string, error) {
	prefs, err := r.Get(userID)
	if err != nil {
		return "", err
	}

	switch format {
	case domain.ExportFormatJSON:
		data, err := json.MarshalIndent(prefs, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal preferences: %w", err)
		}
		return string(data), nil
	case domain.ExportFormatCSV:
		return exportPreferencesToCSV(*prefs), nil
	default:
		return "", errors.New("unsupported export format")
	}
}

func exportPreferencesToCSV(prefs domain.UserPreferences) string {
	csv := "key,value\n"
	csv += fmt.Sprintf("id,%s\n", prefs.ID)
	csv += fmt.Sprintf("user_id,%s\n", prefs.UserID)
	csv += fmt.Sprintf("difficulty,%s\n", prefs.Difficulty)
	csv += fmt.Sprintf("default_duration,%d\n", prefs.DefaultDuration)
	csv += fmt.Sprintf("audio_feedback,%t\n", prefs.AudioFeedback)
	csv += fmt.Sprintf("vibration_feedback,%t\n", prefs.VibrationFeedback)
	csv += fmt.Sprintf("reminder_enabled,%t\n", prefs.ReminderEnabled)
	csv += fmt.Sprintf("reminder_time,%s\n", prefs.ReminderTime.Format("15:04"))
	csv += fmt.Sprintf("reminder_days,%v\n", prefs.ReminderDays)
	csv += fmt.Sprintf("export_format,%s\n", prefs.ExportFormat)
	csv += fmt.Sprintf("screen_reader_enabled,%t\n", prefs.Accessibility.ScreenReaderEnabled)
	csv += fmt.Sprintf("high_contrast_mode,%t\n", prefs.Accessibility.HighContrastMode)
	csv += fmt.Sprintf("text_size_multiplier,%.2f\n", prefs.Accessibility.TextSizeMultiplier)
	csv += fmt.Sprintf("element_size_multiplier,%.2f\n", prefs.Accessibility.ElementSizeMultiplier)
	csv += fmt.Sprintf("keyboard_navigation_enabled,%t\n", prefs.Accessibility.KeyboardNavigationEnabled)
	csv += fmt.Sprintf("color_independent,%t\n", prefs.Accessibility.ColorIndependent)
	csv += fmt.Sprintf("created_at,%s\n", prefs.CreatedAt.Format(time.RFC3339))
	csv += fmt.Sprintf("updated_at,%s\n", prefs.UpdatedAt.Format(time.RFC3339))
	return csv
}
