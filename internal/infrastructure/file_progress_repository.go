package infrastructure

import (
	"errors"
	"fmt"
	"time"

	"speech-practice-app/internal/domain"
)

// FileProgressRepository persists progress data to disk
type FileProgressRepository struct {
	storage *FileStorage
}

// NewFileProgressRepository creates a new FileProgressRepository
func NewFileProgressRepository(storage *FileStorage) *FileProgressRepository {
	return &FileProgressRepository{storage: storage}
}

func (r *FileProgressRepository) progressFilename(userID string, date time.Time) string {
	return fmt.Sprintf("progress_%s_%s.json", userID, date.Format("2006-01-02"))
}

func (r *FileProgressRepository) streakFilename(userID string) string {
	return fmt.Sprintf("streak_%s.json", userID)
}

func (r *FileProgressRepository) achievementsFilename(userID string) string {
	return fmt.Sprintf("achievements_%s.json", userID)
}

func (r *FileProgressRepository) categoryProgressFilename(userID string) string {
	return fmt.Sprintf("category_progress_%s.json", userID)
}

// SaveProgress saves or updates a progress record
func (r *FileProgressRepository) SaveProgress(record *domain.ProgressRecord) error {
	if record.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	date := time.Date(record.Date.Year(), record.Date.Month(), record.Date.Day(), 0, 0, 0, 0, record.Date.Location())
	return r.storage.SaveJSON(r.progressFilename(record.UserID, date), record)
}

// GetProgress returns progress for a specific user and date
func (r *FileProgressRepository) GetProgress(userID string, date time.Time) (*domain.ProgressRecord, error) {
	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	var record domain.ProgressRecord
	if err := r.storage.LoadJSON(r.progressFilename(userID, normalizedDate), &record); err != nil {
		return nil, errors.New("progress record not found")
	}
	return &record, nil
}

// GetStreak returns the streak record for a user
func (r *FileProgressRepository) GetStreak(userID string) (*domain.StreakRecord, error) {
	var streak domain.StreakRecord
	if err := r.storage.LoadJSON(r.streakFilename(userID), &streak); err != nil {
		return domain.NewStreakRecord(userID), nil
	}
	return &streak, nil
}

// GetCategoryProgress returns progress for a specific category
func (r *FileProgressRepository) GetCategoryProgress(userID string, category domain.ExerciseCategory) (*domain.CategoryProgress, error) {
	all, err := r.loadAllCategoryProgress(userID)
	if err != nil {
		return nil, errors.New("category progress not found")
	}
	if cp, ok := all[string(category)]; ok {
		return &cp, nil
	}
	return nil, errors.New("category progress not found")
}

// GetAchievements returns all achievements for a user
func (r *FileProgressRepository) GetAchievements(userID string) ([]domain.Achievement, error) {
	var achievements []domain.Achievement
	if err := r.storage.LoadJSON(r.achievementsFilename(userID), &achievements); err != nil {
		return []domain.Achievement{}, nil
	}
	return achievements, nil
}

// SaveStreak saves or updates a streak record
func (r *FileProgressRepository) SaveStreak(streak *domain.StreakRecord) error {
	if streak.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	streak.UpdatedAt = time.Now()
	return r.storage.SaveJSON(r.streakFilename(streak.UserID), streak)
}

// SaveCategoryProgress saves or updates category progress
func (r *FileProgressRepository) SaveCategoryProgress(progress *domain.CategoryProgress) error {
	if progress.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	all, _ := r.loadAllCategoryProgress(progress.UserID)
	progress.UpdatedAt = time.Now()
	all[string(progress.Category)] = *progress
	return r.storage.SaveJSON(r.categoryProgressFilename(progress.UserID), all)
}

// SaveAchievement saves or updates an achievement
func (r *FileProgressRepository) SaveAchievement(achievement *domain.Achievement) error {
	if achievement.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	achievements, _ := r.GetAchievements(achievement.UserID)
	// Update or append
	found := false
	for i, a := range achievements {
		if a.ID == achievement.ID {
			achievements[i] = *achievement
			found = true
			break
		}
	}
	if !found {
		achievements = append(achievements, *achievement)
	}
	return r.storage.SaveJSON(r.achievementsFilename(achievement.UserID), achievements)
}

// GetAllProgress returns all progress records for a user
func (r *FileProgressRepository) GetAllProgress(userID string) ([]domain.ProgressRecord, error) {
	prefix := fmt.Sprintf("progress_%s_", userID)
	files, err := r.storage.List(prefix)
	if err != nil {
		return []domain.ProgressRecord{}, nil
	}
	var result []domain.ProgressRecord
	for _, f := range files {
		var record domain.ProgressRecord
		if err := r.storage.LoadJSON(f, &record); err == nil {
			result = append(result, record)
		}
	}
	return result, nil
}

// loadAllCategoryProgress loads the category progress map for a user
func (r *FileProgressRepository) loadAllCategoryProgress(userID string) (map[string]domain.CategoryProgress, error) {
	var all map[string]domain.CategoryProgress
	if err := r.storage.LoadJSON(r.categoryProgressFilename(userID), &all); err != nil {
		return make(map[string]domain.CategoryProgress), nil
	}
	return all, nil
}
