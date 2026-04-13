package handler

import (
	"net/http"

	"speech-practice-app/internal/service"
)

// ProgressHandler handles progress-related HTTP requests
type ProgressHandler struct {
	baseHandler    *BaseHandler
	progressService *service.ProgressService
}

// NewProgressHandler creates a new ProgressHandler
func NewProgressHandler(base *BaseHandler, progressSvc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{
		baseHandler:     base,
		progressService: progressSvc,
	}
}

// Index handles the progress page
func (h *ProgressHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	streak, _ := h.progressService.GetCurrentStreak(userID)
	longestStreak, _ := h.progressService.GetLongestStreak(userID)
	totalExercises, _ := h.progressService.GetTotalExercises(userID)
	totalTime, _ := h.progressService.GetTotalPracticeTime(userID)

	data := struct {
		PageData
		CurrentStreak   int
		LongestStreak   int
		TotalExercises  int
		TotalTime       string
	}{
		PageData:        h.baseHandler.NewPageData(r, "Your Progress"),
		CurrentStreak:   streak,
		LongestStreak:   longestStreak,
		TotalExercises:  totalExercises,
		TotalTime:       totalTime.String(),
	}

	h.baseHandler.Render(w, "progress.html", data)
}