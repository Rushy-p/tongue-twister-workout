package service

import (
	"errors"
	"sort"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

// ProgressService handles progress tracking and metrics
type ProgressService struct {
	progressRepo infrastructure.ProgressRepository
	sessionRepo  infrastructure.SessionRepository
	exerciseRepo infrastructure.ExerciseRepository
}

// NewProgressService creates a new ProgressService
func NewProgressService(
	progressRepo infrastructure.ProgressRepository,
	sessionRepo infrastructure.SessionRepository,
	exerciseRepo infrastructure.ExerciseRepository,
) *ProgressService {
	return &ProgressService{
		progressRepo: progressRepo,
		sessionRepo:  sessionRepo,
		exerciseRepo: exerciseRepo,
	}
}

// GetCurrentStreak returns the current practice streak for a user
// Implements Requirement 7.2
func (s *ProgressService) GetCurrentStreak(userID string) (int, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	streak, err := s.progressRepo.GetStreak(userID)
	if err != nil {
		return 0, nil
	}

	return streak.CurrentStreak, nil
}

// GetLongestStreak returns the longest practice streak for a user
func (s *ProgressService) GetLongestStreak(userID string) (int, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	streak, err := s.progressRepo.GetStreak(userID)
	if err != nil {
		return 0, nil
	}

	return streak.LongestStreak, nil
}

// GetTotalExercises returns the total number of exercises completed
// Implements Requirement 7.3
func (s *ProgressService) GetTotalExercises(userID string) (int, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	progress, err := s.progressRepo.GetAllProgress(userID)
	if err != nil {
		return 0, nil
	}

	total := 0
	for _, record := range progress {
		total += record.ExerciseCount
	}

	return total, nil
}

// GetTotalPracticeTime returns the total practice time accumulated
// Implements Requirement 7.4
func (s *ProgressService) GetTotalPracticeTime(userID string) (time.Duration, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	progress, err := s.progressRepo.GetAllProgress(userID)
	if err != nil {
		return 0, nil
	}

	var total time.Duration
	for _, record := range progress {
		total += record.Duration
	}

	return total, nil
}

// GetTotalSessions returns the total number of practice sessions
// Implements Requirement 7.1
func (s *ProgressService) GetTotalSessions(userID string) (int, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	sessions, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return 0, nil
	}

	completed := 0
	for _, session := range sessions {
		if session.IsCompleted() {
			completed++
		}
	}

	return completed, nil
}

// GetCategoryProgress returns progress for a specific category
// Implements Requirement 7.7
func (s *ProgressService) GetCategoryProgress(userID string, category domain.ExerciseCategory) (*domain.CategoryProgress, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	progress, err := s.progressRepo.GetCategoryProgress(userID, category)
	if err != nil {
		// Return empty progress if none exists
		allExercises, _ := s.exerciseRepo.GetByCategory(category)
		return domain.NewCategoryProgress(userID, category, len(allExercises)), nil
	}

	return progress, nil
}

// GetAllCategoryProgress returns progress for all categories
// Implements Requirement 7.7
func (s *ProgressService) GetAllCategoryProgress(userID string) (map[domain.ExerciseCategory]*domain.CategoryProgress, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	categories := []domain.ExerciseCategory{
		domain.CategoryMouthExercise,
		domain.CategoryTongueTwister,
		domain.CategoryDictionStrategy,
		domain.CategoryPacingStrategy,
	}

	result := make(map[domain.ExerciseCategory]*domain.CategoryProgress)

	for _, cat := range categories {
		progress, err := s.GetCategoryProgress(userID, cat)
		if err == nil {
			result[cat] = progress
		}
	}

	return result, nil
}

// GetWeeklyCalendar returns the weekly practice calendar
// Implements Requirement 7.6
func (s *ProgressService) GetWeeklyCalendar(userID string) ([]domain.DayActivity, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	progress, err := s.progressRepo.GetAllProgress(userID)
	if err != nil {
		return []domain.DayActivity{}, nil
	}

	// Filter progress for the last 7 days
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	var recentProgress []domain.ProgressRecord
	for _, p := range progress {
		if p.Date.After(sevenDaysAgo) {
			recentProgress = append(recentProgress, p)
		}
	}

	return domain.GetWeeklyCalendar(recentProgress), nil
}

// GetProgressSummary returns a complete progress summary
// Implements Requirements 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7
func (s *ProgressService) GetProgressSummary(userID string) (*domain.ProgressSummary, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	summary := &domain.ProgressSummary{
		CategoryProgress: make(map[string]domain.CategoryProgress),
		Achievements:     []domain.Achievement{},
		WeeklyActivity:   []domain.DayActivity{},
	}

	// Get streak data
	streak, err := s.progressRepo.GetStreak(userID)
	if err == nil {
		summary.CurrentStreak = streak.CurrentStreak
		summary.LongestStreak = streak.LongestStreak
	}

	// Get total sessions
	sessions, _ := s.sessionRepo.GetByUserID(userID)
	summary.TotalSessions = 0
	for _, session := range sessions {
		if session.IsCompleted() {
			summary.TotalSessions++
		}
	}

	// Get total exercises
	progress, _ := s.progressRepo.GetAllProgress(userID)
	summary.TotalExercises = 0
	summary.TotalPracticeTime = 0
	for _, p := range progress {
		summary.TotalExercises += p.ExerciseCount
		summary.TotalPracticeTime += p.Duration
	}

	// Get category progress
	categoryProgress, _ := s.GetAllCategoryProgress(userID)
	for cat, prog := range categoryProgress {
		summary.CategoryProgress[string(cat)] = *prog
	}

	// Get achievements
	achievements, _ := s.progressRepo.GetAchievements(userID)
	summary.Achievements = achievements

	// Get weekly activity
	weeklyActivity, _ := s.GetWeeklyCalendar(userID)
	summary.WeeklyActivity = weeklyActivity

	return summary, nil
}

// CheckMilestones checks and updates milestone achievements
// Implements Requirement 7.9
func (s *ProgressService) CheckMilestones(userID string) ([]domain.Achievement, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	var newlyUnlocked []domain.Achievement

	// Get current progress
	totalExercises, _ := s.GetTotalExercises(userID)
	currentStreak, _ := s.GetCurrentStreak(userID)
	categoryProgress, _ := s.GetAllCategoryProgress(userID)

	// Check first practice achievement
	if totalExercises >= 1 {
		achievement := s.getOrCreateAchievement(userID, "first_practice", domain.AchievementCompletion, 1)
		if !achievement.IsUnlocked() {
			achievement.UpdateProgress(totalExercises)
			if achievement.IsUnlocked() {
				newlyUnlocked = append(newlyUnlocked, *achievement)
			}
			s.progressRepo.SaveAchievement(achievement)
		}
	}

	// Check week streak achievement
	achievement := s.getOrCreateAchievement(userID, "week_streak", domain.AchievementStreak, 7)
	if !achievement.IsUnlocked() {
		achievement.UpdateProgress(currentStreak)
		if achievement.IsUnlocked() {
			newlyUnlocked = append(newlyUnlocked, *achievement)
		}
		s.progressRepo.SaveAchievement(achievement)
	}

	// Check month streak achievement
	achievement = s.getOrCreateAchievement(userID, "month_streak", domain.AchievementStreak, 30)
	if !achievement.IsUnlocked() {
		achievement.UpdateProgress(currentStreak)
		if achievement.IsUnlocked() {
			newlyUnlocked = append(newlyUnlocked, *achievement)
		}
		s.progressRepo.SaveAchievement(achievement)
	}

	// Check category achievements
	for cat, prog := range categoryProgress {
		switch cat {
		case domain.CategoryTongueTwister:
			achievement = s.getOrCreateAchievement(userID, "tongue_twister_master", domain.AchievementCategory, 100)
			if !achievement.IsUnlocked() {
				achievement.UpdateProgress(prog.CompletedExercises)
				if achievement.IsUnlocked() {
					newlyUnlocked = append(newlyUnlocked, *achievement)
				}
				s.progressRepo.SaveAchievement(achievement)
			}
		}
	}

	return newlyUnlocked, nil
}

// GetAchievements returns all achievements for a user
// Implements Requirement 7.9
func (s *ProgressService) GetAchievements(userID string) ([]domain.Achievement, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	achievements, err := s.progressRepo.GetAchievements(userID)
	if err != nil {
		return []domain.Achievement{}, nil
	}

	// Sort by unlock date, unlocked first
	sort.Slice(achievements, func(i, j int) bool {
		if achievements[i].IsUnlocked() && !achievements[j].IsUnlocked() {
			return true
		}
		if !achievements[i].IsUnlocked() && achievements[j].IsUnlocked() {
			return false
		}
		if achievements[i].IsUnlocked() && achievements[j].IsUnlocked() {
			return achievements[i].UnlockDate.Before(*achievements[j].UnlockDate)
		}
		return achievements[i].Progress > achievements[j].Progress
	})

	return achievements, nil
}

// UnlockAchievement manually unlocks an achievement
func (s *ProgressService) UnlockAchievement(userID, achievementID string) (*domain.Achievement, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if achievementID == "" {
		return nil, errors.New("achievement ID cannot be empty")
	}

	achievements, err := s.progressRepo.GetAchievements(userID)
	if err != nil {
		return nil, err
	}

	for i, a := range achievements {
		if a.ID == achievementID {
			a.Unlock()
			s.progressRepo.SaveAchievement(&a)
			return &achievements[i], nil
		}
	}

	return nil, errors.New("achievement not found")
}

// GetImprovements identifies areas where the user has improved
// Implements Requirement 7.8
func (s *ProgressService) GetImprovements(userID string) (map[string]bool, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	improvements := make(map[string]bool)

	// Get category progress
	categoryProgress, err := s.GetAllCategoryProgress(userID)
	if err != nil {
		return improvements, nil
	}

	// Check each category for improvement
	for cat, prog := range categoryProgress {
		if prog.CompletedExercises > 0 && prog.LastPracticed != nil {
			// Check if practiced recently (within last 7 days)
			sevenDaysAgo := time.Now().AddDate(0, 0, -7)
			if prog.LastPracticed.After(sevenDaysAgo) {
				improvements[string(cat)] = true
			}
		}
	}

	// Check streak improvement
	streak, err := s.progressRepo.GetStreak(userID)
	if err == nil && streak.CurrentStreak > 0 {
		improvements["streak"] = true
	}

	return improvements, nil
}

// getOrCreateAchievement gets or creates an achievement
func (s *ProgressService) getOrCreateAchievement(userID string, achievementID string, achievementType domain.AchievementType, target int) *domain.Achievement {
	achievements, err := s.progressRepo.GetAchievements(userID)
	if err == nil {
		for _, a := range achievements {
			if a.ID == achievementID {
				return &a
			}
		}
	}

	// Create new achievement based on ID
	var name, description, icon, condition string

	switch achievementID {
	case "first_practice":
		name = "First Steps"
		description = "Complete your first exercise"
		icon = "🎯"
		condition = "Complete 1 exercise"
	case "week_streak":
		name = "Week Warrior"
		description = "Practice for 7 consecutive days"
		icon = "🔥"
		condition = "Maintain a 7-day streak"
	case "month_streak":
		name = "Monthly Master"
		description = "Practice for 30 consecutive days"
		icon = "⭐"
		condition = "Maintain a 30-day streak"
	case "tongue_twister_master":
		name = "Tongue Twister Master"
		description = "Complete 100 tongue twisters"
		icon = "🗣️"
		condition = "Complete 100 tongue twisters"
	default:
		name = achievementID
		description = "Achievement"
		icon = "🏆"
		condition = "Complete the challenge"
	}

	return domain.NewAchievement(userID, name, description, icon, condition, achievementType, target)
}

// RecordProgress records progress for a completed exercise
func (s *ProgressService) RecordProgress(userID string, category domain.ExerciseCategory, exerciseCount int, duration time.Duration) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Create or update progress record
	existingProgress, err := s.progressRepo.GetProgress(userID, date)
	if err != nil {
		existingProgress = domain.NewProgressRecord(userID, category)
		existingProgress.Date = date
	}

	existingProgress.AddProgress(exerciseCount, duration)
	existingProgress.MarkCompleted()

	if err := s.progressRepo.SaveProgress(existingProgress); err != nil {
		return err
	}

	// Update category progress
	catProgress, err := s.progressRepo.GetCategoryProgress(userID, category)
	if err != nil {
		allExercises, _ := s.exerciseRepo.GetByCategory(category)
		catProgress = domain.NewCategoryProgress(userID, category, len(allExercises))
	}

	catProgress.AddExercise(duration)
	if err := s.progressRepo.SaveCategoryProgress(catProgress); err != nil {
		return err
	}

	// Update streak
	streak, err := s.progressRepo.GetStreak(userID)
	if err != nil {
		streak = domain.NewStreakRecord(userID)
	}
	streak.UpdateStreak()
	if err := s.progressRepo.SaveStreak(streak); err != nil {
		return err
	}

	// Check for new achievements
	s.CheckMilestones(userID)

	return nil
}

// GetActivityLevel returns the activity level for a specific day
func (s *ProgressService) GetActivityLevel(userID string, date time.Time) (int, error) {
	if userID == "" {
		return 0, errors.New("user ID cannot be empty")
	}

	progress, err := s.progressRepo.GetProgress(userID, date)
	if err != nil {
		return 0, nil
	}

	// Calculate activity level based on exercise count
	switch {
	case progress.ExerciseCount >= 10:
		return 4, nil
	case progress.ExerciseCount >= 7:
		return 3, nil
	case progress.ExerciseCount >= 4:
		return 2, nil
	case progress.ExerciseCount >= 1:
		return 1, nil
	default:
		return 0, nil
	}
}

// GetStreakRecord returns the full streak record
func (s *ProgressService) GetStreakRecord(userID string) (*domain.StreakRecord, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	return s.progressRepo.GetStreak(userID)
}