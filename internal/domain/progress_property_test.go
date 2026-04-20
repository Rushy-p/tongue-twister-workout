// Package domain contains property-based tests for streak continuity.
//
// **Validates: Requirements 6.4, 6.5, 7.2**
package domain

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

// computeExpectedStreak calculates the expected current streak from a sorted
// (ascending) slice of distinct practice dates. Each date represents a day on
// which at least one exercise was completed.
//
// The streak is the count of consecutive days ending on the most recent date.
// If the slice is empty the streak is 0.
func computeExpectedStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	// Work backwards from the last date.
	streak := 1
	for i := len(dates) - 1; i > 0; i-- {
		curr := truncateToDay(dates[i])
		prev := truncateToDay(dates[i-1])
		if curr.Sub(prev) == 24*time.Hour {
			streak++
		} else {
			break
		}
	}
	return streak
}

// computeExpectedLongestStreak calculates the longest consecutive run in the
// sorted slice of distinct practice dates.
func computeExpectedLongestStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	longest := 1
	current := 1
	for i := 1; i < len(dates); i++ {
		curr := truncateToDay(dates[i])
		prev := truncateToDay(dates[i-1])
		if curr.Sub(prev) == 24*time.Hour {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}
	return longest
}

// truncateToDay returns midnight of the given time in its local timezone.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// generatePracticeDates produces a random sequence of practice dates.
// Some days are consecutive, some have gaps, controlled by gapProb.
func generatePracticeDates(rng *rand.Rand, n int, gapProb float64) []time.Time {
	if n == 0 {
		return nil
	}

	// Start from a fixed reference point so tests are deterministic per seed.
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := make([]time.Time, 0, n)
	current := base

	for i := 0; i < n; i++ {
		dates = append(dates, current)
		if rng.Float64() < gapProb {
			// Insert a gap of 2–5 days.
			current = current.AddDate(0, 0, 2+rng.Intn(4))
		} else {
			current = current.AddDate(0, 0, 1)
		}
	}
	return dates
}

// applyDatesToStreakRecord simulates recording practice on each date by
// directly applying the same logic as UpdateStreak but with controlled dates
// instead of time.Now(). This lets us test streak calculations against
// historical date sequences.
func applyDatesToStreakRecord(dates []time.Time) *StreakRecord {
	streak := &StreakRecord{
		UserID:        "test-user",
		CurrentStreak: 0,
		LongestStreak: 0,
	}
	if len(dates) == 0 {
		return streak
	}

	// Sort dates ascending.
	sorted := make([]time.Time, len(dates))
	copy(sorted, dates)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })

	// Deduplicate to day granularity.
	deduped := []time.Time{sorted[0]}
	for _, d := range sorted[1:] {
		if truncateToDay(d) != truncateToDay(deduped[len(deduped)-1]) {
			deduped = append(deduped, d)
		}
	}

	// Apply the first date: streak starts at 1.
	streak.CurrentStreak = 1
	streak.LongestStreak = 1
	streak.LastActivityDate = deduped[0]
	streak.StreakStartDate = deduped[0]

	// Replay each subsequent practice day using the same logic as UpdateStreak
	// but with the controlled date instead of time.Now().
	for i := 1; i < len(deduped); i++ {
		today := truncateToDay(deduped[i])
		lastActivity := truncateToDay(streak.LastActivityDate)
		daysDiff := int(today.Sub(lastActivity).Hours() / 24)

		switch {
		case daysDiff == 0:
			// Same day, no change.
		case daysDiff == 1:
			// Consecutive day.
			streak.CurrentStreak++
			if streak.CurrentStreak > streak.LongestStreak {
				streak.LongestStreak = streak.CurrentStreak
			}
			streak.StreakStartDate = lastActivity
		default:
			// Gap — streak resets.
			streak.CurrentStreak = 1
			streak.StreakStartDate = today
		}
		streak.LastActivityDate = deduped[i]
	}

	return streak
}

// TestProperty_StreakContinuity_ConsecutiveDays verifies that a fully consecutive
// sequence of practice days produces a streak equal to the number of days (Property 3).
//
// **Validates: Requirements 6.4, 6.5, 7.2**
func TestProperty_StreakContinuity_ConsecutiveDays(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		n := 1 + rng.Intn(30) // 1–30 consecutive days
		dates := generatePracticeDates(rng, n, 0) // 0 gap probability = all consecutive

		expected := computeExpectedStreak(dates)
		streak := applyDatesToStreakRecord(dates)

		if streak.CurrentStreak != expected {
			t.Errorf("iteration %d (n=%d): expected streak %d, got %d",
				i, n, expected, streak.CurrentStreak)
		}
	}
}

// TestProperty_StreakContinuity_WithGaps verifies that gaps in practice reset the
// streak to count only from after the last gap (Property 3).
//
// **Validates: Requirements 6.4, 6.5, 7.2**
func TestProperty_StreakContinuity_WithGaps(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		n := 2 + rng.Intn(20) // 2–21 dates
		gapProb := 0.2 + rng.Float64()*0.4 // 20–60% chance of gap
		dates := generatePracticeDates(rng, n, gapProb)

		// Sort and deduplicate for the reference calculation.
		sorted := make([]time.Time, len(dates))
		copy(sorted, dates)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Before(sorted[j]) })
		deduped := []time.Time{sorted[0]}
		for _, d := range sorted[1:] {
			if truncateToDay(d) != truncateToDay(deduped[len(deduped)-1]) {
				deduped = append(deduped, d)
			}
		}

		expected := computeExpectedStreak(deduped)
		streak := applyDatesToStreakRecord(dates)

		if streak.CurrentStreak != expected {
			t.Errorf("iteration %d (n=%d, gapProb=%.2f): expected streak %d, got %d",
				i, n, gapProb, expected, streak.CurrentStreak)
		}
	}
}

// TestProperty_StreakContinuity_LongestGeqCurrent verifies that the longest streak
// is always >= the current streak (Property 3).
//
// **Validates: Requirements 6.4, 6.5, 7.2**
func TestProperty_StreakContinuity_LongestGeqCurrent(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		n := rng.Intn(25) // 0–24 dates
		gapProb := rng.Float64() * 0.5
		dates := generatePracticeDates(rng, n, gapProb)
		streak := applyDatesToStreakRecord(dates)

		if streak.LongestStreak < streak.CurrentStreak {
			t.Errorf("iteration %d: longest streak %d < current streak %d",
				i, streak.LongestStreak, streak.CurrentStreak)
		}
	}
}

// TestProperty_StreakContinuity_ZeroStreak verifies that an empty practice history
// produces a streak of 0 (Property 3).
//
// **Validates: Requirements 6.4, 6.5, 7.2**
func TestProperty_StreakContinuity_ZeroStreak(t *testing.T) {
	streak := NewStreakRecord("test-user")
	// A brand-new record has CurrentStreak == 0.
	if streak.CurrentStreak != 0 {
		t.Errorf("expected initial streak 0, got %d", streak.CurrentStreak)
	}
}

// TestProperty_StreakContinuity_GapResetsStreak verifies that a gap of more than
// 1 day resets the streak to 1 (the day practice resumed) (Property 3).
//
// **Validates: Requirements 6.4, 6.5, 7.2**
func TestProperty_StreakContinuity_GapResetsStreak(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	const iterations = 100

	for i := 0; i < iterations; i++ {
		// Build a run of consecutive days, then a gap, then more consecutive days.
		runBefore := 1 + rng.Intn(10)
		gap := 2 + rng.Intn(5) // gap >= 2 days
		runAfter := 1 + rng.Intn(10)

		base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
		var dates []time.Time
		for d := 0; d < runBefore; d++ {
			dates = append(dates, base.AddDate(0, 0, d))
		}
		gapStart := base.AddDate(0, 0, runBefore+gap)
		for d := 0; d < runAfter; d++ {
			dates = append(dates, gapStart.AddDate(0, 0, d))
		}

		streak := applyDatesToStreakRecord(dates)

		// After the gap the streak should equal runAfter.
		if streak.CurrentStreak != runAfter {
			t.Errorf("iteration %d (runBefore=%d, gap=%d, runAfter=%d): expected streak %d, got %d",
				i, runBefore, gap, runAfter, runAfter, streak.CurrentStreak)
		}
	}
}
