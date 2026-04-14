package handler

import (
	"net/http"

	"speech-practice-app/internal/service"
)

// RecommendationsHandler handles recommendation-related HTTP requests
type RecommendationsHandler struct {
	baseHandler            *BaseHandler
	recommendationService  *service.RecommendationService
}

// NewRecommendationsHandler creates a new RecommendationsHandler
func NewRecommendationsHandler(base *BaseHandler, recSvc *service.RecommendationService) *RecommendationsHandler {
	return &RecommendationsHandler{
		baseHandler:           base,
		recommendationService: recSvc,
	}
}

// RecommendationsPageData holds data for the recommendations page (Req 10.4)
type RecommendationsPageData struct {
	PageData
	Summary *service.RecommendationSummary
	Error   string
}

// Index handles the daily recommendations page (Req 10.4)
func (h *RecommendationsHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	summary, err := h.recommendationService.GetDailyRecommendationSummary(userID)
	if err != nil {
		data := RecommendationsPageData{
			PageData: h.baseHandler.NewPageData(r, "Recommendations"),
			Error:    "Unable to load recommendations. Please try again later.",
		}
		h.baseHandler.Render(w, "recommendations.html", data)
		return
	}

	data := RecommendationsPageData{
		PageData: h.baseHandler.NewPageData(r, "Recommendations"),
		Summary:  summary,
	}

	h.baseHandler.Render(w, "recommendations.html", data)
}

// Accept handles accepting a recommendation (Req 10.6)
// POST /recommendations/accept?exercise_id=<id>
func (h *RecommendationsHandler) Accept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.baseHandler.Error(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	exerciseID := r.FormValue("exercise_id")
	if exerciseID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Exercise ID required")
		return
	}

	userID := h.baseHandler.GetUserID(r)

	if err := h.recommendationService.AcceptRecommendation(userID, exerciseID); err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to accept recommendation")
		return
	}

	// Redirect to the exercise detail page so the user can start practising
	h.baseHandler.Redirect(w, r, "/exercises/"+exerciseID, http.StatusFound)
}

// Reject handles rejecting a recommendation (Req 10.6)
// POST /recommendations/reject?exercise_id=<id>
func (h *RecommendationsHandler) Reject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.baseHandler.Error(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	exerciseID := r.FormValue("exercise_id")
	if exerciseID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Exercise ID required")
		return
	}

	userID := h.baseHandler.GetUserID(r)

	if err := h.recommendationService.RejectRecommendation(userID, exerciseID); err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to reject recommendation")
		return
	}

	// Redirect back to recommendations so the user sees a refreshed list
	h.baseHandler.Redirect(w, r, "/recommendations", http.StatusFound)
}
