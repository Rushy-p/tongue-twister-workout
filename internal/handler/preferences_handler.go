package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/service"
)

// PreferencesHandler handles preferences-related HTTP requests
type PreferencesHandler struct {
	baseHandler        *BaseHandler
	preferencesService *service.PreferencesService
}

// NewPreferencesHandler creates a new PreferencesHandler
func NewPreferencesHandler(base *BaseHandler, prefsSvc *service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{
		baseHandler:        base,
		preferencesService: prefsSvc,
	}
}

// PreferencesPageData holds data for the preferences page (Req 9.1–9.6)
type PreferencesPageData struct {
	PageData
	Preferences *domain.UserPreferences
	Success     string
	Error       string
}

// Index handles the preferences display page (Req 9.1, 9.2, 9.3, 9.4, 9.6)
func (h *PreferencesHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	prefs, err := h.preferencesService.GetPreferences(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	success := r.URL.Query().Get("success")

	data := PreferencesPageData{
		PageData:    h.baseHandler.NewPageData(r, "Settings"),
		Preferences: prefs,
		Success:     success,
	}

	h.baseHandler.Render(w, "preferences.html", data)
}

// Update handles updating user preferences (Req 9.1, 9.2, 9.3, 9.4, 9.5)
func (h *PreferencesHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.baseHandler.Error(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	userID := h.baseHandler.GetUserID(r)
	updates := make(map[string]interface{})

	// Difficulty (Req 9.1)
	if difficulty := r.FormValue("difficulty"); difficulty != "" {
		updates["difficulty"] = difficulty
	}

	// Default duration in seconds (Req 9.2)
	if durationStr := r.FormValue("default_duration"); durationStr != "" {
		if secs, err := strconv.ParseFloat(durationStr, 64); err == nil {
			updates["default_duration"] = secs
		}
	}

	// Audio feedback (Req 9.3)
	updates["audio_feedback"] = r.FormValue("audio_feedback") == "on"

	// Vibration feedback (Req 9.4)
	updates["vibration_feedback"] = r.FormValue("vibration_feedback") == "on"

	// Reminder settings (Req 8.1–8.4)
	updates["reminder_enabled"] = r.FormValue("reminder_enabled") == "on"
	if reminderTime := r.FormValue("reminder_time"); reminderTime != "" {
		updates["reminder_time"] = reminderTime
	}

	// Export format (Req 9.7)
	if exportFormat := r.FormValue("export_format"); exportFormat != "" {
		updates["export_format"] = exportFormat
	}

	_, err := h.preferencesService.UpdatePreferences(userID, updates)
	if err != nil {
		prefs, _ := h.preferencesService.GetPreferences(userID)
		if prefs == nil {
			prefs = domain.NewUserPreferences(userID)
		}
		data := PreferencesPageData{
			PageData:    h.baseHandler.NewPageData(r, "Settings"),
			Preferences: prefs,
			Error:       fmt.Sprintf("Failed to save preferences: %v", err),
		}
		h.baseHandler.Render(w, "preferences.html", data)
		return
	}

	h.baseHandler.Redirect(w, r, "/preferences?success=saved", http.StatusFound)
}

// Export handles exporting user data (Req 9.7)
func (h *PreferencesHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	prefs, err := h.preferencesService.GetPreferences(userID)
	if err != nil {
		prefs = domain.NewUserPreferences(userID)
	}

	format := domain.ExportFormat(r.URL.Query().Get("format"))
	if format == "" {
		format = prefs.ExportFormat
	}
	if format != domain.ExportFormatJSON && format != domain.ExportFormatCSV {
		format = domain.ExportFormatJSON
	}

	data, err := h.preferencesService.ExportData(userID, format)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to export data")
		return
	}

	switch format {
	case domain.ExportFormatCSV:
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=\"speech-practice-data.csv\"")
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"speech-practice-data.json\"")
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, data)
}
