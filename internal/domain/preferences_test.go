package domain

import (
	"testing"
	"time"
)

// TestSetDefaultDuration_ValidDurations verifies all allowed durations are accepted.
func TestSetDefaultDuration_ValidDurations(t *testing.T) {
	validDurations := []time.Duration{
		30 * time.Second,
		60 * time.Second,
		90 * time.Second,
		120 * time.Second,
	}

	for _, d := range validDurations {
		prefs := NewUserPreferences("user-1")
		if err := prefs.SetDefaultDuration(d); err != nil {
			t.Errorf("expected no error for duration %v, got: %v", d, err)
		}
		if prefs.DefaultDuration != d {
			t.Errorf("expected duration %v, got %v", d, prefs.DefaultDuration)
		}
	}
}

// TestSetDefaultDuration_InvalidDuration verifies that an unsupported duration is rejected.
func TestSetDefaultDuration_InvalidDuration(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.SetDefaultDuration(45 * time.Second); err == nil {
		t.Error("expected error for invalid duration 45s, got nil")
	}
}

// TestSetReminderTime_ValidHours verifies boundary hours are accepted.
func TestSetReminderTime_ValidHours(t *testing.T) {
	cases := []struct{ hour, minute int }{
		{6, 0},
		{9, 30},
		{22, 0},
	}
	for _, c := range cases {
		prefs := NewUserPreferences("user-1")
		if err := prefs.SetReminderTime(c.hour, c.minute); err != nil {
			t.Errorf("expected no error for %02d:%02d, got: %v", c.hour, c.minute, err)
		}
		if prefs.GetReminderTimeHour() != c.hour {
			t.Errorf("expected hour %d, got %d", c.hour, prefs.GetReminderTimeHour())
		}
		if prefs.GetReminderTimeMinute() != c.minute {
			t.Errorf("expected minute %d, got %d", c.minute, prefs.GetReminderTimeMinute())
		}
	}
}

// TestSetReminderTime_TooEarly verifies that hours before 6 are rejected.
func TestSetReminderTime_TooEarly(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.SetReminderTime(5, 0); err == nil {
		t.Error("expected error for hour 5, got nil")
	}
}

// TestSetReminderTime_TooLate verifies that hours after 22 are rejected.
func TestSetReminderTime_TooLate(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.SetReminderTime(23, 0); err == nil {
		t.Error("expected error for hour 23, got nil")
	}
}

// TestSetDifficulty_ValidLevels verifies all difficulty levels are accepted.
func TestSetDifficulty_ValidLevels(t *testing.T) {
	levels := []DifficultyLevel{DifficultyBeginner, DifficultyIntermediate, DifficultyAdvanced}
	for _, l := range levels {
		prefs := NewUserPreferences("user-1")
		if err := prefs.SetDifficulty(l); err != nil {
			t.Errorf("expected no error for difficulty %s, got: %v", l, err)
		}
		if prefs.Difficulty != l {
			t.Errorf("expected difficulty %s, got %s", l, prefs.Difficulty)
		}
	}
}

// TestSetDifficulty_InvalidLevel verifies that an unknown difficulty is rejected.
func TestSetDifficulty_InvalidLevel(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.SetDifficulty("expert"); err == nil {
		t.Error("expected error for invalid difficulty 'expert', got nil")
	}
}

// TestNewUserPreferences_Defaults verifies default values are set correctly.
func TestNewUserPreferences_Defaults(t *testing.T) {
	prefs := NewUserPreferences("user-1")

	if prefs.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", prefs.UserID)
	}
	if prefs.Difficulty != DifficultyBeginner {
		t.Errorf("expected beginner difficulty, got %s", prefs.Difficulty)
	}
	if prefs.DefaultDuration != 60*time.Second {
		t.Errorf("expected 60s default duration, got %v", prefs.DefaultDuration)
	}
	if prefs.ReminderEnabled {
		t.Error("reminders should be disabled by default")
	}
	if !prefs.AudioFeedback {
		t.Error("audio feedback should be enabled by default")
	}
}

// TestUserPreferences_Validate_Valid verifies a valid preferences object passes.
func TestUserPreferences_Validate_Valid(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestUserPreferences_Validate_MissingUserID verifies that missing user ID fails validation.
func TestUserPreferences_Validate_MissingUserID(t *testing.T) {
	prefs := NewUserPreferences("")
	prefs.UserID = ""
	if err := prefs.Validate(); err == nil {
		t.Error("expected error for missing user ID")
	}
}

// TestSetReminderDays_ValidDays verifies valid day values are accepted.
func TestSetReminderDays_ValidDays(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	days := []int{1, 2, 3, 4, 5}
	if err := prefs.SetReminderDays(days); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestSetReminderDays_InvalidDay verifies that out-of-range day values are rejected.
func TestSetReminderDays_InvalidDay(t *testing.T) {
	prefs := NewUserPreferences("user-1")
	if err := prefs.SetReminderDays([]int{0, 7}); err == nil {
		t.Error("expected error for day value 7, got nil")
	}
}
