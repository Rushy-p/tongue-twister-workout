package infrastructure

import (
	"errors"
	"fmt"
	"time"

	"speech-practice-app/internal/domain"
)

// sessionsIndex maps userID to a list of session IDs
type sessionsIndex map[string][]string

// FileSessionRepository persists practice sessions to disk
type FileSessionRepository struct {
	storage *FileStorage
}

// NewFileSessionRepository creates a new FileSessionRepository
func NewFileSessionRepository(storage *FileStorage) *FileSessionRepository {
	return &FileSessionRepository{storage: storage}
}

func (r *FileSessionRepository) sessionFilename(sessionID string) string {
	return fmt.Sprintf("session_%s.json", sessionID)
}

const sessionsIndexFile = "sessions_index.json"

func (r *FileSessionRepository) loadIndex() sessionsIndex {
	var idx sessionsIndex
	if err := r.storage.LoadJSON(sessionsIndexFile, &idx); err != nil {
		return make(sessionsIndex)
	}
	return idx
}

func (r *FileSessionRepository) saveIndex(idx sessionsIndex) error {
	return r.storage.SaveJSON(sessionsIndexFile, idx)
}

// Save stores a practice session and updates the index
func (r *FileSessionRepository) Save(session *domain.PracticeSession) error {
	if session.ID == "" {
		return errors.New("session ID cannot be empty")
	}
	if session.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	session.UpdatedAt = time.Since(session.StartTime)
	if err := r.storage.SaveJSON(r.sessionFilename(session.ID), session); err != nil {
		return err
	}
	// Update index
	idx := r.loadIndex()
	ids := idx[session.UserID]
	found := false
	for _, id := range ids {
		if id == session.ID {
			found = true
			break
		}
	}
	if !found {
		idx[session.UserID] = append(ids, session.ID)
	}
	return r.saveIndex(idx)
}

// GetByID returns a session by its ID
func (r *FileSessionRepository) GetByID(id string) (*domain.PracticeSession, error) {
	var session domain.PracticeSession
	if err := r.storage.LoadJSON(r.sessionFilename(id), &session); err != nil {
		return nil, errors.New("session not found")
	}
	return &session, nil
}

// GetByUserID returns all sessions for a user
func (r *FileSessionRepository) GetByUserID(userID string) ([]domain.PracticeSession, error) {
	idx := r.loadIndex()
	ids := idx[userID]
	var result []domain.PracticeSession
	for _, id := range ids {
		s, err := r.GetByID(id)
		if err == nil {
			result = append(result, *s)
		}
	}
	return result, nil
}

// GetByDateRange returns sessions within a date range
func (r *FileSessionRepository) GetByDateRange(start, end time.Time) ([]domain.PracticeSession, error) {
	files, err := r.storage.List("session_")
	if err != nil {
		return nil, err
	}
	var result []domain.PracticeSession
	for _, f := range files {
		var s domain.PracticeSession
		if err := r.storage.LoadJSON(f, &s); err != nil {
			continue
		}
		if (s.StartTime.After(start) || s.StartTime.Equal(start)) &&
			(s.StartTime.Before(end) || s.StartTime.Equal(end)) {
			result = append(result, s)
		}
	}
	return result, nil
}

// GetIncompleteSessions returns all incomplete sessions for a user
func (r *FileSessionRepository) GetIncompleteSessions(userID string) ([]domain.PracticeSession, error) {
	sessions, err := r.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	var result []domain.PracticeSession
	for _, s := range sessions {
		if s.Status == domain.SessionStatusInProgress || s.Status == domain.SessionStatusSaved {
			result = append(result, s)
		}
	}
	return result, nil
}

// Delete removes a session and updates the index
func (r *FileSessionRepository) Delete(id string) error {
	session, err := r.GetByID(id)
	if err != nil {
		return errors.New("session not found")
	}
	if err := r.storage.Delete(r.sessionFilename(id)); err != nil {
		return err
	}
	// Remove from index
	idx := r.loadIndex()
	ids := idx[session.UserID]
	newIDs := ids[:0]
	for _, sid := range ids {
		if sid != id {
			newIDs = append(newIDs, sid)
		}
	}
	idx[session.UserID] = newIDs
	return r.saveIndex(idx)
}
