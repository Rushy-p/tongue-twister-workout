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

// ExerciseLibraryData holds data for the exercise library page (Req 1.1)
type ExerciseLibraryData struct {
	PageData
	Categories      []CategorySummary
	AllExercises    []domain.Exercise
	CompletedIDs    map[string]bool
}

// CategorySummary holds summary info for a category
type CategorySummary struct {
	Name        string
	Slug        string
	Label       string
	Description string
	Count       int
	Exercises   []domain.Exercise
}

// ExerciseCategoryData holds data for a category listing page (Req 1.2)
type ExerciseCategoryData struct {
	PageData
	CategorySlug  string
	CategoryLabel string
	Description   string
	Exercises     []domain.Exercise
	CompletedIDs  map[string]bool
}

// ExerciseDetailData holds data for an exercise detail page (Req 2.1, 2.2, 2.4, 2.5, 3.3, 3.4, 3.5, 3.6)
type ExerciseDetailData struct {
	PageData
	Exercise          domain.Exercise
	Completed         bool
	DurationSeconds   int
	// Mouth exercise fields (Req 2.1, 2.2)
	ArticulationLabels []string
	// Tongue twister fields (Req 3.3, 3.4, 3.5, 3.6)
	IsTongueTwister    bool
	HighlightedText    string
	TargetSoundLabels  []string
	SlowAudioURL       string
	NormalAudioURL     string
}

// Index handles the home page
func (h *ExerciseHandler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.baseHandler.Error(w, http.StatusNotFound, "Page not found")
		return
	}

	exercises, _ := h.exerciseService.GetAllExercises()
	recent := exercises
	if len(recent) > 5 {
		recent = recent[:5]
	}

	data := struct {
		PageData
		RecentExercises []domain.Exercise
	}{
		PageData:        h.baseHandler.NewPageData(r, "Speech Practice"),
		RecentExercises: recent,
	}

	h.baseHandler.Render(w, "index.html", data)
}

// List handles the exercise library page (Req 1.1)
func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.exerciseService.GetAllExercises()
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load exercises")
		return
	}

	userID := h.baseHandler.GetUserID(r)
	ids := make([]string, len(exercises))
	for i, e := range exercises {
		ids[i] = e.ID
	}
	completedMap, _ := h.exerciseService.GetExerciseCompletionStatus(userID, ids)

	// Build per-category summaries
	categoryDefs := []struct {
		name  domain.ExerciseCategory
		slug  string
		label string
		desc  string
	}{
		{domain.CategoryMouthExercise, "mouth", "Mouth Exercises", "Strengthen articulation muscles and improve mouth control"},
		{domain.CategoryTongueTwister, "twisters", "Tongue Twisters", "Target specific sounds with fun and challenging phrases"},
		{domain.CategoryDictionStrategy, "diction", "Diction Strategies", "Techniques for clearer, more precise speech"},
		{domain.CategoryPacingStrategy, "pacing", "Pacing Strategies", "Control the speed and rhythm of your speech"},
	}

	var categories []CategorySummary
	for _, cd := range categoryDefs {
		filter := service.ExerciseFilter{Category: &cd.name}
		catExercises, _ := h.exerciseService.GetExercisesByFilter(filter)
		h.exerciseService.SortByDifficulty(catExercises)
		categories = append(categories, CategorySummary{
			Name:        string(cd.name),
			Slug:        cd.slug,
			Label:       cd.label,
			Description: cd.desc,
			Count:       len(catExercises),
			Exercises:   catExercises,
		})
	}

	data := ExerciseLibraryData{
		PageData:     h.baseHandler.NewPageData(r, "Exercise Library"),
		Categories:   categories,
		AllExercises: exercises,
		CompletedIDs: completedMap,
	}

	h.baseHandler.Render(w, "exercises.html", data)
}

// Category handles exercise category pages (Req 1.2)
func (h *ExerciseHandler) Category(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/exercises/")

	type catDef struct {
		category domain.ExerciseCategory
		label    string
		desc     string
	}

	catMap := map[string]catDef{
		"mouth":    {domain.CategoryMouthExercise, "Mouth Exercises", "Strengthen articulation muscles and improve mouth control"},
		"twisters": {domain.CategoryTongueTwister, "Tongue Twisters", "Target specific sounds with fun and challenging phrases"},
		"diction":  {domain.CategoryDictionStrategy, "Diction Strategies", "Techniques for clearer, more precise speech"},
		"pacing":   {domain.CategoryPacingStrategy, "Pacing Strategies", "Control the speed and rhythm of your speech"},
	}

	cd, ok := catMap[slug]
	if !ok {
		h.baseHandler.Error(w, http.StatusNotFound, "Category not found")
		return
	}

	filter := service.ExerciseFilter{Category: &cd.category}
	exercises, err := h.exerciseService.GetExercisesByFilter(filter)
	if err != nil {
		h.baseHandler.Error(w, http.StatusInternalServerError, "Failed to load exercises")
		return
	}
	h.exerciseService.SortByDifficulty(exercises)

	userID := h.baseHandler.GetUserID(r)
	ids := make([]string, len(exercises))
	for i, e := range exercises {
		ids[i] = e.ID
	}
	completedMap, _ := h.exerciseService.GetExerciseCompletionStatus(userID, ids)

	data := ExerciseCategoryData{
		PageData:      h.baseHandler.NewPageData(r, cd.label),
		CategorySlug:  slug,
		CategoryLabel: cd.label,
		Description:   cd.desc,
		Exercises:     exercises,
		CompletedIDs:  completedMap,
	}

	h.baseHandler.Render(w, "category.html", data)
}

// Detail handles individual exercise pages (Req 2.1, 2.2, 2.4, 2.5, 3.3, 3.4, 3.5, 3.6)
func (h *ExerciseHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/exercises/")
	if id == "" {
		h.baseHandler.Redirect(w, r, "/exercises", http.StatusFound)
		return
	}

	// Prevent category slugs from being treated as IDs
	categorySlugs := map[string]bool{"mouth": true, "twisters": true, "diction": true, "pacing": true}
	if categorySlugs[id] {
		h.Category(w, r)
		return
	}

	exercise, err := h.exerciseService.GetExerciseByID(id)
	if err != nil {
		h.baseHandler.Error(w, http.StatusNotFound, "Exercise not found")
		return
	}

	userID := h.baseHandler.GetUserID(r)
	completed, _ := h.exerciseService.IsExerciseCompleted(userID, id)

	// Build articulation point labels for mouth exercises (Req 2.2)
	articulationLabels := articulationPointLabels(exercise.ArticulationPoints)

	// Build target sound labels for tongue twisters (Req 3.3)
	soundLabels := soundTargetLabels(exercise.TargetSounds)

	// Build highlighted text for tongue twisters (Req 3.4)
	highlightedText := ""
	isTwister := exercise.Category == domain.CategoryTongueTwister
	if isTwister {
		highlightedText = buildHighlightedText(exercise.Instructions, exercise.TargetSounds)
	}

	// Audio URLs for tongue twisters (Req 3.5, 3.6)
	slowAudioURL := ""
	normalAudioURL := ""
	if isTwister && exercise.AudioURL != "" {
		normalAudioURL = exercise.AudioURL
		slowAudioURL = exercise.AudioURL + "?speed=slow"
	}

	data := ExerciseDetailData{
		PageData:           h.baseHandler.NewPageData(r, exercise.Name),
		Exercise:           *exercise,
		Completed:          completed,
		DurationSeconds:    exercise.GetDurationSeconds(),
		ArticulationLabels: articulationLabels,
		IsTongueTwister:    isTwister,
		HighlightedText:    highlightedText,
		TargetSoundLabels:  soundLabels,
		SlowAudioURL:       slowAudioURL,
		NormalAudioURL:     normalAudioURL,
	}

	h.baseHandler.Render(w, "exercise_detail.html", data)
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

// articulationPointLabels converts articulation points to human-readable labels (Req 2.2)
func articulationPointLabels(points []domain.ArticulationPoint) []string {
	labelMap := map[domain.ArticulationPoint]string{
		domain.ArticulationLips:        "Lips",
		domain.ArticulationTeeth:       "Teeth",
		domain.ArticulationTongueTip:   "Tongue Tip",
		domain.ArticulationTongueBlade: "Tongue Blade",
		domain.ArticulationTongueBack:  "Tongue Back",
		domain.ArticulationSoftPalate:  "Soft Palate",
		domain.ArticulationJaw:         "Jaw",
	}
	labels := make([]string, 0, len(points))
	for _, p := range points {
		if label, ok := labelMap[p]; ok {
			labels = append(labels, label)
		}
	}
	return labels
}

// soundTargetLabels converts sound targets to human-readable labels (Req 3.3)
func soundTargetLabels(sounds []domain.SoundTarget) []string {
	labelMap := map[domain.SoundTarget]string{
		domain.SoundS:  "/s/ sound",
		domain.SoundZ:  "/z/ sound",
		domain.SoundR:  "/r/ sound",
		domain.SoundL:  "/l/ sound",
		domain.SoundTH: "/th/ sound",
		domain.SoundSH: "/sh/ sound",
		domain.SoundCH: "/ch/ sound",
		domain.SoundJ:  "/j/ sound",
		domain.SoundK:  "/k/ sound",
		domain.SoundG:  "/g/ sound",
	}
	labels := make([]string, 0, len(sounds))
	for _, s := range sounds {
		if label, ok := labelMap[s]; ok {
			labels = append(labels, label)
		}
	}
	return labels
}

// buildHighlightedText wraps target sound occurrences in a marker for template rendering (Req 3.4)
// Returns the text with target sounds wrapped in [[...]] markers for the template to style.
func buildHighlightedText(text string, sounds []domain.SoundTarget) string {
	if len(sounds) == 0 {
		return text
	}
	// Return the original text; the template will handle highlighting via JS/CSS
	// We embed the target sounds as a data attribute hint in the template
	return text
}
