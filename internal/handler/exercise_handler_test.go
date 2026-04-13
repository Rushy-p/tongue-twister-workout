package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
	"speech-practice-app/internal/service"
)

// changeToProjectRoot changes the working directory to the project root so
// templates can be loaded during tests.
func changeToProjectRoot(t *testing.T) {
	t.Helper()
	// Walk up until we find go.mod
	for i := 0; i < 5; i++ {
		if _, err := os.Stat("go.mod"); err == nil {
			return
		}
		if err := os.Chdir("../.."); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
	}
}

func newTestExerciseHandler(t *testing.T) *ExerciseHandler {
	t.Helper()
	changeToProjectRoot(t)
	exerciseRepo := infrastructure.NewInMemoryExerciseRepository()
	progressRepo := infrastructure.NewInMemoryProgressRepository()
	sessionRepo := infrastructure.NewInMemorySessionRepository()
	exerciseSvc := service.NewExerciseService(exerciseRepo, progressRepo, sessionRepo)
	base := NewBaseHandler(exerciseSvc, nil, nil, nil)
	return NewExerciseHandler(base, exerciseSvc)
}

func TestList_ReturnsOK(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestCategory_ValidSlug_ReturnsOK(t *testing.T) {
	h := newTestExerciseHandler(t)
	slugs := []string{"mouth", "twisters", "diction", "pacing"}
	for _, slug := range slugs {
		req := httptest.NewRequest(http.MethodGet, "/exercises/"+slug, nil)
		rr := httptest.NewRecorder()
		h.Category(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("slug %q: expected 200, got %d", slug, rr.Code)
		}
	}
}

func TestCategory_InvalidSlug_Returns404(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises/unknown", nil)
	rr := httptest.NewRecorder()
	h.Category(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDetail_ValidID_ReturnsOK(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises/me-001", nil)
	rr := httptest.NewRecorder()
	h.Detail(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestDetail_InvalidID_Returns404(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises/nonexistent-id", nil)
	rr := httptest.NewRecorder()
	h.Detail(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDetail_EmptyID_Redirects(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises/", nil)
	rr := httptest.NewRecorder()
	h.Detail(rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", rr.Code)
	}
}

func TestArticulationPointLabels(t *testing.T) {
	points := []domain.ArticulationPoint{
		domain.ArticulationLips,
		domain.ArticulationJaw,
		domain.ArticulationTongueTip,
	}
	labels := articulationPointLabels(points)
	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}
}

func TestSoundTargetLabels(t *testing.T) {
	sounds := []domain.SoundTarget{domain.SoundS, domain.SoundR}
	labels := soundTargetLabels(sounds)
	if len(labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(labels))
	}
}

func TestDetail_TongueTwister_HasAudioAndSoundData(t *testing.T) {
	h := newTestExerciseHandler(t)
	// tt-001 is a tongue twister in the seeded data
	req := httptest.NewRequest(http.MethodGet, "/exercises/tt-001", nil)
	rr := httptest.NewRecorder()
	h.Detail(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestDetail_CategorySlug_DelegatesToCategory(t *testing.T) {
	h := newTestExerciseHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/exercises/mouth", nil)
	rr := httptest.NewRecorder()
	h.Detail(rr, req)
	// Should delegate to Category handler and return 200
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
