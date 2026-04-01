package infrastructure

import (
	"errors"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

// SessionRepository defines the interface for practice session data access
type SessionRepository interface {
	Save(session *domain.PracticeSession) error
	GetByID(id string) (*domain.PracticeSession, error)
	GetByUserID(userID string) ([]domain.PracticeSession, error)
	GetByDateRange(start, end time.Time) ([]domain.PracticeSession, error)
	GetIncompleteSessions(userID string) ([]domain.PracticeSession, error)
	Delete(id string) error
}

// InMemorySessionRepository provides in-memory storage for practice sessions
type InMemorySessionRepository struct {
	sessions map[string]domain.PracticeSession
	mu       sync.RWMutex
}

// NewInMemorySessionRepository creates a new in-memory session repository
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]domain.PracticeSession),
	}
}

// Save stores a practice session
func (r *InMemorySessionRepository) Save(session *domain.PracticeSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == "" {
		return errors.New("session ID cannot be empty")
	}
	if session.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	session.UpdatedAt = time.Since(session.StartTime)
	r.sessions[session.ID] = *session
	return nil
}

// GetByID returns a session by its ID
func (r *InMemorySessionRepository) GetByID(id string) (*domain.PracticeSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return &session, nil
}

// GetByUserID returns all sessions for a user
func (r *InMemorySessionRepository) GetByUserID(userID string) ([]domain.PracticeSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetByDateRange returns sessions within a date range
func (r *InMemorySessionRepository) GetByDateRange(start, end time.Time) ([]domain.PracticeSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.sessions {
		if (s.StartTime.After(start) || s.StartTime.Equal(start)) &&
			(s.StartTime.Before(end) || s.StartTime.Equal(end)) {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetIncompleteSessions returns all incomplete sessions for a user
func (r *InMemorySessionRepository) GetIncompleteSessions(userID string) ([]domain.PracticeSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.sessions {
		if s.UserID == userID && (s.Status == domain.SessionStatusInProgress || s.Status == domain.SessionStatusSaved) {
			result = append(result, s)
		}
	}
	return result, nil
}

// Delete removes a session by ID
func (r *InMemorySessionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[id]; !exists {
		return errors.New("session not found")
	}
	delete(r.sessions, id)
	return nil
}