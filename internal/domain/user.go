package domain

import (
	"time"
)

// UserProfile represents the user's overall profile
type UserProfile struct {
	ID                  string           `json:"id"`
	Username            string           `json:"username"`
	CreatedAt           time.Time        `json:"created_at"`
	LastPracticeDate    *time.Time       `json:"last_practice_date,omitempty"`
	CurrentStreak       int              `json:"current_streak"`
	LongestStreak       int              `json:"longest_streak"`
	TotalPracticeTime   time.Duration    `json:"total_practice_time"`
	TotalExercisesCompleted int           `json:"total_exercises_completed"`
	AchievementStatus   []string         `json:"achievement_status"`
	PreferencesID       string           `json:"preferences_id"`
	Preferences         *UserPreferences `json:"preferences,omitempty"`
}

// NewUserProfile creates a new user profile
func NewUserProfile(id, username string) *UserProfile {
	return &UserProfile{
		ID:                    id,
		Username:              username,
		CreatedAt:             time.Now(),
		CurrentStreak:         0,
		LongestStreak:         0,
		TotalPracticeTime:     0,
		TotalExercisesCompleted: 0,
		AchievementStatus:     []string{},
	}
}

// UpdateStreak updates the user's streak based on practice activity
func (u *UserProfile) UpdateStreak() {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	if u.LastPracticeDate == nil {
		// First practice
		u.CurrentStreak = 1
		u.LongestStreak = 1
		u.LastPracticeDate = &now
		return
	}
	
	lastPractice := *u.LastPracticeDate
	lastPracticeDay := time.Date(lastPractice.Year(), lastPractice.Month(), lastPractice.Day(), 0, 0, 0, 0, lastPractice.Location())
	
	daysDiff := int(today.Sub(lastPracticeDay).Hours() / 24)
	
	switch {
	case daysDiff == 0:
		// Same day, no streak change
		return
	case daysDiff == 1:
		// Consecutive day
		u.CurrentStreak++
		if u.CurrentStreak > u.LongestStreak {
			u.LongestStreak = u.CurrentStreak
		}
	default:
		// Streak broken
		u.CurrentStreak = 1
	}
	
	u.LastPracticeDate = &now
}

// AddPracticeTime adds practice time to the user's total
func (u *UserProfile) AddPracticeTime(duration time.Duration) {
	u.TotalPracticeTime += duration
}

// IncrementExercises increments the total exercises completed
func (u *UserProfile) IncrementExercises() {
	u.TotalExercisesCompleted++
}

// HasAchievement checks if the user has a specific achievement
func (u *UserProfile) HasAchievement(achievementID string) bool {
	for _, a := range u.AchievementStatus {
		if a == achievementID {
			return true
		}
	}
	return false
}

// AddAchievement adds an achievement to the user's profile
func (u *UserProfile) AddAchievement(achievementID string) {
	if !u.HasAchievement(achievementID) {
		u.AchievementStatus = append(u.AchievementStatus, achievementID)
	}
}

// Validate checks if the user profile has valid data
func (u *UserProfile) Validate() error {
	if u.ID == "" {
		return &ValidationError{"user ID cannot be empty"}
	}
	if u.Username == "" {
		return &ValidationError{"username cannot be empty"}
	}
	return nil
}

// GetStreakDays returns the current streak in days
func (u *UserProfile) GetStreakDays() int {
	return u.CurrentStreak
}

// GetTotalPracticeMinutes returns total practice time in minutes
func (u *UserProfile) GetTotalPracticeMinutes() int {
	return int(u.TotalPracticeTime.Minutes())
}