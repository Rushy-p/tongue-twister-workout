package service

import (
	"errors"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

// SessionService handles session-related business logic
type SessionService struct {
	sessionRepo  infrastructure.SessionRepository
	progressRepo infrastructure.ProgressRepository
	exerciseRepo infrastructure.ExerciseRepository
}

// NewSessionService creates a new SessionService
func NewSessionService(
	sessionRepo infrastructure.SessionRepository,
	progressRepo infrastructure.ProgressRepository,
	exerciseRepo infrastructure.ExerciseRepository,
) *SessionService {
	return &SessionService{
		sessionRepo:  sessionRepo,
		progressRepo: progressRepo,
		exerciseRepo: exerciseRepo,
	}
}

// StartSession creates a new practice session
// Implements Requirement 6.1
func (s *SessionService) StartSession(userID string) (*domain.PracticeSession, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	session := domain.NewPracticeSession(generateID(), userID)
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

// ResumeSession resumes an incomplete session
// Implements Requirement 6.8
func (s *SessionService) ResumeSession(sessionID string) (*domain.PracticeSession, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	// Mark as in progress
	session.Status = domain.SessionStatusInProgress
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

// CompleteSession completes a practice session
// Implements Requirement 6.2
func (s *SessionService) CompleteSession(sessionID string) (*domain.PracticeSession, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	// Complete the session
	session.Complete()

	// Save session data
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	// Update streak
	if err := s.updateStreak(session.UserID); err != nil {
		return nil, err
	}

	// Record progress
	if err := s.recordProgress(session); err != nil {
		return nil, err
	}

	return session, nil
}

// SaveSession saves a session for later resumption
// Implements Requirement 6.7
func (s *SessionService) SaveSession(sessionID string) (*domain.PracticeSession, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	session.Save()
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return session, nil
}

// AbandonSession abandons a session without saving
func (s *SessionService) AbandonSession(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return err
	}

	session.Abandon()
	return s.sessionRepo.Save(session)
}

// AddExerciseToSession adds an exercise to a session
// Implements Requirement 6.3
func (s *SessionService) AddExerciseToSession(sessionID string, exerciseID string, repetitions int, score int, notes string) (*domain.SessionExercise, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}
	if exerciseID == "" {
		return nil, errors.New("exercise ID cannot be empty")
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	// Get exercise details
	exercise, err := s.exerciseRepo.GetByID(exerciseID)
	if err != nil {
		return nil, err
	}

	// Create session exercise
	sessionExercise := domain.SessionExercise{
		ID:                   generateID(),
		SessionID:            sessionID,
		ExerciseID:           exerciseID,
		ExerciseName:         exercise.Name,
		CompletionTime:       time.Now(),
		RepetitionsCompleted: repetitions,
		Duration:             exercise.Duration,
		PerformanceNotes:     notes,
		Score:                score,
	}

	// Add to session
	session.AddExercise(sessionExercise)

	// Save session
	if err := s.sessionRepo.Save(session); err != nil {
		return nil, err
	}

	return &sessionExercise, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(sessionID string) (*domain.PracticeSession, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}
	return s.sessionRepo.GetByID(sessionID)
}

// GetUserSessions retrieves all sessions for a user
func (s *SessionService) GetUserSessions(userID string) ([]domain.PracticeSession, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.sessionRepo.GetByUserID(userID)
}

// GetIncompleteSessions retrieves incomplete sessions for a user
// Implements Requirement 6.8
func (s *SessionService) GetIncompleteSessions(userID string) ([]domain.PracticeSession, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.sessionRepo.GetIncompleteSessions(userID)
}

// GetSessionStats retrieves statistics for a session
func (s *SessionService) GetSessionStats(sessionID string) (*domain.SessionStats, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	stats := session.CalculateSessionStats()
	return &stats, nil
}

// GetUserStats retrieves overall statistics for a user
func (s *SessionService) GetUserStats(userID string) (*UserStats, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	sessions, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	stats := &UserStats{
		TotalSessions:      len(sessions),
		TotalExercises:     0,
		TotalRepetitions:   0,
		TotalDuration:      0,
		CompletedSessions:  0,
		IncompleteSessions: 0,
	}

	for _, session := range sessions {
		stats.TotalExercises += session.GetExerciseCount()
		stats.TotalDuration += session.TotalDuration

		if session.IsCompleted() {
			stats.CompletedSessions++
		} else if session.IsInProgress() || session.IsSaved() {
			stats.IncompleteSessions++
		}

		for _, e := range session.Exercises {
			stats.TotalRepetitions += e.RepetitionsCompleted
		}
	}

	// Get streak
	streak, err := s.progressRepo.GetStreak(userID)
	if err == nil {
		stats.CurrentStreak = streak.CurrentStreak
		stats.LongestStreak = streak.LongestStreak
	}

	return stats, nil
}

// UserStats represents overall user statistics
type UserStats struct {
	TotalSessions      int
	TotalExercises     int
	TotalRepetitions   int
	TotalDuration      time.Duration
	CompletedSessions  int
	IncompleteSessions int
	CurrentStreak      int
	LongestStreak      int
}

// updateStreak updates the user's practice streak
// Implements Requirements 6.4, 6.5
func (s *SessionService) updateStreak(userID string) error {
	streak, err := s.progressRepo.GetStreak(userID)
	if err != nil {
		streak = domain.NewStreakRecord(userID)
	}

	streak.UpdateStreak()
	return s.progressRepo.SaveStreak(streak)
}

// recordProgress records progress for a completed session
func (s *SessionService) recordProgress(session *domain.PracticeSession) error {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Create progress record
	record := domain.NewProgressRecord(session.UserID, domain.CategoryMouthExercise)
	record.Date = date
	record.ExerciseCount = session.GetExerciseCount()
	record.Duration = session.TotalDuration
	record.Completed = true

	if err := s.progressRepo.SaveProgress(record); err != nil {
		return err
	}

	// Update category progress
	for _, se := range session.Exercises {
		exercise, err := s.exerciseRepo.GetByID(se.ExerciseID)
		if err != nil {
			continue
		}

		catProgress, err := s.progressRepo.GetCategoryProgress(session.UserID, exercise.Category)
		if err != nil {
			// Create new category progress
			allExercises, _ := s.exerciseRepo.GetByCategory(exercise.Category)
			catProgress = domain.NewCategoryProgress(session.UserID, exercise.Category, len(allExercises))
		}

		catProgress.CompletedExercises++
		catProgress.TotalTime += se.Duration

		if err := s.progressRepo.SaveCategoryProgress(catProgress); err != nil {
			return err
		}
	}

	return nil
}

// TimerService manages exercise timers
type TimerService struct {
	activeTimers map[string]*Timer
}

// Timer represents an active timer
type Timer struct {
	ExerciseID   string
	SessionID    string
	StartTime    time.Time
	Duration     time.Duration
	Remaining    time.Duration
	IsRunning    bool
	OnExpiration func() error
}

// NewTimerService creates a new TimerService
func NewTimerService() *TimerService {
	return &TimerService{
		activeTimers: make(map[string]*Timer),
	}
}

// StartTimer starts a timer for an exercise
// Implements Requirement 6.6
func (ts *TimerService) StartTimer(exerciseID, sessionID string, duration time.Duration) *Timer {
	timer := &Timer{
		ExerciseID: exerciseID,
		SessionID:  sessionID,
		StartTime:  time.Now(),
		Duration:   duration,
		Remaining:  duration,
		IsRunning:  true,
	}

	ts.activeTimers[sessionID] = timer
	return timer
}

// GetTimer returns the timer for a session
func (ts *TimerService) GetTimer(sessionID string) (*Timer, bool) {
	timer, exists := ts.activeTimers[sessionID]
	return timer, exists
}

// UpdateTimer updates the remaining time for a timer
func (ts *TimerService) UpdateTimer(sessionID string) {
	if timer, exists := ts.activeTimers[sessionID]; exists && timer.IsRunning {
		elapsed := time.Since(timer.StartTime)
		timer.Remaining = timer.Duration - elapsed
		if timer.Remaining <= 0 {
			timer.Remaining = 0
		}
	}
}

// StopTimer stops a timer
func (ts *TimerService) StopTimer(sessionID string) {
	if timer, exists := ts.activeTimers[sessionID]; exists {
		timer.IsRunning = false
		ts.UpdateTimer(sessionID)
	}
}

// IsTimerExpired checks if a timer has expired
func (ts *TimerService) IsTimerExpired(sessionID string) bool {
	if timer, exists := ts.activeTimers[sessionID]; exists {
		ts.UpdateTimer(sessionID)
		return timer.Remaining <= 0
	}
	return false
}

// RemoveTimer removes a timer
func (ts *TimerService) RemoveTimer(sessionID string) {
	delete(ts.activeTimers, sessionID)
}

// GetDefaultDuration returns the default duration for an exercise
func (s *SessionService) GetDefaultDuration(exerciseID string) (time.Duration, error) {
	exercise, err := s.exerciseRepo.GetByID(exerciseID)
	if err != nil {
		return 0, err
	}
	return exercise.Duration, nil
}

// generateID generates a unique ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

// MarkExerciseCompleted marks an exercise as completed standalone (outside a session).
// Used when a user clicks "Mark as Complete" from the exercise detail page without an active session.
func (s *SessionService) MarkExerciseCompleted(exerciseID string) error {
	exercise, err := s.exerciseRepo.GetByID(exerciseID)
	if err != nil {
		return err
	}
	return s.exerciseRepo.UpdateCompletionCount(exerciseID, exercise.CompletionCount+1)
}
