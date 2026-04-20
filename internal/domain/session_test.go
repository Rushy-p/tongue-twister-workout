package domain

import (
	"testing"
	"time"
)

// TestCalculateSessionStats_Empty verifies stats for a session with no exercises.
func TestCalculateSessionStats_Empty(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	stats := session.CalculateSessionStats()

	if stats.TotalExercises != 0 {
		t.Errorf("expected 0 exercises, got %d", stats.TotalExercises)
	}
	if stats.TotalRepetitions != 0 {
		t.Errorf("expected 0 repetitions, got %d", stats.TotalRepetitions)
	}
	if stats.AverageScore != 0 {
		t.Errorf("expected 0 average score, got %d", stats.AverageScore)
	}
}

// TestCalculateSessionStats_WithExercises verifies stats are correctly aggregated.
func TestCalculateSessionStats_WithExercises(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	session.AddExercise(SessionExercise{
		ID:                   "se-1",
		SessionID:            "s-001",
		ExerciseID:           "ex-1",
		RepetitionsCompleted: 5,
		Duration:             30 * time.Second,
		Score:                80,
	})
	session.AddExercise(SessionExercise{
		ID:                   "se-2",
		SessionID:            "s-001",
		ExerciseID:           "ex-2",
		RepetitionsCompleted: 3,
		Duration:             60 * time.Second,
		Score:                60,
	})

	stats := session.CalculateSessionStats()

	if stats.TotalExercises != 2 {
		t.Errorf("expected 2 exercises, got %d", stats.TotalExercises)
	}
	if stats.TotalRepetitions != 8 {
		t.Errorf("expected 8 repetitions, got %d", stats.TotalRepetitions)
	}
	if stats.TotalScore != 140 {
		t.Errorf("expected total score 140, got %d", stats.TotalScore)
	}
	if stats.AverageScore != 70 {
		t.Errorf("expected average score 70, got %d", stats.AverageScore)
	}
}

// TestCalculateSessionStats_AverageScore_SingleExercise verifies average with one exercise.
func TestCalculateSessionStats_AverageScore_SingleExercise(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	session.AddExercise(SessionExercise{
		ID:         "se-1",
		SessionID:  "s-001",
		ExerciseID: "ex-1",
		Score:      90,
		Duration:   45 * time.Second,
	})

	stats := session.CalculateSessionStats()

	if stats.AverageScore != 90 {
		t.Errorf("expected average score 90, got %d", stats.AverageScore)
	}
}

// TestPracticeSession_StatusTransitions verifies status transitions.
func TestPracticeSession_StatusTransitions(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")

	if !session.IsInProgress() {
		t.Error("new session should be in progress")
	}

	session.Save()
	if !session.IsSaved() {
		t.Error("session should be saved after Save()")
	}

	session.Complete()
	if !session.IsCompleted() {
		t.Error("session should be completed after Complete()")
	}
}

// TestPracticeSession_Complete_SetsEndTime verifies that Complete sets the end time.
func TestPracticeSession_Complete_SetsEndTime(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	if session.EndTime != nil {
		t.Error("end time should be nil before completion")
	}

	session.Complete()

	if session.EndTime == nil {
		t.Error("end time should be set after completion")
	}
}

// TestPracticeSession_AddExercise_AccumulatesDuration verifies duration accumulation.
func TestPracticeSession_AddExercise_AccumulatesDuration(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	session.AddExercise(SessionExercise{Duration: 30 * time.Second})
	session.AddExercise(SessionExercise{Duration: 45 * time.Second})

	if session.TotalDuration != 75*time.Second {
		t.Errorf("expected 75s total duration, got %v", session.TotalDuration)
	}
	if session.GetExerciseCount() != 2 {
		t.Errorf("expected 2 exercises, got %d", session.GetExerciseCount())
	}
}

// TestPracticeSession_Validate_Valid verifies a valid session passes.
func TestPracticeSession_Validate_Valid(t *testing.T) {
	session := NewPracticeSession("s-001", "user-1")
	if err := session.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// TestPracticeSession_Validate_MissingID verifies that missing ID returns error.
func TestPracticeSession_Validate_MissingID(t *testing.T) {
	session := NewPracticeSession("", "user-1")
	if err := session.Validate(); err == nil {
		t.Error("expected error for missing session ID")
	}
}

// TestPracticeSession_Validate_MissingUserID verifies that missing user ID returns error.
func TestPracticeSession_Validate_MissingUserID(t *testing.T) {
	session := NewPracticeSession("s-001", "")
	if err := session.Validate(); err == nil {
		t.Error("expected error for missing user ID")
	}
}
