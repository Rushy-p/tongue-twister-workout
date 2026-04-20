// Package domain contains property-based tests for preference persistence.
//
// **Validates: Requirements 9.5, 9.6**
package domain

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"
)

// randomDifficulty picks a random valid difficulty level.
func randomDifficulty(rng *rand.Rand) DifficultyLevel {
	levels := []DifficultyLevel{DifficultyBeginner, DifficultyIntermediate, DifficultyAdvanced}
	return levels[rng.Intn(len(levels))]
}

// randomDuration picks a random valid default duration.
func randomDuration(rng *rand.Rand) time.Duration {
	durations := []time.Duration{30 * time.Second, 60 * time.Second, 90 * time.Second, 120 * time.Second}
	return durations[rng.Intn(len(durations))]
}

// randomReminderHour picks a valid reminder hour (6–22).
func randomReminderHour(rng *rand.Rand) int {
	return 6 + rng.Intn(17) // 6..22
}

// randomReminderMinute picks a valid reminder minute (0–59).
func randomReminderMinute(rng *rand.Rand) int {
	return rng.Intn(60)
}

// randomReminderDays picks a non-empty subset of valid days (1–7 days from 0–6).
func randomReminderDays(rng *rand.Rand) []int {
	n := 1 + rng.Intn(7)
	perm := rng.Perm(7) // shuffle 0..6
	days := make([]int, n)
	for i := 0; i < n; i++ {
		days[i] = perm[i]
	}
	return days
}

// randomTextSizeMultiplier picks a valid text size multiplier (0.5–2.0).
func randomTextSizeMultiplier(rng *rand.Rand) float64 {
	// 0.5 + random in [0, 1.5]
	return 0.5 + rng.Float64()*1.5
}

// serializePreferences marshals preferences to JSON and back, simulating
// a save-to-storage / load-from-storage round-trip.
func serializePreferences(prefs *UserPreferences) (*UserPreferences, error) {
	data, err := json.Marshal(prefs)
	if err != nil {
		return nil, err
	}
	var loaded UserPreferences
	if err := json.Unmarshal(data, &loaded); err != nil {
		return nil, err
	}
	return &loaded, nil
}

// TestProperty_PreferencePersistence_RoundTrip verifies that reading a preference
// immediately after setting it returns the same value (Property 6).
//
// **Validates: Requirements 9.5, 9.6**
func TestProperty_PreferencePersistence_RoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		prefs := NewUserPreferences("user-prop-test")

		// Set random valid values.
		diff := randomDifficulty(rng)
		dur := randomDuration(rng)
		hour := randomReminderHour(rng)
		minute := randomReminderMinute(rng)
		days := randomReminderDays(rng)
		audio := rng.Intn(2) == 1
		vibration := rng.Intn(2) == 1
		reminderEnabled := rng.Intn(2) == 1

		if err := prefs.SetDifficulty(diff); err != nil {
			t.Fatalf("iteration %d: SetDifficulty failed: %v", i, err)
		}
		if err := prefs.SetDefaultDuration(dur); err != nil {
			t.Fatalf("iteration %d: SetDefaultDuration failed: %v", i, err)
		}
		if err := prefs.SetReminderTime(hour, minute); err != nil {
			t.Fatalf("iteration %d: SetReminderTime failed: %v", i, err)
		}
		if err := prefs.SetReminderDays(days); err != nil {
			t.Fatalf("iteration %d: SetReminderDays failed: %v", i, err)
		}
		if audio {
			prefs.EnableAudioFeedback()
		} else {
			prefs.DisableAudioFeedback()
		}
		if vibration {
			prefs.EnableVibrationFeedback()
		} else {
			prefs.DisableVibrationFeedback()
		}
		if reminderEnabled {
			prefs.EnableReminders()
		} else {
			prefs.DisableReminders()
		}

		// Invariant: reading back immediately returns the same values.
		if prefs.Difficulty != diff {
			t.Errorf("iteration %d: difficulty round-trip: set %q, got %q", i, diff, prefs.Difficulty)
		}
		if prefs.DefaultDuration != dur {
			t.Errorf("iteration %d: duration round-trip: set %v, got %v", i, dur, prefs.DefaultDuration)
		}
		if prefs.GetReminderTimeHour() != hour {
			t.Errorf("iteration %d: reminder hour round-trip: set %d, got %d", i, hour, prefs.GetReminderTimeHour())
		}
		if prefs.GetReminderTimeMinute() != minute {
			t.Errorf("iteration %d: reminder minute round-trip: set %d, got %d", i, minute, prefs.GetReminderTimeMinute())
		}
		if prefs.AudioFeedback != audio {
			t.Errorf("iteration %d: audio feedback round-trip: set %v, got %v", i, audio, prefs.AudioFeedback)
		}
		if prefs.VibrationFeedback != vibration {
			t.Errorf("iteration %d: vibration feedback round-trip: set %v, got %v", i, vibration, prefs.VibrationFeedback)
		}
		if prefs.ReminderEnabled != reminderEnabled {
			t.Errorf("iteration %d: reminder enabled round-trip: set %v, got %v", i, reminderEnabled, prefs.ReminderEnabled)
		}
	}
}

// TestProperty_PreferencePersistence_SerializationRoundTrip verifies that
// serializing preferences to storage and deserializing preserves all values
// exactly (Property 6).
//
// **Validates: Requirements 9.5, 9.6**
func TestProperty_PreferencePersistence_SerializationRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		prefs := NewUserPreferences("user-serial-test")

		diff := randomDifficulty(rng)
		dur := randomDuration(rng)
		hour := randomReminderHour(rng)
		minute := randomReminderMinute(rng)
		days := randomReminderDays(rng)
		audio := rng.Intn(2) == 1
		vibration := rng.Intn(2) == 1
		reminderEnabled := rng.Intn(2) == 1
		highContrast := rng.Intn(2) == 1
		screenReader := rng.Intn(2) == 1

		_ = prefs.SetDifficulty(diff)
		_ = prefs.SetDefaultDuration(dur)
		_ = prefs.SetReminderTime(hour, minute)
		_ = prefs.SetReminderDays(days)
		if audio {
			prefs.EnableAudioFeedback()
		} else {
			prefs.DisableAudioFeedback()
		}
		if vibration {
			prefs.EnableVibrationFeedback()
		} else {
			prefs.DisableVibrationFeedback()
		}
		if reminderEnabled {
			prefs.EnableReminders()
		} else {
			prefs.DisableReminders()
		}
		if highContrast {
			prefs.Accessibility.EnableHighContrast()
		} else {
			prefs.Accessibility.DisableHighContrast()
		}
		if screenReader {
			prefs.Accessibility.EnableScreenReader()
		} else {
			prefs.Accessibility.DisableScreenReader()
		}

		// Simulate save + load via JSON serialization.
		loaded, err := serializePreferences(prefs)
		if err != nil {
			t.Fatalf("iteration %d: serialization failed: %v", i, err)
		}

		// Invariant: all values are preserved after the round-trip.
		if loaded.Difficulty != prefs.Difficulty {
			t.Errorf("iteration %d: difficulty: want %q, got %q", i, prefs.Difficulty, loaded.Difficulty)
		}
		if loaded.DefaultDuration != prefs.DefaultDuration {
			t.Errorf("iteration %d: default_duration: want %v, got %v", i, prefs.DefaultDuration, loaded.DefaultDuration)
		}
		if loaded.GetReminderTimeHour() != prefs.GetReminderTimeHour() {
			t.Errorf("iteration %d: reminder_hour: want %d, got %d", i, prefs.GetReminderTimeHour(), loaded.GetReminderTimeHour())
		}
		if loaded.GetReminderTimeMinute() != prefs.GetReminderTimeMinute() {
			t.Errorf("iteration %d: reminder_minute: want %d, got %d", i, prefs.GetReminderTimeMinute(), loaded.GetReminderTimeMinute())
		}
		if loaded.AudioFeedback != prefs.AudioFeedback {
			t.Errorf("iteration %d: audio_feedback: want %v, got %v", i, prefs.AudioFeedback, loaded.AudioFeedback)
		}
		if loaded.VibrationFeedback != prefs.VibrationFeedback {
			t.Errorf("iteration %d: vibration_feedback: want %v, got %v", i, prefs.VibrationFeedback, loaded.VibrationFeedback)
		}
		if loaded.ReminderEnabled != prefs.ReminderEnabled {
			t.Errorf("iteration %d: reminder_enabled: want %v, got %v", i, prefs.ReminderEnabled, loaded.ReminderEnabled)
		}
		if loaded.Accessibility.HighContrastMode != prefs.Accessibility.HighContrastMode {
			t.Errorf("iteration %d: high_contrast: want %v, got %v", i, prefs.Accessibility.HighContrastMode, loaded.Accessibility.HighContrastMode)
		}
		if loaded.Accessibility.ScreenReaderEnabled != prefs.Accessibility.ScreenReaderEnabled {
			t.Errorf("iteration %d: screen_reader: want %v, got %v", i, prefs.Accessibility.ScreenReaderEnabled, loaded.Accessibility.ScreenReaderEnabled)
		}
	}
}

// TestProperty_PreferencePersistence_InvalidValuesRejected verifies that invalid
// preference values are rejected and do not corrupt existing preferences (Property 6).
//
// **Validates: Requirements 9.5, 9.6**
func TestProperty_PreferencePersistence_InvalidValuesRejected(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		prefs := NewUserPreferences("user-invalid-test")

		// Set a known-good baseline.
		baseline := randomDifficulty(rng)
		baselineDur := randomDuration(rng)
		_ = prefs.SetDifficulty(baseline)
		_ = prefs.SetDefaultDuration(baselineDur)

		// Attempt to set invalid difficulty — must be rejected.
		invalidDifficulties := []DifficultyLevel{"expert", "novice", "", "BEGINNER"}
		for _, bad := range invalidDifficulties {
			err := prefs.SetDifficulty(bad)
			if err == nil {
				t.Errorf("iteration %d: expected error for invalid difficulty %q, got nil", i, bad)
			}
			// Invariant: existing value must be unchanged.
			if prefs.Difficulty != baseline {
				t.Errorf("iteration %d: invalid difficulty %q corrupted existing value: want %q, got %q",
					i, bad, baseline, prefs.Difficulty)
			}
		}

		// Attempt to set invalid duration — must be rejected.
		invalidDurations := []time.Duration{0, 45 * time.Second, 150 * time.Second, -1 * time.Second}
		for _, bad := range invalidDurations {
			err := prefs.SetDefaultDuration(bad)
			if err == nil {
				t.Errorf("iteration %d: expected error for invalid duration %v, got nil", i, bad)
			}
			// Invariant: existing value must be unchanged.
			if prefs.DefaultDuration != baselineDur {
				t.Errorf("iteration %d: invalid duration %v corrupted existing value: want %v, got %v",
					i, bad, baselineDur, prefs.DefaultDuration)
			}
		}

		// Attempt to set invalid reminder hour — must be rejected.
		invalidHours := []int{-1, 5, 23, 24, 100}
		for _, bad := range invalidHours {
			origHour := prefs.GetReminderTimeHour()
			err := prefs.SetReminderTime(bad, 0)
			if err == nil {
				t.Errorf("iteration %d: expected error for invalid hour %d, got nil", i, bad)
			}
			// Invariant: existing value must be unchanged.
			if prefs.GetReminderTimeHour() != origHour {
				t.Errorf("iteration %d: invalid hour %d corrupted existing value: want %d, got %d",
					i, bad, origHour, prefs.GetReminderTimeHour())
			}
		}

		// Attempt to set invalid reminder days — must be rejected.
		invalidDaysSets := [][]int{{-1}, {7}, {0, 7}, {8, 9}}
		for _, bad := range invalidDaysSets {
			origDays := make([]int, len(prefs.ReminderDays))
			copy(origDays, prefs.ReminderDays)
			err := prefs.SetReminderDays(bad)
			if err == nil {
				t.Errorf("iteration %d: expected error for invalid days %v, got nil", i, bad)
			}
		}
	}
}

// TestProperty_PreferencePersistence_AccessibilityRoundTrip verifies that
// accessibility settings survive a serialization round-trip (Property 6).
//
// **Validates: Requirements 9.5, 9.6**
func TestProperty_PreferencePersistence_AccessibilityRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		prefs := NewUserPreferences("user-a11y-test")

		textSize := randomTextSizeMultiplier(rng)
		if err := prefs.Accessibility.SetTextSize(textSize); err != nil {
			t.Fatalf("iteration %d: SetTextSize(%v) failed: %v", i, textSize, err)
		}

		loaded, err := serializePreferences(prefs)
		if err != nil {
			t.Fatalf("iteration %d: serialization failed: %v", i, err)
		}

		// Invariant: text size multiplier is preserved.
		if loaded.Accessibility.TextSizeMultiplier != prefs.Accessibility.TextSizeMultiplier {
			t.Errorf("iteration %d: text_size_multiplier: want %v, got %v",
				i, prefs.Accessibility.TextSizeMultiplier, loaded.Accessibility.TextSizeMultiplier)
		}
	}
}
