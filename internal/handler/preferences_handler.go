package handler

import (
	"net/http"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/service"
)

// PreferencesHandler handles preferences-related HTTP requests
type PreferencesHandler struct {
	baseHandler       *BaseHandler
	preferencesService *service.PreferencesService
}

// NewPreferencesHandler creates a new PreferencesHandler
func NewPreferencesHandler(base *BaseHandler, prefsSvc *service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{
		baseHandler:       base,
		preferencesService: prefsSvc,
	}
}

// Index handles the preferences page
func (h *PreferencesHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	prefs, err := h.preferencesService.GetPreferences(userID)
	if err != nil {
		// Use defaults if preferences not found
		prefs = &domain.UserPreferences{
			Difficulty:      domain.DifficultyBeginner,
			DefaultDuration: 60,
		}
	}

	data := struct {
		PageData
		Preferences *domain.UserPreferences
	}{
		PageData:     h.baseHandler.NewPageData(r, "Settings"),
		Preferences: prefs,
	}

	h.baseHandler.Render(w, "preferences.html", data)
}

// Update handles updating preferences
func (h *PreferencesHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := h.baseHandler.GetUserID(r)

	updates := make(map[string]interface{})
	
	if difficulty := r.FormValue("difficulty"); difficulty != "" {
		updates["difficulty"] = domain.DifficultyLevel(difficulty)
	}
	if duration := r.FormValue("duration"); duration != "" {
		updates["default_duration"] = duration
	}
	if audio := r.FormValue("audio"); audio != "" {
		updates["audio_enabled"] = audio == "on"
	}

	_, err := h.preferencesService.UpdatePreferences(userID, updates)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to update preferences")
		return
	}

	h.baseHandler.Redirect(w, r, "/preferences", http.StatusFound)
}