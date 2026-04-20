package domain

import (
	"testing"
	"time"
)

// TestExerciseValidate_ValidExercise verifies a fully populated exercise passes validation.
func TestExerciseValidate_ValidExercise(t *testing.T) {
	e := Exercise{
		ID:         "ex-001",
		Name:       "Lip Rounding",
		Category:   CategoryMouthExercise,
		Difficulty: DifficultyBeginner,
		Duration:   60 * time.Second,
		CreatedAt:  time.Now(),
	}
	if err := e.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestExerciseValidate_MissingID verifies that an empty ID returns an error.
func TestExerciseValidate_MissingID(t *testing.T) {
	e := Exercise{
		Name:       "Lip Rounding",
		Category:   CategoryMouthExercise,
		Difficulty: DifficultyBeginner,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing ID, got nil")
	}
}

// TestExerciseValidate_MissingName verifies that an empty name returns an error.
func TestExerciseValidate_MissingName(t *testing.T) {
	e := Exercise{
		ID:         "ex-001",
		Category:   CategoryMouthExercise,
		Difficulty: DifficultyBeginner,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// TestExerciseValidate_MissingCategory verifies that an empty category returns an error.
func TestExerciseValidate_MissingCategory(t *testing.T) {
	e := Exercise{
		ID:         "ex-001",
		Name:       "Lip Rounding",
		Difficulty: DifficultyBeginner,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing category, got nil")
	}
}

// TestExerciseValidate_MissingDifficulty verifies that an empty difficulty returns an error.
func TestExerciseValidate_MissingDifficulty(t *testing.T) {
	e := Exercise{
		ID:       "ex-001",
		Name:     "Lip Rounding",
		Category: CategoryMouthExercise,
	}
	if err := e.Validate(); err == nil {
		t.Error("expected error for missing difficulty, got nil")
	}
}

// TestExerciseDifficultyLevels verifies all three difficulty constants are distinct.
func TestExerciseDifficultyLevels(t *testing.T) {
	levels := []DifficultyLevel{DifficultyBeginner, DifficultyIntermediate, DifficultyAdvanced}
	seen := make(map[DifficultyLevel]bool)
	for _, l := range levels {
		if seen[l] {
			t.Errorf("duplicate difficulty level: %s", l)
		}
		seen[l] = true
	}
}

// TestExerciseIsCompleted verifies completion status based on CompletionCount.
func TestExerciseIsCompleted(t *testing.T) {
	e := Exercise{CompletionCount: 0}
	if e.IsCompleted() {
		t.Error("expected not completed when count is 0")
	}
	e.CompletionCount = 1
	if !e.IsCompleted() {
		t.Error("expected completed when count > 0")
	}
}

// TestExerciseGetDurationSeconds verifies duration conversion.
func TestExerciseGetDurationSeconds(t *testing.T) {
	e := Exercise{Duration: 90 * time.Second}
	if got := e.GetDurationSeconds(); got != 90 {
		t.Errorf("expected 90, got %d", got)
	}
}

// TestExerciseCategories verifies all four category constants are distinct.
func TestExerciseCategories(t *testing.T) {
	cats := []ExerciseCategory{
		CategoryMouthExercise,
		CategoryTongueTwister,
		CategoryDictionStrategy,
		CategoryPacingStrategy,
	}
	seen := make(map[ExerciseCategory]bool)
	for _, c := range cats {
		if seen[c] {
			t.Errorf("duplicate category: %s", c)
		}
		seen[c] = true
	}
}
