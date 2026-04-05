package domain

import (
	"time"
)

// ProgressRecord represents a single progress record
type ProgressRecord struct {
	ID            string           `json:"id"`
	UserID        string           `json:"user_id"`
	Date          time.Time        `json:"date"`
	Category      ExerciseCategory `json:"category"`
	ExerciseCount int              `json:"exercise_count"`
	Duration      time.Duration    `json:"duration"`
	Completed     bool             `json:"completed"`
	CreatedAt     time.Time        `json:"created_at"`
}

// StreakRecord represents streak data for a user
type StreakRecord struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	CurrentStreak int        `json:"current_streak"`
	LongestStreak int        `json:"longest_streak"`
	StreakStartDate time.Time `json:"streak_start_date"`
	LastActivityDate time.Time `json:"last_activity_date"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// AchievementType represents the type of achievement
type AchievementType string

const (
	AchievementStreak      AchievementType = "streak"
	AchievementCompletion  AchievementType = "completion"
	AchievementMilestone   AchievementType = "milestone"
	AchievementCategory    AchievementType = "category"
	AchievementSpecial     AchievementType = "special"
)

// Achievement represents a user achievement
type Achievement struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	AchievementType AchievementType  `json:"achievement_type"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Icon            string           `json:"icon"`
	UnlockCondition string           `json:"unlock_condition"`
	UnlockDate      *time.Time       `json:"unlock_date,omitempty"`
	Progress        int              `json:"progress"`
	Target          int              `json:"target"`
	CreatedAt       time.Time        `json:"created_at"`
}

// CategoryProgress represents progress within a specific category
type CategoryProgress struct {
	UserID             string           `json:"user_id"`
	Category           ExerciseCategory `json:"category"`
	TotalExercises     int              `json:"total_exercises"`
	CompletedExercises int              `json:"completed_exercises"`
	TotalTime          time.Duration    `json:"total_time"`
	AverageSessionLength time.Duration  `json:"average_session_length"`
	LastPracticed      *time.Time       `json:"last_practiced,omitempty"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// NewProgressRecord creates a new progress record
func NewProgressRecord(userID string, category ExerciseCategory) *ProgressRecord {
	now := time.Now()
	return &ProgressRecord{
		ID:            generateID(),
		UserID:        userID,
		Date:          now,
		Category:      category,
		ExerciseCount: 0,
		Duration:      0,
		Completed:     false,
		CreatedAt:     now,
	}
}

// NewStreakRecord creates a new streak record
func NewStreakRecord(userID string) *StreakRecord {
	now := time.Now()
	return &StreakRecord{
		ID:              generateID(),
		UserID:          userID,
		CurrentStreak:   0,
		LongestStreak:   0,
		StreakStartDate: now,
		LastActivityDate: now,
		UpdatedAt:       now,
	}
}

// NewAchievement creates a new achievement
func NewAchievement(userID, name, description, icon, unlockCondition string, achievementType AchievementType, target int) *Achievement {
	return &Achievement{
		ID:              generateID(),
		UserID:          userID,
		AchievementType: achievementType,
		Name:            name,
		Description:     description,
		Icon:            icon,
		UnlockCondition: unlockCondition,
		Progress:        0,
		Target:          target,
		CreatedAt:       time.Now(),
	}
}

// NewCategoryProgress creates a new category progress record
func NewCategoryProgress(userID string, category ExerciseCategory, totalExercises int) *CategoryProgress {
	return &CategoryProgress{
		UserID:             userID,
		Category:           category,
		TotalExercises:     totalExercises,
		CompletedExercises: 0,
		TotalTime:          0,
		AverageSessionLength: 0,
		UpdatedAt:          time.Now(),
	}
}

// UpdateStreak updates the streak record based on practice activity
func (s *StreakRecord) UpdateStreak() {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	lastActivity := time.Date(s.LastActivityDate.Year(), s.LastActivityDate.Month(), s.LastActivityDate.Day(), 0, 0, 0, 0, s.LastActivityDate.Location())
	
	daysDiff := int(today.Sub(lastActivity).Hours() / 24)
	
	switch {
	case daysDiff == 0:
		// Same day, no streak change
		return
	case daysDiff == 1:
		// Consecutive day
		s.CurrentStreak++
		if s.CurrentStreak > s.LongestStreak {
			s.LongestStreak = s.CurrentStreak
		}
		s.StreakStartDate = lastActivity
	default:
		// Streak broken
		s.CurrentStreak = 1
		s.StreakStartDate = today
	}
	
	s.LastActivityDate = now
	s.UpdatedAt = now
}

// BreakStreak resets the current streak
func (s *StreakRecord) BreakStreak() {
	s.CurrentStreak = 0
	s.UpdatedAt = time.Now()
}

// IsUnlocked returns true if the achievement is unlocked
func (a *Achievement) IsUnlocked() bool {
	return a.UnlockDate != nil
}

// Unlock marks the achievement as unlocked
func (a *Achievement) Unlock() {
	now := time.Now()
	a.UnlockDate = &now
	a.Progress = a.Target
}

// UpdateProgress updates the achievement progress
func (a *Achievement) UpdateProgress(progress int) {
	a.Progress = progress
	if a.Progress >= a.Target && !a.IsUnlocked() {
		a.Unlock()
	}
}

// GetCompletionPercentage returns the completion percentage
func (a *Achievement) GetCompletionPercentage() float64 {
	if a.Target == 0 {
		return 0
	}
	return float64(a.Progress) / float64(a.Target) * 100
}

// GetCompletionPercentage returns the completion percentage for category
func (c *CategoryProgress) GetCompletionPercentage() float64 {
	if c.TotalExercises == 0 {
		return 0
	}
	return float64(c.CompletedExercises) / float64(c.TotalExercises) * 100
}

// AddExercise adds an exercise completion to the category progress
func (c *CategoryProgress) AddExercise(duration time.Duration) {
	c.CompletedExercises++
	c.TotalTime += duration
	now := time.Now()
	c.LastPracticed = &now
	c.UpdatedAt = now
	
	// Update average session length
	if c.CompletedExercises > 0 {
		c.AverageSessionLength = c.TotalTime / time.Duration(c.CompletedExercises)
	}
}

// AddProgress adds exercise count and duration to the progress record
func (p *ProgressRecord) AddProgress(exerciseCount int, duration time.Duration) {
	p.ExerciseCount += exerciseCount
	p.Duration += duration
}

// MarkCompleted marks the progress record as completed for the day
func (p *ProgressRecord) MarkCompleted() {
	p.Completed = true
}

// GetProgressSummary returns a summary of all progress
type ProgressSummary struct {
	CurrentStreak       int                    `json:"current_streak"`
	LongestStreak       int                    `json:"longest_streak"`
	TotalSessions       int                    `json:"total_sessions"`
	TotalExercises      int                    `json:"total_exercises"`
	TotalPracticeTime   time.Duration          `json:"total_practice_time"`
	CategoryProgress    map[string]CategoryProgress `json:"category_progress"`
	Achievements        []Achievement          `json:"achievements"`
	WeeklyActivity      []DayActivity          `json:"weekly_activity"`
}

// DayActivity represents activity for a single day
type DayActivity struct {
	Date        time.Time `json:"date"`
	ExerciseCount int     `json:"exercise_count"`
	Duration    time.Duration `json:"duration"`
	Level       int      `json:"level"` // 0-4 representing activity level
}

// GetWeeklyCalendar generates weekly activity data
func GetWeeklyCalendar(records []ProgressRecord) []DayActivity {
	now := time.Now()
	calendar := make([]DayActivity, 7)
	
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)
		
		dayRecord := DayActivity{
			Date:         dayStart,
			ExerciseCount: 0,
			Duration:    0,
			Level:       0,
		}
		
		for _, r := range records {
			if r.Date.After(dayStart) && r.Date.Before(dayEnd) {
				dayRecord.ExerciseCount += r.ExerciseCount
				dayRecord.Duration += r.Duration
			}
		}
		
		// Calculate activity level
		switch {
		case dayRecord.ExerciseCount >= 10:
			dayRecord.Level = 4
		case dayRecord.ExerciseCount >= 7:
			dayRecord.Level = 3
		case dayRecord.ExerciseCount >= 4:
			dayRecord.Level = 2
		case dayRecord.ExerciseCount >= 1:
			dayRecord.Level = 1
		default:
			dayRecord.Level = 0
		}
		
		calendar[6-i] = dayRecord
	}
	
	return calendar
}

// Predefined achievements
var (
	AchievementFirstPractice = &Achievement{
		ID:               "first_practice",
		AchievementType: AchievementCompletion,
		Name:             "First Steps",
		Description:      "Complete your first exercise",
		Icon:             "🎯",
		UnlockCondition:  "Complete 1 exercise",
		Target:           1,
	}
	
	AchievementWeekStreak = &Achievement{
		ID:               "week_streak",
		AchievementType: AchievementStreak,
		Name:             "Week Warrior",
		Description:      "Practice for 7 consecutive days",
		Icon:             "🔥",
		UnlockCondition:  "Maintain a 7-day streak",
		Target:           7,
	}
	
	AchievementMonthStreak = &Achievement{
		ID:               "month_streak",
		AchievementType: AchievementStreak,
		Name:             "Monthly Master",
		Description:      "Practice for 30 consecutive days",
		Icon:             "⭐",
		UnlockCondition:  "Maintain a 30-day streak",
		Target:           30,
	}
	
	AchievementTongueTwisterMaster = &Achievement{
		ID:               "tongue_twister_master",
		AchievementType: AchievementCategory,
		Name:             "Tongue Twister Master",
		Description:      "Complete 100 tongue twisters",
		Icon:             "🗣️",
		UnlockCondition:  "Complete 100 tongue twisters",
		Target:           100,
	}
)

// generateID generates a unique ID (placeholder implementation)
func generateID() string {
	return time.Now().Format("20060102150405")
}

// RecommendationRecord represents a recommendation that was shown to a user
type RecommendationRecord struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ExerciseID    string    `json:"exercise_id"`
	CreatedAt     time.Time `json:"created_at"`
	Viewed        bool      `json:"viewed"`
	Clicked       bool      `json:"clicked"`
}

// RejectedRecommendation represents a recommendation that was rejected by a user
type RejectedRecommendation struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ExerciseID string    `json:"exercise_id"`
	RejectedAt time.Time `json:"rejected_at"`
}

// AcceptedRecommendation represents a recommendation that was accepted by a user
type AcceptedRecommendation struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ExerciseID string    `json:"exercise_id"`
	AcceptedAt time.Time `json:"accepted_at"`
}