package handler

import (
	"net/http"
	"strings"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/service"
)

// ExerciseHandler handles exercise-related HTTP requests
type ExerciseHandler struct {
	baseHandler     *BaseHandler
	exerciseService *service.ExerciseService
}

// NewExerciseHandler creates a new ExerciseHandler
func NewExerciseHandler(base *BaseHandler, exerciseSvc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		baseHandler:     base,
		exerciseService: exerciseSvc,
	}
}

// Index handles the home page
func (h *ExerciseHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.baseHandler.Error(w, http.StatusNotFound, "Page not found")
		return
	}

	data := struct {
		PageData
		RecentExercises []domain.Exercise
	}{
		PageData: h.baseHandler.NewPageData(r, "Speech Practice"),
	}

	exercises, _ := h.exerciseService.GetAllExercises()
	if len(exercises) > 5 {
		data.RecentExercises = exercises[:5]
	}

	h.baseHandler.Render(w, "index.html", data)
}

// List handles the exercise library page
func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.exerciseService.GetAllExercises()
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load exercises")
		return
	}

	data := struct {
		PageData
		Exercises []domain.Exercise
	}{
		PageData:  h.baseHandler.NewPageData(r, "Exercise Library"),
		Exercises: exercises,
	}

	h.baseHandler.Render(w, "exercises.html", data)
}

// Detail handles individual exercise pages
func (h *ExerciseHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/exercises/")
	if id == "" {
		h.baseHandler.Redirect(w, r, "/exercises", http.StatusFound)
		return
	}

	exercise, err := h.exerciseService.GetExerciseByID(id)
	if err != nil {
		h.baseHandler.Error(w, http.StatusNotFound, "Exercise not found")
		return
	}

	userID := h.baseHandler.GetUserID(r)
	completed, _ := h.exerciseService.IsExerciseCompleted(userID, id)

	data := struct {
		PageData
		Exercise  domain.Exercise
		Completed bool
	}{
		PageData:  h.baseHandler.NewPageData(r, exercise.Name),
		Exercise:  *exercise,
		Completed: completed,
	}

	h.baseHandler.Render(w, "exercise_detail.html", data)
}

// Category handles exercise category pages
func (h *ExerciseHandler) Category(w http.ResponseWriter, r *http.Request) {
	cat := strings.TrimPrefix(r.URL.Path, "/exercises/")

	var category domain.ExerciseCategory
	switch cat {
	case "mouth":
		category = domain.CategoryMouthExercise
	case "twisters":
		category = domain.CategoryTongueTwister
	case "diction":
		category = domain.CategoryDictionStrategy
	case "pacing":
		category = domain.CategoryPacingStrategy
	default:
		h.baseHandler.Error(w, http.StatusNotFound, "Category not found")
		return
	}

	filter := service.ExerciseFilter{
		Category: &category,
	}

	exercises, err := h.exerciseService.GetExercisesByFilter(filter)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load exercises")
		return
	}

	data := struct {
		PageData
		Category  string
		Exercises []domain.Exercise
	}{
		PageData:  h.baseHandler.NewPageData(r, "Category"),
		Category:  cat,
		Exercises: exercises,
	}

	h.baseHandler.Render(w, "category.html", data)
}

// Recommendations handles the recommendations page
func (h *ExerciseHandler) Recommendations(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageData
		Message string
	}{
		PageData: h.baseHandler.NewPageData(r, "Recommendations"),
		Message:  "Complete more exercises to receive personalized recommendations",
	}

	h.baseHandler.Render(w, "recommendations.html", data)
}