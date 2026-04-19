package infrastructure

import (
	"errors"
	"time"

	"speech-practice-app/internal/domain"
)

// FileSessionRepository wraps InMemorySessionRepository with file persistence (Req 14.1)
type FileSessionRepository struct {
	*InMemorySessionRepository
	storage *FileStorage
}

// NewFileSessionRepository creates a FileSessionRepository, loading existing data from disk
func NewFileSessionRepository(storage *FileStorage) (*FileSessionRepository, error) {
	repo := &FileSessionRepository{
		InMemorySessionRepository: NewInMemorySessionRepository(),
		storage:                   storage,
	}

	stored, err := storage.Load()
	if err != nil {
		return nil, err
	}

	// Populate in-memory store from disk
	for _, s := range stored.Sessions {
		s := s // capture
		repo.InMemorySessionRepository.sessions[s.ID] = s
	}

	return repo, nil
}

// Save stores a session in memory and persists all sessions to disk
func (r *FileSessionRepository) Save(session *domain.PracticeSession) error {
	if err := r.InMemorySessionRepository.Save(session); err != nil {
		return err
	}
	return r.persist()
}

// Delete removes a session from memory and persists the change to disk
func (r *FileSessionRepository) Delete(id string) error {
	if err := r.InMemorySessionRepository.Delete(id); err != nil {
		return err
	}
	return r.persist()
}

// persist writes the current in-memory sessions to the shared data file
func (r *FileSessionRepository) persist() error {
	r.InMemorySessionRepository.mu.RLock()
	sessions := make([]domain.PracticeSession, 0, len(r.InMemorySessionRepository.sessions))
	for _, s := range r.InMemorySessionRepository.sessions {
		sessions = append(sessions, s)
	}
	r.InMemorySessionRepository.mu.RUnlock()

	stored, err := r.storage.Load()
	if err != nil {
		return err
	}
	stored.Sessions = sessions
	return r.storage.Save(stored)
}

// GetByDateRange returns sessions within a date range
func (r *FileSessionRepository) GetByDateRange(start, end time.Time) ([]domain.PracticeSession, error) {
	r.InMemorySessionRepository.mu.RLock()
	defer r.InMemorySessionRepository.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.InMemorySessionRepository.sessions {
		if (s.StartTime.After(start) || s.StartTime.Equal(start)) &&
			(s.StartTime.Before(end) || s.StartTime.Equal(end)) {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetIncompleteSessions returns all incomplete sessions for a user
func (r *FileSessionRepository) GetIncompleteSessions(userID string) ([]domain.PracticeSession, error) {
	r.InMemorySessionRepository.mu.RLock()
	defer r.InMemorySessionRepository.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.InMemorySessionRepository.sessions {
		if s.UserID == userID && (s.Status == domain.SessionStatusInProgress || s.Status == domain.SessionStatusSaved) {
			result = append(result, s)
		}
	}
	return result, nil
}

// Ensure FileSessionRepository satisfies SessionRepository
var _ SessionRepository = (*FileSessionRepository)(nil)

// GetByID returns a session by its ID
func (r *FileSessionRepository) GetByID(id string) (*domain.PracticeSession, error) {
	r.InMemorySessionRepository.mu.RLock()
	defer r.InMemorySessionRepository.mu.RUnlock()

	session, exists := r.InMemorySessionRepository.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return &session, nil
}

// GetByUserID returns all sessions for a user
func (r *FileSessionRepository) GetByUserID(userID string) ([]domain.PracticeSession, error) {
	r.InMemorySessionRepository.mu.RLock()
	defer r.InMemorySessionRepository.mu.RUnlock()

	var result []domain.PracticeSession
	for _, s := range r.InMemorySessionRepository.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}
