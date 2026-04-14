package handler

import (
	"fmt"
	"net/http"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/service"
)

// ProgressHandler handles progress-related HTTP requests
type ProgressHandler struct {
	baseHandler     *BaseHandler
	progressService *service.ProgressService
}

// NewProgressHandler creates a new ProgressHandler
func NewProgressHandler(base *BaseHandler, progressSvc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{
		baseHandler:     base,
		progressService: progressSvc,
	}
}

// ProgressDashboardData holds data for the main progress page (Req 7.2, 7.3, 7.4, 7.5)
type ProgressDashboardData struct {
	PageData
	CurrentStreak     int
	LongestStreak     int
	TotalExercises    int
	TotalTimeMinutes  int
	TotalTimeSecs     int
	TotalSessions     int
	CategoryProgress  map[string]domain.CategoryProgress
	WeeklyActivity    []domain.DayActivity
	Achievements      []domain.Achievement
}

// StreakData holds data for the streak display (Req 7.2)
type StreakData struct {
	PageData
	CurrentStreak    int
	LongestStreak    int
	StreakStartDate  string
	LastActivityDate string
	Achievements     []domain.Achievement
}

// WeeklyCalendarData holds data for the weekly calendar display (Req 7.6)
type WeeklyCalendarData struct {
	PageData
	WeeklyActivity []domain.DayActivity
	TotalThisWeek  int
	ActiveDays     int
}

// Index handles the main progress dashboard page (Req 7.2, 7.3, 7.4, 7.5)
func (h *ProgressHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	summary, err := h.progressService.GetProgressSummary(userID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load progress data")
		return
	}

	totalSecs := int(summary.TotalPracticeTime.Seconds())

	data := ProgressDashboardData{
		PageData:         h.baseHandler.NewPageData(r, "Your Progress"),
		CurrentStreak:    summary.CurrentStreak,
		LongestStreak:    summary.LongestStreak,
		TotalExercises:   summary.TotalExercises,
		TotalTimeMinutes: totalSecs / 60,
		TotalTimeSecs:    totalSecs % 60,
		TotalSessions:    summary.TotalSessions,
		CategoryProgress: summary.CategoryProgress,
		WeeklyActivity:   summary.WeeklyActivity,
		Achievements:     summary.Achievements,
	}

	h.baseHandler.Render(w, "progress.html", data)
}

// Streak handles the streak display page (Req 7.2)
func (h *ProgressHandler) Streak(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	streakRecord, err := h.progressService.GetStreakRecord(userID)

	var currentStreak, longestStreak int
	var streakStart, lastActivity string

	if err == nil && streakRecord != nil {
		currentStreak = streakRecord.CurrentStreak
		longestStreak = streakRecord.LongestStreak
		streakStart = streakRecord.StreakStartDate.Format("Jan 2, 2006")
		lastActivity = streakRecord.LastActivityDate.Format("Jan 2, 2006")
	} else {
		currentStreak, _ = h.progressService.GetCurrentStreak(userID)
		longestStreak, _ = h.progressService.GetLongestStreak(userID)
		streakStart = time.Now().Format("Jan 2, 2006")
		lastActivity = time.Now().Format("Jan 2, 2006")
	}

	achievements, _ := h.progressService.GetAchievements(userID)

	data := StreakData{
		PageData:         h.baseHandler.NewPageData(r, "Practice Streak"),
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
		StreakStartDate:  streakStart,
		LastActivityDate: lastActivity,
		Achievements:     achievements,
	}

	h.baseHandler.Render(w, "progress_streak.html", data)
}

// WeeklyCalendar handles the weekly calendar display page (Req 7.6)
func (h *ProgressHandler) WeeklyCalendar(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	weeklyActivity, err := h.progressService.GetWeeklyCalendar(userID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load weekly calendar")
		return
	}

	totalThisWeek := 0
	activeDays := 0
	for _, day := range weeklyActivity {
		totalThisWeek += day.ExerciseCount
		if day.ExerciseCount > 0 {
			activeDays++
		}
	}

	data := WeeklyCalendarData{
		PageData:       h.baseHandler.NewPageData(r, "Weekly Activity"),
		WeeklyActivity: weeklyActivity,
		TotalThisWeek:  totalThisWeek,
		ActiveDays:     activeDays,
	}

	h.baseHandler.Render(w, "progress_calendar.html", data)
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
