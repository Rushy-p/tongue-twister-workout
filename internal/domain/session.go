package domain

import (
	"time"
)

// SessionStatus represents the status of a practice session
type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusCompleted  SessionStatus = "completed"
	SessionStatusSaved      SessionStatus = "saved"
	SessionStatusAbandoned  SessionStatus = "abandoned"
)

// PracticeSession represents a single practice session
type PracticeSession struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         *time.Time       `json:"end_time,omitempty"`
	Exercises       []SessionExercise `json:"exercises"`
	TotalDuration   time.Duration    `json:"total_duration"`
	Status          SessionStatus    `json:"status"`
	Notes           string           `json:"notes"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Duration    `json:"updated_at"`
}

// SessionExercise represents an exercise within a practice session
type SessionExercise struct {
	ID              string        `json:"id"`
	SessionID       string        `json:"session_id"`
	ExerciseID      string        `json:"exercise_id"`
	ExerciseName    string        `json:"exercise_name"`
	CompletionTime  time.Time     `json:"completion_time"`
	RepetitionsCompleted int      `json:"repetitions_completed"`
	Duration        time.Duration `json:"duration"`
	PerformanceNotes string       `json:"performance_notes,omitempty"`
	Score           int           `json:"score"`
}

// NewPracticeSession creates a new practice session
func NewPracticeSession(id, userID string) *PracticeSession {
	now := time.Now()
	return &PracticeSession{
		ID:            id,
		UserID:        userID,
		StartTime:     now,
		Exercises:     []SessionExercise{},
		TotalDuration: 0,
		Status:        SessionStatusInProgress,
		CreatedAt:     now,
	}
}

// AddExercise adds an exercise to the session
func (s *PracticeSession) AddExercise(exercise SessionExercise) {
	s.Exercises = append(s.Exercises, exercise)
	s.TotalDuration += exercise.Duration
}

// Complete marks the session as completed
func (s *PracticeSession) Complete() {
	now := time.Now()
	s.EndTime = &now
	s.Status = SessionStatusCompleted
	s.TotalDuration = now.Sub(s.StartTime)
}

// Save saves the session for later resumption
func (s *PracticeSession) Save() {
	s.Status = SessionStatusSaved
}

// Abandon marks the session as abandoned
func (s *PracticeSession) Abandon() {
	s.Status = SessionStatusAbandoned
}

// GetExerciseCount returns the number of exercises completed
func (s *PracticeSession) GetExerciseCount() int {
	return len(s.Exercises)
}

// GetCompletedExercises returns only completed exercises
func (s *PracticeSession) GetCompletedExercises() []SessionExercise {
	completed := make([]SessionExercise, 0)
	for _, e := range s.Exercises {
		if e.Score > 0 || e.RepetitionsCompleted > 0 {
			completed = append(completed, e)
		}
	}
	return completed
}

// IsCompleted returns true if the session is completed
func (s *PracticeSession) IsCompleted() bool {
	return s.Status == SessionStatusCompleted
}

// IsInProgress returns true if the session is in progress
func (s *PracticeSession) IsInProgress() bool {
	return s.Status == SessionStatusInProgress
}

// IsSaved returns true if the session is saved for later
func (s *PracticeSession) IsSaved() bool {
	return s.Status == SessionStatusSaved
}

// GetDuration returns the actual session duration
func (s *PracticeSession) GetDuration() time.Duration {
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}

// CalculateSessionStats calculates session statistics
func (s *PracticeSession) CalculateSessionStats() SessionStats {
	stats := SessionStats{
		TotalExercises: len(s.Exercises),
		TotalDuration:  s.GetDuration(),
	}
	
	for _, e := range s.Exercises {
		stats.TotalRepetitions += e.RepetitionsCompleted
		stats.TotalScore += e.Score
	}
	
	if stats.TotalExercises > 0 {
		stats.AverageScore = stats.TotalScore / stats.TotalExercises
	}
	
	return stats
}

// SessionStats represents statistics for a practice session
type SessionStats struct {
	TotalExercises   int           `json:"total_exercises"`
	TotalRepetitions int           `json:"total_repetitions"`
	TotalDuration    time.Duration `json:"total_duration"`
	TotalScore       int           `json:"total_score"`
	AverageScore     int           `json:"average_score"`
}

// Validate checks if the session has valid data
func (s *PracticeSession) Validate() error {
	if s.ID == "" {
		return &ValidationError{"session ID cannot be empty"}
	}
	if s.UserID == "" {
		return &ValidationError{"user ID cannot be empty"}
	}
	return nil
}

// Validate checks if the session exercise has valid data
func (e *SessionExercise) Validate() error {
	if e.ID == "" {
		return &ValidationError{"session exercise ID cannot be empty"}
	}
	if e.SessionID == "" {
		return &ValidationError{"session ID cannot be empty"}
	}
	if e.ExerciseID == "" {
		return &ValidationError{"exercise ID cannot be empty"}
	}
	return nil
}