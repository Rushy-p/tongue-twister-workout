package infrastructure

import (
	"errors"
	"time"

	"speech-practice-app/internal/domain"
)

// FilePreferencesRepository wraps InMemoryPreferencesRepository with file persistence (Req 14.1)
type FilePreferencesRepository struct {
	*InMemoryPreferencesRepository
	storage *FileStorage
}

// NewFilePreferencesRepository creates a FilePreferencesRepository, loading existing data from disk
func NewFilePreferencesRepository(storage *FileStorage) (*FilePreferencesRepository, error) {
	repo := &FilePreferencesRepository{
		InMemoryPreferencesRepository: NewInMemoryPreferencesRepository(),
		storage:                       storage,
	}

	stored, err := storage.Load()
	if err != nil {
		return nil, err
	}

	for userID, prefs := range stored.Preferences {
		prefs := prefs // capture
		repo.InMemoryPreferencesRepository.preferences[userID] = prefs
	}

	return repo, nil
}

// Save stores preferences in memory and persists to disk
func (r *FilePreferencesRepository) Save(preferences *domain.UserPreferences) error {
	if preferences.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	preferences.UpdatedAt = time.Now()

	r.InMemoryPreferencesRepository.mu.Lock()
	r.InMemoryPreferencesRepository.preferences[preferences.UserID] = *preferences
	r.InMemoryPreferencesRepository.mu.Unlock()

	return r.persist()
}

// Delete removes preferences from memory and persists the change
func (r *FilePreferencesRepository) Delete(userID string) error {
	r.InMemoryPreferencesRepository.mu.Lock()
	if _, exists := r.InMemoryPreferencesRepository.preferences[userID]; !exists {
		r.InMemoryPreferencesRepository.mu.Unlock()
		return errors.New("preferences not found")
	}
	delete(r.InMemoryPreferencesRepository.preferences, userID)
	r.InMemoryPreferencesRepository.mu.Unlock()

	return r.persist()
}

// Export exports all user data using FileStorage for a complete export (Req 9.7, 14.5)
func (r *FilePreferencesRepository) Export(userID string, format domain.ExportFormat) (string, error) {
	return r.storage.ExportUserData(userID, format)
}

// persist writes the current in-memory preferences to the shared data file
func (r *FilePreferencesRepository) persist() error {
	r.InMemoryPreferencesRepository.mu.RLock()
	prefs := make(map[string]domain.UserPreferences, len(r.InMemoryPreferencesRepository.preferences))
	for k, v := range r.InMemoryPreferencesRepository.preferences {
		prefs[k] = v
	}
	r.InMemoryPreferencesRepository.mu.RUnlock()

	stored, err := r.storage.Load()
	if err != nil {
		return err
	}
	stored.Preferences = prefs
	return r.storage.Save(stored)
}

// Ensure FilePreferencesRepository satisfies PreferencesRepository
var _ PreferencesRepository = (*FilePreferencesRepository)(nil)
