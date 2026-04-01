package infrastructure

import (
	"errors"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

// ProgressRepository defines the interface for progress data access
type ProgressRepository interface {
	SaveProgress(record *domain.ProgressRecord) error
	GetProgress(userID string, date time.Time) (*domain.ProgressRecord, error)
	GetStreak(userID string) (*domain.StreakRecord, error)
	GetCategoryProgress(userID string, category domain.ExerciseCategory) (*domain.CategoryProgress, error)
	GetAchievements(userID string) ([]domain.Achievement, error)
	SaveStreak(streak *domain.StreakRecord) error
	SaveCategoryProgress(progress *domain.CategoryProgress) error
	SaveAchievement(achievement *domain.Achievement) error
	GetAllProgress(userID string) ([]domain.ProgressRecord, error)
}

// InMemoryProgressRepository provides in-memory storage for progress data
type InMemoryProgressRepository struct {
	progress         map[string]map[time.Time]domain.ProgressRecord // userID -> date -> record
	streaks          map[string]domain.StreakRecord                  // userID -> streak
	categoryProgress map[string]map[domain.ExerciseCategory]domain.CategoryProgress // userID -> category -> progress
	achievements     map[string]map[string]domain.Achievement        // userID -> achievementID -> achievement
	mu               sync.RWMutex
}

// NewInMemoryProgressRepository creates a new in-memory progress repository
func NewInMemoryProgressRepository() *InMemoryProgressRepository {
	return &InMemoryProgressRepository{
		progress:         make(map[string]map[time.Time]domain.ProgressRecord),
		streaks:          make(map[string]domain.StreakRecord),
		categoryProgress: make(map[string]map[domain.ExerciseCategory]domain.CategoryProgress),
		achievements:     make(map[string]map[string]domain.Achievement),
	}
}

// SaveProgress saves or updates a progress record
func (r *InMemoryProgressRepository) SaveProgress(record *domain.ProgressRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if record.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Normalize date to start of day
	date := time.Date(record.Date.Year(), record.Date.Month(), record.Date.Day(), 0, 0, 0, 0, record.Date.Location())

	if r.progress[record.UserID] == nil {
		r.progress[record.UserID] = make(map[time.Time]domain.ProgressRecord)
	}
	r.progress[record.UserID][date] = *record
	return nil
}

// GetProgress returns progress for a specific user and date
func (r *InMemoryProgressRepository) GetProgress(userID string, date time.Time) (*domain.ProgressRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	if userProgress, exists := r.progress[userID]; exists {
		if record, found := userProgress[normalizedDate]; found {
			return &record, nil
		}
	}
	return nil, errors.New("progress record not found")
}

// GetStreak returns the streak record for a user
func (r *InMemoryProgressRepository) GetStreak(userID string) (*domain.StreakRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if streak, exists := r.streaks[userID]; exists {
		return &streak, nil
	}
	// Return a new streak record if none exists
	return domain.NewStreakRecord(userID), nil
}

// GetCategoryProgress returns progress for a specific category
func (r *InMemoryProgressRepository) GetCategoryProgress(userID string, category domain.ExerciseCategory) (*domain.CategoryProgress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userProgress, exists := r.categoryProgress[userID]; exists {
		if progress, found := userProgress[category]; found {
			return &progress, nil
		}
	}
	return nil, errors.New("category progress not found")
}

// GetAchievements returns all achievements for a user
func (r *InMemoryProgressRepository) GetAchievements(userID string) ([]domain.Achievement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userAchievements, exists := r.achievements[userID]; exists {
		result := make([]domain.Achievement, 0, len(userAchievements))
		for _, a := range userAchievements {
			result = append(result, a)
		}
		return result, nil
	}
	return []domain.Achievement{}, nil
}

// SaveStreak saves or updates a streak record
func (r *InMemoryProgressRepository) SaveStreak(streak *domain.StreakRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if streak.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	streak.UpdatedAt = time.Now()
	r.streaks[streak.UserID] = *streak
	return nil
}

// SaveCategoryProgress saves or updates category progress
func (r *InMemoryProgressRepository) SaveCategoryProgress(progress *domain.CategoryProgress) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if progress.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	if r.categoryProgress[progress.UserID] == nil {
		r.categoryProgress[progress.UserID] = make(map[domain.ExerciseCategory]domain.CategoryProgress)
	}
	progress.UpdatedAt = time.Now()
	r.categoryProgress[progress.UserID][progress.Category] = *progress
	return nil
}

// SaveAchievement saves or updates an achievement
func (r *InMemoryProgressRepository) SaveAchievement(achievement *domain.Achievement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if achievement.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	if r.achievements[achievement.UserID] == nil {
		r.achievements[achievement.UserID] = make(map[string]domain.Achievement)
	}
	r.achievements[achievement.UserID][achievement.ID] = *achievement
	return nil
}

// GetAllProgress returns all progress records for a user
func (r *InMemoryProgressRepository) GetAllProgress(userID string) ([]domain.ProgressRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userProgress, exists := r.progress[userID]; exists {
		result := make([]domain.ProgressRecord, 0, len(userProgress))
		for _, record := range userProgress {
			result = append(result, record)
		}
		return result, nil
	}
	return []domain.ProgressRecord{}, nil
}