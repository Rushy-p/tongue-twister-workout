package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

// PreferencesRepository defines the interface for user preferences data access
type PreferencesRepository interface {
	Get(userID string) (*domain.UserPreferences, error)
	Save(preferences *domain.UserPreferences) error
	Delete(userID string) error
	Export(userID string, format domain.ExportFormat) (string, error)
}

// InMemoryPreferencesRepository provides in-memory storage for user preferences
type InMemoryPreferencesRepository struct {
	preferences map[string]domain.UserPreferences
	mu          sync.RWMutex
}

// NewInMemoryPreferencesRepository creates a new in-memory preferences repository
func NewInMemoryPreferencesRepository() *InMemoryPreferencesRepository {
	return &InMemoryPreferencesRepository{
		preferences: make(map[string]domain.UserPreferences),
	}
}

// Get returns preferences for a user
func (r *InMemoryPreferencesRepository) Get(userID string) (*domain.UserPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if prefs, exists := r.preferences[userID]; exists {
		return &prefs, nil
	}
	// Return default preferences if none exist
	defaultPrefs := domain.NewUserPreferences(userID)
	return defaultPrefs, nil
}

// Save stores user preferences
func (r *InMemoryPreferencesRepository) Save(preferences *domain.UserPreferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if preferences.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	preferences.UpdatedAt = time.Now()
	r.preferences[preferences.UserID] = *preferences
	return nil
}

// Delete removes preferences for a user
func (r *InMemoryPreferencesRepository) Delete(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.preferences[userID]; !exists {
		return errors.New("preferences not found")
	}
	delete(r.preferences, userID)
	return nil
}

// Export exports user preferences in the specified format
func (r *InMemoryPreferencesRepository) Export(userID string, format domain.ExportFormat) (string, error) {
	r.mu.RLock()
	prefs, exists := r.preferences[userID]
	r.mu.RUnlock()

	if !exists {
		return "", errors.New("preferences not found")
	}

	switch format {
	case domain.ExportFormatJSON:
		data, err := json.MarshalIndent(prefs, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal preferences: %w", err)
		}
		return string(data), nil

	case domain.ExportFormatCSV:
		return r.exportToCSV(prefs), nil

	default:
		return "", errors.New("unsupported export format")
	}
}

// exportToCSV converts preferences to CSV format
func (r *InMemoryPreferencesRepository) exportToCSV(prefs domain.UserPreferences) string {
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