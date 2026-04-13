package handler

import (
	"net/http"
	"strings"

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

// Index handles the practice session page
func (h *SessionHandler) Index(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageData
		HasActiveSession bool
	}{
		PageData:         h.baseHandler.NewPageData(r, "Practice Session"),
		HasActiveSession: false,
	}

	h.baseHandler.Render(w, "session.html", data)
}

// Start handles starting a new session
func (h *SessionHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := h.baseHandler.GetUserID(r)
	
	session, err := h.sessionService.StartSession(userID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to start session")
		return
	}

	data := struct {
		PageData
		SessionID string
	}{
		PageData:  h.baseHandler.NewPageData(r, "Session in Progress"),
		SessionID: session.ID,
	}

	h.baseHandler.Render(w, "session_active.html", data)
}

// Complete handles completing a session
func (h *SessionHandler) Complete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		h.baseHandler.Error(w, http.StatusBadRequest, "Session ID required")
		return
	}

	session, err := h.sessionService.CompleteSession(sessionID)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to complete session")
		return
	}

	data := struct {
		PageData
		Session interface{}
	}{
		PageData: h.baseHandler.NewPageData(r, "Session Complete"),
		Session:  session,
	}

	h.baseHandler.Render(w, "session_complete.html", data)
}

// Save handles saving a session
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

// Resume handles resuming a saved session
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

	data := struct {
		PageData
		Session interface{}
	}{
		PageData: h.baseHandler.NewPageData(r, "Resume Session"),
		Session:  session,
	}

	h.baseHandler.Render(w, "session_active.html", data)
}