package handler

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"speech-practice-app/internal/service"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	exerciseService  *service.ExerciseService
	sessionService   *service.SessionService
	progressService  *service.ProgressService
	preferencesService *service.PreferencesService
	templates        *template.Template
}

// NewBaseHandler creates a new BaseHandler with all services
func NewBaseHandler(
	exerciseService *service.ExerciseService,
	sessionService *service.SessionService,
	progressService *service.ProgressService,
	preferencesService *service.PreferencesService,
) *BaseHandler {
	// Load templates
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
		templates = template.New("default")
	}

	return &BaseHandler{
		exerciseService:    exerciseService,
		sessionService:     sessionService,
		progressService:    progressService,
		preferencesService: preferencesService,
		templates:          templates,
	}
}

// Render renders a template with the given name and data
func (h *BaseHandler) Render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// RenderString renders a raw string as HTML response
func (h *BaseHandler) RenderString(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// Error sends an error response
func (h *BaseHandler) Error(w http.ResponseWriter, code int, message string) {
	http.Error(w, message, code)
}

// Redirect redirects to a given URL
func (h *BaseHandler) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

// GetUserID gets the user ID from the request (simplified - would use session in production)
func (h *BaseHandler) GetUserID(r *http.Request) string {
	// For now, use a default user ID
	// In production, this would come from session/auth
	return "default-user"
}

// GetStartTime extracts the start time from request context for performance tracking
func GetStartTime(r *http.Request) time.Time {
	if start, ok := r.Context().Value("startTime").(time.Time); ok {
		return start
	}
	return time.Now()
}

// Common page data
type PageData struct {
	Title       string
	UserID      string
	StartTime   time.Time
	RequestTime time.Duration
}

// NewPageData creates new page data with timing info
func (h *BaseHandler) NewPageData(r *http.Request, title string) PageData {
	return PageData{
		Title:       title,
		UserID:      h.GetUserID(r),
		StartTime:   GetStartTime(r),
		RequestTime: time.Since(GetStartTime(r)),
	}
}