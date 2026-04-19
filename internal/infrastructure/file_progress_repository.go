package infrastructure

import (
	"errors"
	"time"

	"speech-practice-app/internal/domain"
)

// FileProgressRepository wraps InMemoryProgressRepository with file persistence (Req 14.1)
type FileProgressRepository struct {
	*InMemoryProgressRepository
	storage *FileStorage
}

// NewFileProgressRepository creates a FileProgressRepository, loading existing data from disk
func NewFileProgressRepository(storage *FileStorage) (*FileProgressRepository, error) {
	repo := &FileProgressRepository{
		InMemoryProgressRepository: NewInMemoryProgressRepository(),
		storage:                    storage,
	}

	stored, err := storage.Load()
	if err != nil {
		return nil, err
	}

	// Load progress records
	for _, p := range stored.ProgressRecords {
		p := p
		date := time.Date(p.Date.Year(), p.Date.Month(), p.Date.Day(), 0, 0, 0, 0, p.Date.Location())
		if repo.InMemoryProgressRepository.progress[p.UserID] == nil {
			repo.InMemoryProgressRepository.progress[p.UserID] = make(map[time.Time]domain.ProgressRecord)
		}
		repo.InMemoryProgressRepository.progress[p.UserID][date] = p
	}

	// Load streak records
	for userID, streak := range stored.StreakRecords {
		streak := streak
		repo.InMemoryProgressRepository.streaks[userID] = streak
	}

	// Load category progress
	for userID, catMap := range stored.CategoryProgress {
		repo.InMemoryProgressRepository.categoryProgress[userID] = make(map[domain.ExerciseCategory]domain.CategoryProgress)
		for cat, cp := range catMap {
			cp := cp
			repo.InMemoryProgressRepository.categoryProgress[userID][domain.ExerciseCategory(cat)] = cp
		}
	}

	// Load achievements
	for userID, achList := range stored.Achievements {
		repo.InMemoryProgressRepository.achievements[userID] = make(map[string]domain.Achievement)
		for _, a := range achList {
			a := a
			repo.InMemoryProgressRepository.achievements[userID][a.ID] = a
		}
	}

	return repo, nil
}

// SaveProgress saves a progress record in memory and persists to disk
func (r *FileProgressRepository) SaveProgress(record *domain.ProgressRecord) error {
	if err := r.InMemoryProgressRepository.SaveProgress(record); err != nil {
		return err
	}
	return r.persist()
}

// SaveStreak saves a streak record in memory and persists to disk
func (r *FileProgressRepository) SaveStreak(streak *domain.StreakRecord) error {
	if err := r.InMemoryProgressRepository.SaveStreak(streak); err != nil {
		return err
	}
	return r.persist()
}

// SaveCategoryProgress saves category progress in memory and persists to disk
func (r *FileProgressRepository) SaveCategoryProgress(progress *domain.CategoryProgress) error {
	if err := r.InMemoryProgressRepository.SaveCategoryProgress(progress); err != nil {
		return err
	}
	return r.persist()
}

// SaveAchievement saves an achievement in memory and persists to disk
func (r *FileProgressRepository) SaveAchievement(achievement *domain.Achievement) error {
	if err := r.InMemoryProgressRepository.SaveAchievement(achievement); err != nil {
		return err
	}
	return r.persist()
}

// persist writes the current in-memory progress data to the shared data file
func (r *FileProgressRepository) persist() error {
	r.InMemoryProgressRepository.mu.RLock()

	// Flatten progress records
	var progressRecords []domain.ProgressRecord
	for _, dateMap := range r.InMemoryProgressRepository.progress {
		for _, rec := range dateMap {
			progressRecords = append(progressRecords, rec)
		}
	}

	// Copy streak records
	streaks := make(map[string]domain.StreakRecord, len(r.InMemoryProgressRepository.streaks))
	for k, v := range r.InMemoryProgressRepository.streaks {
		streaks[k] = v
	}

	// Copy category progress (convert ExerciseCategory key to string)
	catProgress := make(map[string]map[string]domain.CategoryProgress)
	for userID, catMap := range r.InMemoryProgressRepository.categoryProgress {
		catProgress[userID] = make(map[string]domain.CategoryProgress)
		for cat, cp := range catMap {
			catProgress[userID][string(cat)] = cp
		}
	}

	// Copy achievements (convert map to slice)
	achievements := make(map[string][]domain.Achievement)
	for userID, achMap := range r.InMemoryProgressRepository.achievements {
		list := make([]domain.Achievement, 0, len(achMap))
		for _, a := range achMap {
			list = append(list, a)
		}
		achievements[userID] = list
	}

	r.InMemoryProgressRepository.mu.RUnlock()

	stored, err := r.storage.Load()
	if err != nil {
		return err
	}

	stored.ProgressRecords = progressRecords
	stored.StreakRecords = streaks
	stored.CategoryProgress = catProgress
	stored.Achievements = achievements

	return r.storage.Save(stored)
}

// Ensure FileProgressRepository satisfies ProgressRepository
var _ ProgressRepository = (*FileProgressRepository)(nil)

// GetProgress returns progress for a specific user and date
func (r *FileProgressRepository) GetProgress(userID string, date time.Time) (*domain.ProgressRecord, error) {
	return r.InMemoryProgressRepository.GetProgress(userID, date)
}

// GetStreak returns the streak record for a user
func (r *FileProgressRepository) GetStreak(userID string) (*domain.StreakRecord, error) {
	return r.InMemoryProgressRepository.GetStreak(userID)
}

// GetCategoryProgress returns progress for a specific category
func (r *FileProgressRepository) GetCategoryProgress(userID string, category domain.ExerciseCategory) (*domain.CategoryProgress, error) {
	return r.InMemoryProgressRepository.GetCategoryProgress(userID, category)
}

// GetAchievements returns all achievements for a user
func (r *FileProgressRepository) GetAchievements(userID string) ([]domain.Achievement, error) {
	return r.InMemoryProgressRepository.GetAchievements(userID)
}

// GetAllProgress returns all progress records for a user
func (r *FileProgressRepository) GetAllProgress(userID string) ([]domain.ProgressRecord, error) {
	return r.InMemoryProgressRepository.GetAllProgress(userID)
}

// GetByUserID returns all sessions for a user (satisfies ProgressRepository if needed)
func (r *FileProgressRepository) GetByUserID(userID string) ([]domain.ProgressRecord, error) {
	return r.InMemoryProgressRepository.GetAllProgress(userID)
}

// DeleteUserData removes all data for a user via FileStorage (Req 14.4)
func (r *FileProgressRepository) DeleteUserData(userID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	return r.storage.DeleteUserData(userID)
}
