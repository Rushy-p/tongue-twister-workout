package service

import (
	"testing"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

func newTestExerciseService(t *testing.T) *ExerciseService {
	t.Helper()
	exerciseRepo := infrastructure.NewInMemoryExerciseRepository()
	progressRepo := infrastructure.NewInMemoryProgressRepository()
	sessionRepo := infrastructure.NewInMemorySessionRepository()
	return NewExerciseService(exerciseRepo, progressRepo, sessionRepo)
}

// TestGetExercisesByFilter_NoFilter verifies that an empty filter returns all exercises.
func TestGetExercisesByFilter_NoFilter(t *testing.T) {
	svc := newTestExerciseService(t)
	all, err := svc.GetAllExercises()
	if err != nil {
		t.Fatalf("GetAllExercises: %v", err)
	}

	filtered, err := svc.GetExercisesByFilter(ExerciseFilter{})
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	if len(filtered) != len(all) {
		t.Errorf("expected %d exercises, got %d", len(all), len(filtered))
	}
}

// TestGetExercisesByFilter_ByCategory verifies filtering by category returns only matching exercises.
func TestGetExercisesByFilter_ByCategory(t *testing.T) {
	svc := newTestExerciseService(t)
	cat := domain.CategoryTongueTwister
	filter := ExerciseFilter{Category: &cat}

	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one tongue twister exercise")
	}
	for _, e := range results {
		if e.Category != domain.CategoryTongueTwister {
			t.Errorf("expected category %s, got %s", domain.CategoryTongueTwister, e.Category)
		}
	}
}

// TestGetExercisesByFilter_ByDifficulty verifies filtering by difficulty returns only matching exercises.
func TestGetExercisesByFilter_ByDifficulty(t *testing.T) {
	svc := newTestExerciseService(t)
	diff := domain.DifficultyBeginner
	filter := ExerciseFilter{Difficulty: &diff}

	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one beginner exercise")
	}
	for _, e := range results {
		if e.Difficulty != domain.DifficultyBeginner {
			t.Errorf("expected difficulty %s, got %s", domain.DifficultyBeginner, e.Difficulty)
		}
	}
}

// TestGetExercisesByFilter_ByTargetSound verifies filtering by target sound.
func TestGetExercisesByFilter_ByTargetSound(t *testing.T) {
	svc := newTestExerciseService(t)
	sound := domain.SoundS
	filter := ExerciseFilter{TargetSound: &sound}

	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one exercise targeting sound S")
	}
	for _, e := range results {
		found := false
		for _, s := range e.TargetSounds {
			if s == domain.SoundS {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("exercise %s does not target sound S", e.ID)
		}
	}
}

// TestGetExercisesByFilter_CompletedOnly verifies that CompletedOnly excludes exercises with zero completions.
func TestGetExercisesByFilter_CompletedOnly(t *testing.T) {
	svc := newTestExerciseService(t)

	// Mark one exercise as completed
	err := svc.MarkExerciseCompleted("user-1", "me-001")
	if err != nil {
		t.Fatalf("MarkExerciseCompleted: %v", err)
	}

	filter := ExerciseFilter{CompletedOnly: true}
	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	for _, e := range results {
		if e.CompletionCount == 0 {
			t.Errorf("CompletedOnly filter returned exercise %s with zero completions", e.ID)
		}
	}
}

// TestGetExercisesByFilter_IncompleteOnly verifies that IncompleteOnly excludes completed exercises.
func TestGetExercisesByFilter_IncompleteOnly(t *testing.T) {
	svc := newTestExerciseService(t)

	// Mark one exercise as completed
	err := svc.MarkExerciseCompleted("user-1", "me-001")
	if err != nil {
		t.Fatalf("MarkExerciseCompleted: %v", err)
	}

	filter := ExerciseFilter{IncompleteOnly: true}
	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	for _, e := range results {
		if e.CompletionCount > 0 {
			t.Errorf("IncompleteOnly filter returned completed exercise %s", e.ID)
		}
	}
}

// TestGetExercisesByFilter_CategoryAndDifficulty verifies combined category + difficulty filtering.
func TestGetExercisesByFilter_CategoryAndDifficulty(t *testing.T) {
	svc := newTestExerciseService(t)
	cat := domain.CategoryMouthExercise
	diff := domain.DifficultyBeginner
	filter := ExerciseFilter{Category: &cat, Difficulty: &diff}

	results, err := svc.GetExercisesByFilter(filter)
	if err != nil {
		t.Fatalf("GetExercisesByFilter: %v", err)
	}

	for _, e := range results {
		if e.Category != domain.CategoryMouthExercise {
			t.Errorf("expected category mouth_exercise, got %s", e.Category)
		}
		if e.Difficulty != domain.DifficultyBeginner {
			t.Errorf("expected difficulty beginner, got %s", e.Difficulty)
		}
	}
}

// TestSortByDifficulty verifies exercises are sorted beginner → intermediate → advanced.
func TestSortByDifficulty(t *testing.T) {
	svc := newTestExerciseService(t)
	exercises := []domain.Exercise{
		{ID: "a", Difficulty: domain.DifficultyAdvanced},
		{ID: "b", Difficulty: domain.DifficultyBeginner},
		{ID: "c", Difficulty: domain.DifficultyIntermediate},
	}

	svc.SortByDifficulty(exercises)

	order := []domain.DifficultyLevel{domain.DifficultyBeginner, domain.DifficultyIntermediate, domain.DifficultyAdvanced}
	for i, e := range exercises {
		if e.Difficulty != order[i] {
			t.Errorf("position %d: expected %s, got %s", i, order[i], e.Difficulty)
		}
	}
}
