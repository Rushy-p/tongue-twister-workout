package domain

import (
	"time"
)

// ReminderDays represents the days of the week for reminders
type ReminderDays []int // 0=Sunday, 1=Monday, etc.

// ExportFormat represents the format for data export
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// UserPreferences represents all user preferences
type UserPreferences struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	Difficulty      DifficultyLevel  `json:"difficulty"`
	DefaultDuration time.Duration    `json:"default_duration"`
	AudioFeedback   bool             `json:"audio_feedback"`
	VibrationFeedback bool           `json:"vibration_feedback"`
	ReminderEnabled bool             `json:"reminder_enabled"`
	ReminderTime    time.Time        `json:"reminder_time"`
	ReminderDays    ReminderDays     `json:"reminder_days"`
	ExportFormat    ExportFormat     `json:"export_format"`
	Accessibility   AccessibilitySettings `json:"accessibility"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// AccessibilitySettings represents accessibility configuration
type AccessibilitySettings struct {
	ScreenReaderEnabled  bool    `json:"screen_reader_enabled"`
	HighContrastMode     bool    `json:"high_contrast_mode"`
	TextSizeMultiplier   float64 `json:"text_size_multiplier"`
	ElementSizeMultiplier float64 `json:"element_size_multiplier"`
	KeyboardNavigationEnabled bool `json:"keyboard_navigation_enabled"`
	ColorIndependent     bool    `json:"color_independent"`
}

// NewUserPreferences creates default user preferences
func NewUserPreferences(userID string) *UserPreferences {
	now := time.Now()
	defaultDuration := 60 * time.Second
	defaultReminderTime := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	
	return &UserPreferences{
		ID:                generateID(),
		UserID:            userID,
		Difficulty:        DifficultyBeginner,
		DefaultDuration:   defaultDuration,
		AudioFeedback:     true,
		VibrationFeedback: true,
		ReminderEnabled:   false,
		ReminderTime:      defaultReminderTime,
		ReminderDays:      []int{1, 2, 3, 4, 5, 6, 0}, // All days
		ExportFormat:      ExportFormatJSON,
		Accessibility: AccessibilitySettings{
			ScreenReaderEnabled:      false,
			HighContrastMode:         false,
			TextSizeMultiplier:       1.0,
			ElementSizeMultiplier:    1.0,
			KeyboardNavigationEnabled: true,
			ColorIndependent:         true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetDifficulty sets the difficulty level
func (p *UserPreferences) SetDifficulty(level DifficultyLevel) error {
	switch level {
	case DifficultyBeginner, DifficultyIntermediate, DifficultyAdvanced:
		p.Difficulty = level
		p.UpdatedAt = time.Now()
		return nil
	default:
		return &ValidationError{"invalid difficulty level"}
	}
}

// SetDefaultDuration sets the default exercise duration
func (p *UserPreferences) SetDefaultDuration(duration time.Duration) error {
	validDurations := []time.Duration{
		30 * time.Second,
		60 * time.Second,
		90 * time.Second,
		120 * time.Second,
	}
	
	for _, d := range validDurations {
		if duration == d {
			p.DefaultDuration = duration
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return &ValidationError{"invalid duration. Must be 30, 60, 90, or 120 seconds"}
}

// SetReminderTime sets the reminder time
func (p *UserPreferences) SetReminderTime(hour, minute int) error {
	if hour < 6 || hour > 22 {
		return &ValidationError{"reminder time must be between 6:00 AM and 10:00 PM"}
	}
	
	p.ReminderTime = time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
	p.UpdatedAt = time.Now()
	return nil
}

// SetReminderDays sets the reminder days
func (p *UserPreferences) SetReminderDays(days []int) error {
	for _, d := range days {
		if d < 0 || d > 6 {
			return &ValidationError{"reminder days must be 0-6 (Sunday-Saturday)"}
		}
	}
	p.ReminderDays = days
	p.UpdatedAt = time.Now()
	return nil
}

// EnableReminders enables reminders
func (p *UserPreferences) EnableReminders() {
	p.ReminderEnabled = true
	p.UpdatedAt = time.Now()
}

// DisableReminders disables reminders
func (p *UserPreferences) DisableReminders() {
	p.ReminderEnabled = false
	p.UpdatedAt = time.Now()
}

// EnableAudioFeedback enables audio feedback
func (p *UserPreferences) EnableAudioFeedback() {
	p.AudioFeedback = true
	p.UpdatedAt = time.Now()
}

// DisableAudioFeedback disables audio feedback
func (p *UserPreferences) DisableAudioFeedback() {
	p.AudioFeedback = false
	p.UpdatedAt = time.Now()
}

// EnableVibrationFeedback enables vibration feedback
func (p *UserPreferences) EnableVibrationFeedback() {
	p.VibrationFeedback = true
	p.UpdatedAt = time.Now()
}

// DisableVibrationFeedback disables vibration feedback
func (p *UserPreferences) DisableVibrationFeedback() {
	p.VibrationFeedback = false
	p.UpdatedAt = time.Now()
}

// SetExportFormat sets the export format
func (p *UserPreferences) SetExportFormat(format ExportFormat) error {
	switch format {
	case ExportFormatJSON, ExportFormatCSV:
		p.ExportFormat = format
		p.UpdatedAt = time.Now()
		return nil
	default:
		return &ValidationError{"invalid export format"}
	}
}

// EnableAccessibilityMode enables accessibility mode
func (a *AccessibilitySettings) EnableAccessibilityMode() {
	a.TextSizeMultiplier = 1.25
	a.ElementSizeMultiplier = 1.25
}

// DisableAccessibilityMode disables accessibility mode
func (a *AccessibilitySettings) DisableAccessibilityMode() {
	a.TextSizeMultiplier = 1.0
	a.ElementSizeMultiplier = 1.0
}

// SetTextSize sets the text size multiplier (0.5 to 2.0)
func (a *AccessibilitySettings) SetTextSize(multiplier float64) error {
	if multiplier < 0.5 || multiplier > 2.0 {
		return &ValidationError{"text size multiplier must be between 0.5 and 2.0"}
	}
	a.TextSizeMultiplier = multiplier
	return nil
}

// EnableHighContrast enables high contrast mode
func (a *AccessibilitySettings) EnableHighContrast() {
	a.HighContrastMode = true
}

// DisableHighContrast disables high contrast mode
func (a *AccessibilitySettings) DisableHighContrast() {
	a.HighContrastMode = false
}

// EnableScreenReader enables screen reader support
func (a *AccessibilitySettings) EnableScreenReader() {
	a.ScreenReaderEnabled = true
}

// DisableScreenReader disables screen reader support
func (a *AccessibilitySettings) DisableScreenReader() {
	a.ScreenReaderEnabled = false
}

// EnableKeyboardNavigation enables keyboard navigation
func (a *AccessibilitySettings) EnableKeyboardNavigation() {
	a.KeyboardNavigationEnabled = true
}

// DisableKeyboardNavigation disables keyboard navigation
func (a *AccessibilitySettings) DisableKeyboardNavigation() {
	a.KeyboardNavigationEnabled = false
}

// Validate checks if the preferences are valid
func (p *UserPreferences) Validate() error {
	if p.UserID == "" {
		return &ValidationError{"user ID cannot be empty"}
	}
	
	if p.Difficulty == "" {
		return &ValidationError{"difficulty cannot be empty"}
	}
	
	if p.DefaultDuration == 0 {
		return &ValidationError{"default duration cannot be zero"}
	}
	
	return nil
}

// GetReminderTimeHour returns the reminder hour
func (p *UserPreferences) GetReminderTimeHour() int {
	return p.ReminderTime.Hour()
}

// GetReminderTimeMinute returns the reminder minute
func (p *UserPreferences) GetReminderTimeMinute() int {
	return p.ReminderTime.Minute()
}

// IsReminderDay checks if a day is a reminder day
func (p *UserPreferences) IsReminderDay(day int) bool {
	for _, d := range p.ReminderDays {
		if d == day {
			return true
		}
	}
	return false
}

// GetTextSizePercentage returns the text size as a percentage
func (a *AccessibilitySettings) GetTextSizePercentage() int {
	return int(a.TextSizeMultiplier * 100)
}

// GetElementSizePercentage returns the element size as a percentage
func (a *AccessibilitySettings) GetElementSizePercentage() int {
	return int(a.ElementSizeMultiplier * 100)
}

// IsWeekdaysOnly returns true if reminders are set for weekdays only
func (p *UserPreferences) IsWeekdaysOnly() bool {
	return len(p.ReminderDays) == 5 &&
		p.IsReminderDay(1) && p.IsReminderDay(2) && p.IsReminderDay(3) &&
		p.IsReminderDay(4) && p.IsReminderDay(5) && !p.IsReminderDay(0) && !p.IsReminderDay(6)
}

// IsWeekendsOnly returns true if reminders are set for weekends only
func (p *UserPreferences) IsWeekendsOnly() bool {
	return len(p.ReminderDays) == 2 &&
		!p.IsReminderDay(1) && !p.IsReminderDay(2) && !p.IsReminderDay(3) &&
		!p.IsReminderDay(4) && !p.IsReminderDay(5) && p.IsReminderDay(0) && p.IsReminderDay(6)
}

// IsAllDays returns true if reminders are set for all days
func (p *UserPreferences) IsAllDays() bool {
	return len(p.ReminderDays) == 7
}