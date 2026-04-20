package domain

import (
	"testing"
	"time"
)

// TestUpdateStreak_SameDay verifies that calling UpdateStreak on the same day does not change the streak.
func TestUpdateStreak_SameDay(t *testing.T) {
	streak := NewStreakRecord("user-1")
	streak.CurrentStreak = 3
	streak.LongestStreak = 5
	streak.LastActivityDate = time.Now()

	streak.UpdateStreak()

	if streak.CurrentStreak != 3 {
		t.Errorf("same-day: expected streak 3, got %d", streak.CurrentStreak)
	}
}

// TestUpdateStreak_ConsecutiveDay verifies that practicing on the next day increments the streak.
func TestUpdateStreak_ConsecutiveDay(t *testing.T) {
	streak := NewStreakRecord("user-1")
	streak.CurrentStreak = 3
	streak.LongestStreak = 5
	streak.LastActivityDate = time.Now().AddDate(0, 0, -1)

	streak.UpdateStreak()

	if streak.CurrentStreak != 4 {
		t.Errorf("consecutive day: expected streak 4, got %d", streak.CurrentStreak)
	}
	if streak.LongestStreak != 5 {
		t.Errorf("longest streak should remain 5, got %d", streak.LongestStreak)
	}
}

// TestUpdateStreak_ConsecutiveDay_UpdatesLongest verifies that the longest streak is updated when current exceeds it.
func TestUpdateStreak_ConsecutiveDay_UpdatesLongest(t *testing.T) {
	streak := NewStreakRecord("user-1")
	streak.CurrentStreak = 5
	streak.LongestStreak = 5
	streak.LastActivityDate = time.Now().AddDate(0, 0, -1)

	streak.UpdateStreak()

	if streak.CurrentStreak != 6 {
		t.Errorf("expected streak 6, got %d", streak.CurrentStreak)
	}
	if streak.LongestStreak != 6 {
		t.Errorf("expected longest streak 6, got %d", streak.LongestStreak)
	}
}

// TestUpdateStreak_BrokenStreak verifies that missing a day resets the streak to 1.
func TestUpdateStreak_BrokenStreak(t *testing.T) {
	streak := NewStreakRecord("user-1")
	streak.CurrentStreak = 10
	streak.LongestStreak = 10
	streak.LastActivityDate = time.Now().AddDate(0, 0, -3)

	streak.UpdateStreak()

	if streak.CurrentStreak != 1 {
		t.Errorf("broken streak: expected streak 1, got %d", streak.CurrentStreak)
	}
	// Longest streak should be preserved
	if streak.LongestStreak != 10 {
		t.Errorf("longest streak should remain 10, got %d", streak.LongestStreak)
	}
}

// TestBreakStreak verifies that BreakStreak resets the current streak to 0.
func TestBreakStreak(t *testing.T) {
	streak := NewStreakRecord("user-1")
	streak.CurrentStreak = 7

	streak.BreakStreak()

	if streak.CurrentStreak != 0 {
		t.Errorf("expected streak 0 after break, got %d", streak.CurrentStreak)
	}
}

// TestNewStreakRecord verifies initial values.
func TestNewStreakRecord(t *testing.T) {
	streak := NewStreakRecord("user-42")
	if streak.UserID != "user-42" {
		t.Errorf("expected user-42, got %s", streak.UserID)
	}
	if streak.CurrentStreak != 0 {
		t.Errorf("expected initial streak 0, got %d", streak.CurrentStreak)
	}
	if streak.LongestStreak != 0 {
		t.Errorf("expected initial longest streak 0, got %d", streak.LongestStreak)
	}
}

// TestAchievementUpdateProgress_Unlock verifies that reaching the target unlocks the achievement.
func TestAchievementUpdateProgress_Unlock(t *testing.T) {
	a := NewAchievement("user-1", "First Steps", "Complete first exercise", "🎯", "Complete 1 exercise", AchievementCompletion, 1)
	if a.IsUnlocked() {
		t.Error("expected achievement to be locked initially")
	}

	a.UpdateProgress(1)

	if !a.IsUnlocked() {
		t.Error("expected achievement to be unlocked after reaching target")
	}
}

// TestAchievementGetCompletionPercentage verifies percentage calculation.
func TestAchievementGetCompletionPercentage(t *testing.T) {
	a := NewAchievement("user-1", "Week Warrior", "7-day streak", "🔥", "7 days", AchievementStreak, 7)
	a.Progress = 3

	pct := a.GetCompletionPercentage()
	expected := 3.0 / 7.0 * 100
	if pct != expected {
		t.Errorf("expected %.2f%%, got %.2f%%", expected, pct)
	}
}

// TestCategoryProgressAddExercise verifies that adding an exercise updates counts and average.
func TestCategoryProgressAddExercise(t *testing.T) {
	cp := NewCategoryProgress("user-1", CategoryTongueTwister, 10)
	cp.AddExercise(30 * time.Second)
	cp.AddExercise(60 * time.Second)

	if cp.CompletedExercises != 2 {
		t.Errorf("expected 2 completed exercises, got %d", cp.CompletedExercises)
	}
	if cp.TotalTime != 90*time.Second {
		t.Errorf("expected 90s total time, got %v", cp.TotalTime)
	}
	expectedAvg := 45 * time.Second
	if cp.AverageSessionLength != expectedAvg {
		t.Errorf("expected avg %v, got %v", expectedAvg, cp.AverageSessionLength)
	}
}

// TestCategoryProgressGetCompletionPercentage verifies percentage calculation.
func TestCategoryProgressGetCompletionPercentage(t *testing.T) {
	cp := NewCategoryProgress("user-1", CategoryMouthExercise, 10)
	cp.CompletedExercises = 5

	pct := cp.GetCompletionPercentage()
	if pct != 50.0 {
		t.Errorf("expected 50%%, got %.2f%%", pct)
	}
}

// TestCategoryProgressGetCompletionPercentage_ZeroTotal verifies no division by zero.
func TestCategoryProgressGetCompletionPercentage_ZeroTotal(t *testing.T) {
	cp := NewCategoryProgress("user-1", CategoryMouthExercise, 0)
	pct := cp.GetCompletionPercentage()
	if pct != 0 {
		t.Errorf("expected 0%% for zero total, got %.2f%%", pct)
	}
}
