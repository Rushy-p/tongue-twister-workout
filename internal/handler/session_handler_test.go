package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"speech-practice-app/internal/infrastructure"
	"speech-practice-app/internal/service"
)

func newTestSessionHandler(t *testing.T) (*SessionHandler, *service.SessionService) {
	t.Helper()
	changeToProjectRoot(t)

	exerciseRepo := infrastructure.NewInMemoryExerciseRepository()
	progressRepo := infrastructure.NewInMemoryProgressRepository()
	sessionRepo := infrastructure.NewInMemorySessionRepository()

	exerciseSvc := service.NewExerciseService(exerciseRepo, progressRepo, sessionRepo)
	sessionSvc := service.NewSessionService(sessionRepo, progressRepo, exerciseRepo)

	base := NewBaseHandler(exerciseSvc, sessionSvc, nil, nil)
	return NewSessionHandler(base, sessionSvc), sessionSvc
}

// TestSessionIndex_ReturnsOK verifies the session index page loads (Req 6.8)
func TestSessionIndex_ReturnsOK(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// TestSessionStart_GET_ReturnsOK verifies the start page renders (Req 6.1)
func TestSessionStart_GET_ReturnsOK(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/start", nil)
	rr := httptest.NewRecorder()
	h.Start(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// TestSessionStart_POST_RedirectsToActiveSession verifies starting a session redirects (Req 6.1)
func TestSessionStart_POST_RedirectsToActiveSession(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/session/start", nil)
	rr := httptest.NewRecorder()
	h.Start(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/session/resume/") {
		t.Errorf("expected redirect to /session/resume/..., got %q", loc)
	}
}

// TestSessionResume_ValidID_ReturnsOK verifies resuming a session works (Req 6.7, 6.8)
func TestSessionResume_ValidID_ReturnsOK(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/session/resume/"+session.ID, nil)
	rr := httptest.NewRecorder()
	h.Resume(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// TestSessionResume_InvalidID_Returns404 verifies missing session returns 404
func TestSessionResume_InvalidID_Returns404(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/resume/nonexistent-id", nil)
	rr := httptest.NewRecorder()
	h.Resume(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// TestSessionResume_EmptyID_ReturnsBadRequest verifies empty session ID returns 400
func TestSessionResume_EmptyID_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/resume/", nil)
	rr := httptest.NewRecorder()
	h.Resume(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// TestSessionSave_POST_RedirectsToIndex verifies saving a session (Req 6.7)
func TestSessionSave_POST_RedirectsToIndex(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	form := url.Values{"session_id": {session.ID}}
	req := httptest.NewRequest(http.MethodPost, "/session/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/session" {
		t.Errorf("expected redirect to /session, got %q", rr.Header().Get("Location"))
	}
}

// TestSessionSave_GET_MethodNotAllowed verifies GET is rejected for save
func TestSessionSave_GET_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/save", nil)
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// TestSessionSave_MissingID_ReturnsBadRequest verifies missing session ID returns 400
func TestSessionSave_MissingID_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/session/save", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Save(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// TestSessionComplete_POST_ShowsSummary verifies session completion shows summary (Req 6.2, 6.6)
func TestSessionComplete_POST_ShowsSummary(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	form := url.Values{"session_id": {session.ID}}
	req := httptest.NewRequest(http.MethodPost, "/session/complete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Complete(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// TestSessionComplete_GET_ShowsSummary verifies GET with query param also works
func TestSessionComplete_GET_ShowsSummary(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/session/complete?id="+session.ID, nil)
	rr := httptest.NewRecorder()
	h.Complete(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

// TestSessionComplete_MissingID_ReturnsBadRequest verifies missing session ID returns 400
func TestSessionComplete_MissingID_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/complete", nil)
	rr := httptest.NewRecorder()
	h.Complete(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// TestCompleteExercise_POST_RecordsAndRedirects verifies exercise completion recording (Req 6.3)
func TestCompleteExercise_POST_RecordsAndRedirects(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	form := url.Values{
		"session_id":  {session.ID},
		"exercise_id": {"me-001"},
	}
	req := httptest.NewRequest(http.MethodPost, "/session/exercise/complete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.CompleteExercise(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "/session/resume/") {
		t.Errorf("expected redirect to /session/resume/..., got %q", loc)
	}
}

// TestCompleteExercise_GET_MethodNotAllowed verifies GET is rejected
func TestCompleteExercise_GET_MethodNotAllowed(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/session/exercise/complete", nil)
	rr := httptest.NewRecorder()
	h.CompleteExercise(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// TestCompleteExercise_MissingIDs_ReturnsBadRequest verifies missing IDs return 400
func TestCompleteExercise_MissingIDs_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestSessionHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/session/exercise/complete", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.CompleteExercise(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// TestSessionIndex_ShowsIncompleteSessions verifies incomplete sessions are returned in data (Req 6.8)
func TestSessionIndex_ShowsIncompleteSessions(t *testing.T) {
	h, svc := newTestSessionHandler(t)

	// Create and save an incomplete session
	session, err := svc.StartSession("default-user")
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}
	_, err = svc.SaveSession(session.ID)
	if err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	// Verify the service returns the incomplete session
	incomplete, err := svc.GetIncompleteSessions("default-user")
	if err != nil {
		t.Fatalf("failed to get incomplete sessions: %v", err)
	}
	if len(incomplete) == 0 {
		t.Error("expected at least one incomplete session")
	}

	// Verify the handler returns 200 (data is passed to template)
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
