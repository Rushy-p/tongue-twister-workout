package handler

import (
	"net/http"

	"speech-practice-app/internal/service"
)

// Router handles HTTP routing
type Router struct {
	mux              *http.ServeMux
	baseHandler      *BaseHandler
	exerciseHandler  *ExerciseHandler
	sessionHandler   *SessionHandler
	progressHandler  *ProgressHandler
	preferencesHandler *PreferencesHandler
}

// NewRouter creates a new Router with all handlers
func NewRouter(
	exerciseService *service.ExerciseService,
	sessionService *service.SessionService,
	progressService *service.ProgressService,
	preferencesService *service.PreferencesService,
) *Router {
	// Create base handler
	baseHandler := NewBaseHandler(
		exerciseService,
		sessionService,
		progressService,
		preferencesService,
	)

	// Create specialized handlers
	exerciseHandler := NewExerciseHandler(baseHandler, exerciseService)
	sessionHandler := NewSessionHandler(baseHandler, sessionService)
	progressHandler := NewProgressHandler(baseHandler, progressService)
	preferencesHandler := NewPreferencesHandler(baseHandler, preferencesService)

	// Create router
	r := &Router{
		mux:                http.NewServeMux(),
		baseHandler:        baseHandler,
		exerciseHandler:    exerciseHandler,
		sessionHandler:     sessionHandler,
		progressHandler:    progressHandler,
		preferencesHandler: preferencesHandler,
	}

	// Register routes
	r.registerRoutes()

	return r
}

// registerRoutes registers all HTTP routes
func (r *Router) registerRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("static"))
	r.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home page
	r.mux.HandleFunc("/", r.exerciseHandler.Index)

	// Exercise routes
	r.mux.HandleFunc("/exercises", r.exerciseHandler.List)
	r.mux.HandleFunc("/exercises/", r.exerciseHandler.Detail)

	// Category routes
	r.mux.HandleFunc("/exercises/mouth", r.exerciseHandler.Category)
	r.mux.HandleFunc("/exercises/twisters", r.exerciseHandler.Category)
	r.mux.HandleFunc("/exercises/diction", r.exerciseHandler.Category)
	r.mux.HandleFunc("/exercises/pacing", r.exerciseHandler.Category)

	// Session routes
	r.mux.HandleFunc("/session", r.sessionHandler.Index)
	r.mux.HandleFunc("/session/start", r.sessionHandler.Start)
	r.mux.HandleFunc("/session/complete", r.sessionHandler.Complete)
	r.mux.HandleFunc("/session/exercise/complete", r.sessionHandler.CompleteExercise)
	r.mux.HandleFunc("/session/save", r.sessionHandler.Save)
	r.mux.HandleFunc("/session/resume/", r.sessionHandler.Resume)

	// Progress routes
	r.mux.HandleFunc("/progress", r.progressHandler.Index)
	r.mux.HandleFunc("/progress/streak", r.progressHandler.Streak)
	r.mux.HandleFunc("/progress/calendar", r.progressHandler.WeeklyCalendar)

	// Preferences routes
	r.mux.HandleFunc("/preferences", r.preferencesHandler.Index)
	r.mux.HandleFunc("/preferences/update", r.preferencesHandler.Update)
	r.mux.HandleFunc("/preferences/export", r.preferencesHandler.Export)

	// Recommendations routes
	r.mux.HandleFunc("/recommendations", r.exerciseHandler.Recommendations)
}

// Handler returns the HTTP handler
func (r *Router) Handler() http.Handler {
	// Apply middleware chain
	handler := Chain(
		r.mux,
		LoggingMiddleware,
		RecoveryMiddleware,
		CORSMiddleware,
		SecurityHeadersMiddleware,
	)
	return handler
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	r.Handler().ServeHTTP(w, rq)
}