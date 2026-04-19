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

// Render renders a named page template wrapped in the base layout.
// Each page template must define a block named after its file (without .html),
// e.g. templates/exercises.html defines {{define "exercises"}}.
// The base layout injects it via {{template "content" .}} after we clone
// the template set and add a "content" alias pointing to the right block.
func (h *BaseHandler) Render(w http.ResponseWriter, name string, data interface{}) {
	// Derive the template block name from the filename (strip .html suffix)
	blockName := name
	if len(blockName) > 5 && blockName[len(blockName)-5:] == ".html" {
		blockName = blockName[:len(blockName)-5]
	}

	// Clone the template set so we can add a "content" alias without mutating
	// the shared set (safe for concurrent requests).
	cloned, err := h.templates.Clone()
	if err != nil {
		log.Printf("Template clone error (%s): %v", name, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Add a "content" template that delegates to the page-specific block.
	alias := `{{define "content"}}{{template "` + blockName + `" .}}{{end}}`
	cloned, err = cloned.New("content_alias").Parse(alias)
	if err != nil {
		log.Printf("Template alias error (%s): %v", name, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := cloned.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Template error (%s): %v", name, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// RenderString renders a raw string as HTML response
func (h *BaseHandler) RenderString(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// ErrorPageData holds data for error page templates (Req 13.1, 13.5)
type ErrorPageData struct {
	PageData
	StatusCode int
	StatusText string
	Message    string
	RetryURL   string // optional — URL to retry the failed action
	HomeURL    string // always "/"
}

// Error sends an HTML error response using the error.html template.
// Falls back to plain http.Error if template rendering fails (Req 13.5).
func (h *BaseHandler) Error(w http.ResponseWriter, code int, message string) {
	data := ErrorPageData{
		PageData: PageData{
			Title: http.StatusText(code),
		},
		StatusCode: code,
		StatusText: http.StatusText(code),
		Message:    message,
		HomeURL:    "/",
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	cloned, err := h.templates.Clone()
	if err != nil {
		http.Error(w, message, code)
		return
	}
	alias := `{{define "content"}}{{template "error" .}}{{end}}`
	cloned, err = cloned.New("content_alias").Parse(alias)
	if err != nil {
		http.Error(w, message, code)
		return
	}
	if err := cloned.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error template render failed (%d): %v", code, err)
		http.Error(w, message, code)
	}
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