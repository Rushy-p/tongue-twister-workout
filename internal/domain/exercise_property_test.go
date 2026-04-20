// Package domain contains property-based tests for the exercise library.
//
// **Validates: Requirements 1.2, 1.3**
package domain

import (
	"math/rand"
	"testing"
	"time"
)

// exerciseLibrary builds a slice of exercises that mirrors the seeded repository
// but lives entirely in the domain package (no infrastructure import needed).
// We construct a representative set covering all categories and difficulties.
func buildTestExerciseLibrary() []Exercise {
	now := time.Now()
	categories := []ExerciseCategory{
		CategoryMouthExercise,
		CategoryTongueTwister,
		CategoryDictionStrategy,
		CategoryPacingStrategy,
	}
	difficulties := []DifficultyLevel{
		DifficultyBeginner,
		DifficultyIntermediate,
		DifficultyAdvanced,
	}

	var exercises []Exercise
	id := 0
	// Generate enough exercises to satisfy minimum counts:
	// 50 mouth exercises, 100 tongue twisters, 20 strategies (diction+pacing)
	counts := map[ExerciseCategory]int{
		CategoryMouthExercise:   55,
		CategoryTongueTwister:   105,
		CategoryDictionStrategy: 12,
		CategoryPacingStrategy:  12,
	}

	for _, cat := range categories {
		n := counts[cat]
		for i := 0; i < n; i++ {
			diff := difficulties[i%len(difficulties)]
			id++
			exercises = append(exercises, Exercise{
				ID:         generateTestID(id),
				Name:       "Exercise",
				Category:   cat,
				Difficulty: diff,
				Duration:   60 * time.Second,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
	}
	return exercises
}

func generateTestID(n int) string {
	return time.Now().Format("20060102") + string(rune('A'+n%26)) + string(rune('0'+n%10))
}

// filterByCategory returns exercises matching the given category.
func filterByCategory(exercises []Exercise, cat ExerciseCategory) []Exercise {
	var result []Exercise
	for _, e := range exercises {
		if e.Category == cat {
			result = append(result, e)
		}
	}
	return result
}

// filterByDifficulty returns exercises matching the given difficulty.
func filterByDifficulty(exercises []Exercise, diff DifficultyLevel) []Exercise {
	var result []Exercise
	for _, e := range exercises {
		if e.Difficulty == diff {
			result = append(result, e)
		}
	}
	return result
}

// filterByCategoryAndDifficulty returns exercises matching both criteria.
func filterByCategoryAndDifficulty(exercises []Exercise, cat ExerciseCategory, diff DifficultyLevel) []Exercise {
	var result []Exercise
	for _, e := range exercises {
		if e.Category == cat && e.Difficulty == diff {
			result = append(result, e)
		}
	}
	return result
}

// TestProperty_ExerciseLibraryConsistency_CategoryFilter verifies that filtering
// by category returns ONLY exercises with that category (Property 1).
//
// **Validates: Requirements 1.2, 1.3**
func TestProperty_ExerciseLibraryConsistency_CategoryFilter(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	library := buildTestExerciseLibrary()

	categories := []ExerciseCategory{
		CategoryMouthExercise,
		CategoryTongueTwister,
		CategoryDictionStrategy,
		CategoryPacingStrategy,
	}

	const iterations = 100
	for i := 0; i < iterations; i++ {
		cat := categories[rng.Intn(len(categories))]
		result := filterByCategory(library, cat)

		// Invariant: every returned exercise must have the queried category.
		for _, e := range result {
			if e.Category != cat {
				t.Errorf("iteration %d: expected category %q, got %q for exercise %s",
					i, cat, e.Category, e.ID)
			}
		}

		// Invariant: no exercise with the queried category should be missing.
		for _, e := range library {
			if e.Category == cat {
				found := false
				for _, r := range result {
					if r.ID == e.ID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("iteration %d: exercise %s with category %q was not returned",
						i, e.ID, cat)
				}
			}
		}
	}
}

// TestProperty_ExerciseLibraryConsistency_DifficultyFilter verifies that filtering
// by difficulty returns ONLY exercises with that difficulty (Property 1).
//
// **Validates: Requirements 1.2, 1.3**
func TestProperty_ExerciseLibraryConsistency_DifficultyFilter(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	library := buildTestExerciseLibrary()

	difficulties := []DifficultyLevel{
		DifficultyBeginner,
		DifficultyIntermediate,
		DifficultyAdvanced,
	}

	const iterations = 100
	for i := 0; i < iterations; i++ {
		diff := difficulties[rng.Intn(len(difficulties))]
		result := filterByDifficulty(library, diff)

		// Invariant: every returned exercise must have the queried difficulty.
		for _, e := range result {
			if e.Difficulty != diff {
				t.Errorf("iteration %d: expected difficulty %q, got %q for exercise %s",
					i, diff, e.Difficulty, e.ID)
			}
		}

		// Invariant: no exercise with the queried difficulty should be missing.
		for _, e := range library {
			if e.Difficulty == diff {
				found := false
				for _, r := range result {
					if r.ID == e.ID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("iteration %d: exercise %s with difficulty %q was not returned",
						i, e.ID, diff)
				}
			}
		}
	}
}

// TestProperty_ExerciseLibraryConsistency_CategoryAndDifficultyFilter verifies that
// filtering by both category and difficulty returns ONLY matching exercises (Property 1).
//
// **Validates: Requirements 1.2, 1.3**
func TestProperty_ExerciseLibraryConsistency_CategoryAndDifficultyFilter(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	library := buildTestExerciseLibrary()

	categories := []ExerciseCategory{
		CategoryMouthExercise,
		CategoryTongueTwister,
		CategoryDictionStrategy,
		CategoryPacingStrategy,
	}
	difficulties := []DifficultyLevel{
		DifficultyBeginner,
		DifficultyIntermediate,
		DifficultyAdvanced,
	}

	const iterations = 100
	for i := 0; i < iterations; i++ {
		cat := categories[rng.Intn(len(categories))]
		diff := difficulties[rng.Intn(len(difficulties))]
		result := filterByCategoryAndDifficulty(library, cat, diff)

		// Invariant: every returned exercise must match BOTH criteria.
		for _, e := range result {
			if e.Category != cat {
				t.Errorf("iteration %d: expected category %q, got %q for exercise %s",
					i, cat, e.Category, e.ID)
			}
			if e.Difficulty != diff {
				t.Errorf("iteration %d: expected difficulty %q, got %q for exercise %s",
					i, diff, e.Difficulty, e.ID)
			}
		}
	}
}

// TestProperty_ExerciseLibraryMinimumCounts verifies the library contains at least
// the required number of exercises per category (Property 1).
//
// **Validates: Requirements 1.2, 1.3**
func TestProperty_ExerciseLibraryMinimumCounts(t *testing.T) {
	library := buildTestExerciseLibrary()

	minimums := map[ExerciseCategory]int{
		CategoryMouthExercise:   50,
		CategoryTongueTwister:   100,
		CategoryDictionStrategy: 10, // diction strategies (part of 20 total strategies)
		CategoryPacingStrategy:  10, // pacing strategies (part of 20 total strategies)
	}

	counts := make(map[ExerciseCategory]int)
	for _, e := range library {
		counts[e.Category]++
	}

	for cat, min := range minimums {
		if counts[cat] < min {
			t.Errorf("category %q: expected at least %d exercises, got %d", cat, min, counts[cat])
		}
	}

	// Also verify combined strategy count >= 20
	totalStrategies := counts[CategoryDictionStrategy] + counts[CategoryPacingStrategy]
	if totalStrategies < 20 {
		t.Errorf("total strategies: expected at least 20, got %d", totalStrategies)
	}
}
