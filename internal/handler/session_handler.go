package handler

import (
	"net/http"
	"strings"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/service"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	baseHandler    *BaseHandler
	sessionService *service.SessionService
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(base *BaseHandler, sessionSvc *service.SessionService) *SessionHandler {
	return &SessionHandler{
		baseHandler:    base,
		sessionService: sessionSvc,
	}
}

// SessionIndexData holds data for the session index page (Req 6.8)
type SessionIndexData struct {
	PageData
	IncompleteSessions []domain.PracticeSession
	HasIncomplete      bool
}

// SessionActiveData holds data for an active session page (Req 6.1)
type SessionActiveData struct {
	PageData
	Session   *domain.PracticeSession
	SessionID string
	StartTime time.Time
	Elapsed   time.Duration
}

// SessionRecoveryData holds data for the session recovery prompt (Req 13.3)
type SessionRecoveryData struct {
	PageData
	SessionID          string
	ExercisesCompleted int
	StartTime          time.Time
	Message            string
}

// SessionSummaryData holds data for the session summary page (Req 6.2, 6.6)
type SessionSummaryData struct {
	PageData
	Session          *domain.PracticeSession
	Stats            domain.SessionStats
	DurationMinutes  int
	DurationSeconds  int
	ExerciseCount    int
}

// Index handles the practice session page (Req 6.8)
// Shows incomplete sessions so the user can resume them.
func (h *SessionHandler) Index(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	incompleteSessions, _ := h.sessionService.GetIncompleteSessions(userID)

	data := SessionIndexData{
		PageData:           h.baseHandler.NewPageData(r, "Practice Session"),
		IncompleteSessions: incompleteSessions,
		HasIncomplete:      len(incompleteSessions) > 0,
	}

	h.baseHandler.Render(w, "session.html", data)
}

// Start handles starting a new session (Req 6.1)
// POST: creates a new session and redirects to the active session page.
// GET: renders the start page (which may show incomplete session prompt per Req 6.8).
func (h *SessionHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)

	if r.Method == http.MethodPost {
		session, err := h.sessionService.StartSession(userID)
		if err != nil {
			h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to start session")
			return
		}
		h.baseHandler.Redirect(w, r, "/session/resume/"+session.ID, http.StatusFound)
		return
	}

	// GET: show start page with any incomplete sessions (Req 6.8)
	incompleteSessions, _ := h.sessionService.GetIncompleteSessions(userID)

	data := SessionIndexData{
		PageData:           h.baseHandler.NewPageData(r, "Start Practice Session"),
		IncompleteSessions: incompleteSessions,
		HasIncomplete:      len(incompleteSessions) > 0,
	}

	h.baseHandler.Render(w, "session_start.html", data)
}

// CompleteExercise handles recording an exercise completion within a session (Req 6.3)
// POST: records the exercise and redirects back to the active session.
func (h *SessionHandler) CompleteExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.FormValue("session_id")
	exerciseID := r.FormValue("exercise_id")
	if sessionID == "" || exerciseID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Session ID and Exercise ID required")
		return
	}

	repetitions := 1
	score := 0
	notes := r.FormValue("notes")

	_, err := h.sessionService.AddExerciseToSession(sessionID, exerciseID, repetitions, score, notes)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to record exercise completion")
		return
	}

	h.baseHandler.Redirect(w, r, "/session/resume/"+sessionID, http.StatusFound)
}

// Complete handles completing a session and showing the summary (Req 6.2, 6.6)
// POST: completes the session and shows the summary.
func (h *SessionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var sessionID string

	if r.Method == http.MethodPost {
		sessionID = r.FormValue("session_id")
	} else {
		sessionID = r.URL.Query().Get("id")
	}

	if sessionID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Session ID required")
		return
	}

	session, err := h.sessionService.CompleteSession(sessionID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to complete session")
		return
	}

	stats := session.CalculateSessionStats()
	totalSecs := int(stats.TotalDuration.Seconds())

	data := SessionSummaryData{
		PageData:        h.baseHandler.NewPageData(r, "Session Complete"),
		Session:         session,
		Stats:           stats,
		DurationMinutes: totalSecs / 60,
		DurationSeconds: totalSecs % 60,
		ExerciseCount:   session.GetExerciseCount(),
	}

	h.baseHandler.Render(w, "session_complete.html", data)
}

// Save handles saving an incomplete session for later resumption (Req 6.7)
// POST: saves the session and redirects to the session index.
func (h *SessionHandler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.baseHandler.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.FormValue("session_id")
	if sessionID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Session ID required")
		return
	}

	_, err := h.sessionService.SaveSession(sessionID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to save session")
		return
	}

	h.baseHandler.Redirect(w, r, "/session", http.StatusFound)
}

// Resume handles resuming a saved session (Req 6.7, 6.8)
// Loads the session and renders the active session page.
func (h *SessionHandler) Resume(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimPrefix(r.URL.Path, "/session/resume/")
	if sessionID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Session ID required")
		return
	}

	session, err := h.sessionService.ResumeSession(sessionID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusNotFound, "Session not found")
		return
	}

	elapsed := time.Since(session.StartTime)

	data := SessionActiveData{
		PageData:  h.baseHandler.NewPageData(r, "Session in Progress"),
		Session:   session,
		SessionID: session.ID,
		StartTime: session.StartTime,
		Elapsed:   elapsed,
	}

	h.baseHandler.Render(w, "session_active.html", data)
}
