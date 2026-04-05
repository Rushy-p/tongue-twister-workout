package service

import (
	"errors"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

// PreferencesService handles user preferences management
type PreferencesService struct {
	preferencesRepo infrastructure.PreferencesRepository
}

// NewPreferencesService creates a new PreferencesService
func NewPreferencesService(preferencesRepo infrastructure.PreferencesRepository) *PreferencesService {
	return &PreferencesService{
		preferencesRepo: preferencesRepo,
	}
}

// GetPreferences returns user preferences
// Implements Requirement 9.1, 9.2, 9.3, 9.4
func (s *PreferencesService) GetPreferences(userID string) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		// Return default preferences if none exist
		return domain.NewUserPreferences(userID), nil
	}

	return prefs, nil
}

// UpdatePreferences updates user preferences
// Implements Requirement 9.5
func (s *PreferencesService) UpdatePreferences(userID string, updates map[string]interface{}) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	// Apply each update
	for key, value := range updates {
		switch key {
		case "difficulty":
			if level, ok := value.(string); ok {
				prefs.SetDifficulty(domain.DifficultyLevel(level))
			}
		case "default_duration":
			if duration, ok := value.(float64); ok {
				prefs.SetDefaultDuration(time.Duration(duration) * time.Second)
			}
		case "audio_feedback":
			if enabled, ok := value.(bool); ok {
				if enabled {
					prefs.EnableAudioFeedback()
				} else {
					prefs.DisableAudioFeedback()
				}
			}
		case "vibration_feedback":
			if enabled, ok := value.(bool); ok {
				if enabled {
					prefs.EnableVibrationFeedback()
				} else {
					prefs.DisableVibrationFeedback()
				}
			}
		case "reminder_enabled":
			if enabled, ok := value.(bool); ok {
				if enabled {
					prefs.EnableReminders()
				} else {
					prefs.DisableReminders()
				}
			}
		case "reminder_time":
			if timeStr, ok := value.(string); ok {
				// Parse time in HH:MM format
				t, err := time.Parse("15:04", timeStr)
				if err == nil {
					prefs.SetReminderTime(t.Hour(), t.Minute())
				}
			}
		case "reminder_days":
			if days, ok := value.([]interface{}); ok {
				intDays := make([]int, len(days))
				for i, d := range days {
					if f, ok := d.(float64); ok {
						intDays[i] = int(f)
					}
				}
				prefs.SetReminderDays(intDays)
			}
		case "export_format":
			if format, ok := value.(string); ok {
				prefs.SetExportFormat(domain.ExportFormat(format))
			}
		}
	}

	// Validate and save
	if err := prefs.Validate(); err != nil {
		return nil, err
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetDifficulty sets the difficulty level
// Implements Requirement 9.1
func (s *PreferencesService) SetDifficulty(userID string, level domain.DifficultyLevel) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if err := prefs.SetDifficulty(level); err != nil {
		return nil, err
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetDefaultDuration sets the default exercise duration
// Implements Requirement 9.2
func (s *PreferencesService) SetDefaultDuration(userID string, duration time.Duration) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if err := prefs.SetDefaultDuration(duration); err != nil {
		return nil, err
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetAudioFeedback enables or disables audio feedback
// Implements Requirement 9.3
func (s *PreferencesService) SetAudioFeedback(userID string, enabled bool) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if enabled {
		prefs.EnableAudioFeedback()
	} else {
		prefs.DisableAudioFeedback()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetVibrationFeedback enables or disables vibration feedback
// Implements Requirement 9.4
func (s *PreferencesService) SetVibrationFeedback(userID string, enabled bool) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if enabled {
		prefs.EnableVibrationFeedback()
	} else {
		prefs.DisableVibrationFeedback()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetReminderConfig configures reminder settings
func (s *PreferencesService) SetReminderConfig(userID string, enabled bool, hour, minute int, days []int) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	// Set reminder time
	if err := prefs.SetReminderTime(hour, minute); err != nil {
		return nil, err
	}

	// Set reminder days
	if err := prefs.SetReminderDays(days); err != nil {
		return nil, err
	}

	// Enable or disable reminders
	if enabled {
		prefs.EnableReminders()
	} else {
		prefs.DisableReminders()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetExportFormat sets the export format
func (s *PreferencesService) SetExportFormat(userID string, format domain.ExportFormat) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if err := prefs.SetExportFormat(format); err != nil {
		return nil, err
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// ExportData exports user data in the specified format
// Implements Requirement 9.7
func (s *PreferencesService) ExportData(userID string, format domain.ExportFormat) (string, error) {
	if userID == "" {
		return "", errors.New("user ID cannot be empty")
	}

	return s.preferencesRepo.Export(userID, format)
}

// DeletePreferences removes user preferences
func (s *PreferencesService) DeletePreferences(userID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	return s.preferencesRepo.Delete(userID)
}

// Accessibility Settings Methods

// SetTextSize sets the text size multiplier
// Implements Requirement 12.3
func (s *PreferencesService) SetTextSize(userID string, multiplier float64) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if err := prefs.Accessibility.SetTextSize(multiplier); err != nil {
		return nil, err
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetHighContrast enables or disables high contrast mode
// Implements Requirement 12.2
func (s *PreferencesService) SetHighContrast(userID string, enabled bool) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if enabled {
		prefs.Accessibility.EnableHighContrast()
	} else {
		prefs.Accessibility.DisableHighContrast()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetScreenReader enables or disables screen reader support
// Implements Requirement 12.1
func (s *PreferencesService) SetScreenReader(userID string, enabled bool) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if enabled {
		prefs.Accessibility.EnableScreenReader()
	} else {
		prefs.Accessibility.DisableScreenReader()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// SetKeyboardNavigation enables or disables keyboard navigation
// Implements Requirement 12.5
func (s *PreferencesService) SetKeyboardNavigation(userID string, enabled bool) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	if enabled {
		prefs.Accessibility.EnableKeyboardNavigation()
	} else {
		prefs.Accessibility.DisableKeyboardNavigation()
	}

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// EnableAccessibilityMode enables accessibility mode
// Implements Requirement 12.7
func (s *PreferencesService) EnableAccessibilityMode(userID string) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	prefs.Accessibility.EnableAccessibilityMode()

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// DisableAccessibilityMode disables accessibility mode
func (s *PreferencesService) DisableAccessibilityMode(userID string) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	prefs.Accessibility.DisableAccessibilityMode()

	if err := s.preferencesRepo.Save(prefs); err != nil {
		return nil, err
	}

	return prefs, nil
}

// GetAccessibilitySettings returns accessibility settings
func (s *PreferencesService) GetAccessibilitySettings(userID string) (*domain.AccessibilitySettings, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	prefs, err := s.preferencesRepo.Get(userID)
	if err != nil {
		defaultPrefs := domain.NewUserPreferences(userID)
		return &defaultPrefs.Accessibility, nil
	}

	return &prefs.Accessibility, nil
}

// ApplyAccessibilitySettings applies accessibility settings to the response
// Returns CSS variables and HTML attributes for accessibility
func (s *PreferencesService) ApplyAccessibilitySettings(userID string) (map[string]string, error) {
	settings, err := s.GetAccessibilitySettings(userID)
	if err != nil {
		return map[string]string{}, nil
	}

	result := make(map[string]string)

	// Text size
	textSize := settings.GetTextSizePercentage()
	result["textSize"] = string(rune(textSize)) + "%"

	// Element size
	elementSize := settings.GetElementSizePercentage()
	result["elementSize"] = string(rune(elementSize)) + "%"

	// High contrast
	if settings.HighContrastMode {
		result["highContrast"] = "true"
	}

	// Screen reader
	if settings.ScreenReaderEnabled {
		result["ariaEnabled"] = "true"
	}

	// Keyboard navigation
	if settings.KeyboardNavigationEnabled {
		result["keyboardNav"] = "true"
	}

	return result, nil
}

// ResetToDefaults resets preferences to default values
func (s *PreferencesService) ResetToDefaults(userID string) (*domain.UserPreferences, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	defaultPrefs := domain.NewUserPreferences(userID)

	if err := s.preferencesRepo.Save(defaultPrefs); err != nil {
		return nil, err
	}

	return defaultPrefs, nil
}