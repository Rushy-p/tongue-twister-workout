package service

import (
	"errors"
	"sort"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

// ExerciseService handles exercise-related business logic
type ExerciseService struct {
	exerciseRepo  infrastructure.ExerciseRepository
	progressRepo  infrastructure.ProgressRepository
	sessionRepo   infrastructure.SessionRepository
}

// NewExerciseService creates a new ExerciseService
func NewExerciseService(
	exerciseRepo infrastructure.ExerciseRepository,
	progressRepo infrastructure.ProgressRepository,
	sessionRepo infrastructure.SessionRepository,
) *ExerciseService {
	return &ExerciseService{
		exerciseRepo:  exerciseRepo,
		progressRepo:  progressRepo,
		sessionRepo:   sessionRepo,
	}
}

// ExerciseFilter contains filter criteria for exercises
type ExerciseFilter struct {
	Category           *domain.ExerciseCategory
	Difficulty         *domain.DifficultyLevel
	TargetSound        *domain.SoundTarget
	ArticulationPoint  *domain.ArticulationPoint
	CompletedOnly      bool
	IncompleteOnly     bool
}

// GetExerciseByID retrieves a single exercise by ID
func (s *ExerciseService) GetExerciseByID(id string) (*domain.Exercise, error) {
	if id == "" {
		return nil, errors.New("exercise ID cannot be empty")
	}
	return s.exerciseRepo.GetByID(id)
}

// GetAllExercises retrieves all exercises
func (s *ExerciseService) GetAllExercises() ([]domain.Exercise, error) {
	return s.exerciseRepo.GetAll()
}

// GetExercisesByFilter retrieves exercises matching the given filter
func (s *ExerciseService) GetExercisesByFilter(filter ExerciseFilter) ([]domain.Exercise, error) {
	var exercises []domain.Exercise
	var err error

	// Start with base query based on primary filter
	if filter.Category != nil {
		exercises, err = s.exerciseRepo.GetByCategory(*filter.Category)
	} else if filter.Difficulty != nil {
		exercises, err = s.exerciseRepo.GetByDifficulty(*filter.Difficulty)
	} else if filter.TargetSound != nil {
		exercises, err = s.exerciseRepo.GetByTargetSound(*filter.TargetSound)
	} else if filter.ArticulationPoint != nil {
		exercises, err = s.exerciseRepo.GetByArticulationPoint(*filter.ArticulationPoint)
	} else {
		exercises, err = s.exerciseRepo.GetAll()
	}

	if err != nil {
		return nil, err
	}

	// Apply additional filters
	result := make([]domain.Exercise, 0)
	for _, e := range exercises {
		if filter.CompletedOnly && e.CompletionCount == 0 {
			continue
		}
		if filter.IncompleteOnly && e.CompletionCount > 0 {
			continue
		}
		// Apply secondary filters
		if filter.Category != nil && e.Category != *filter.Category {
			continue
		}
		if filter.Difficulty != nil && e.Difficulty != *filter.Difficulty {
			continue
		}
		if filter.TargetSound != nil && !containsSound(e.TargetSounds, *filter.TargetSound) {
			continue
		}
		if filter.ArticulationPoint != nil && !containsArticulationPoint(e.ArticulationPoints, *filter.ArticulationPoint) {
			continue
		}
		result = append(result, e)
	}

	return result, nil
}

// IsExerciseCompleted checks if a user has completed a specific exercise
func (s *ExerciseService) IsExerciseCompleted(userID, exerciseID string) (bool, error) {
	if userID == "" {
		return false, errors.New("user ID cannot be empty")
	}
	if exerciseID == "" {
		return false, errors.New("exercise ID cannot be empty")
	}

	exercise, err := s.exerciseRepo.GetByID(exerciseID)
	if err != nil {
		return false, err
	}

	return exercise.CompletionCount > 0, nil
}

// GetExerciseCompletionStatus returns completion status for multiple exercises
func (s *ExerciseService) GetExerciseCompletionStatus(userID string, exerciseIDs []string) (map[string]bool, error) {
	result := make(map[string]bool)

	for _, id := range exerciseIDs {
		completed, err := s.IsExerciseCompleted(userID, id)
		if err != nil {
			// If exercise not found, mark as not completed
			result[id] = false
			continue
		}
		result[id] = completed
	}

	return result, nil
}

// UserPerformance holds performance analysis data for a user
type UserPerformance struct {
	UserID              string
	TotalExercises      int
	CompletedExercises  int
	CompletionRate      float64
	CategoryProgress    map[domain.ExerciseCategory]*CategoryPerformance
	SoundPerformance    map[domain.SoundTarget]*SoundPerformance
	RecentSessions      []domain.PracticeSession
	WeakestCategory     *domain.ExerciseCategory
	WeakestSound        *domain.SoundTarget
}

// CategoryPerformance holds performance data for a specific category
type CategoryPerformance struct {
	Category         domain.ExerciseCategory
	TotalExercises   int
	CompletedExercises int
	CompletionRate   float64
	AverageScore     float64
}

// SoundPerformance holds performance data for a specific sound
type SoundPerformance struct {
	Sound          domain.SoundTarget
	TotalAttempts  int
	SuccessfulAttempts int
	SuccessRate    float64
	AverageScore   float64
}

// AnalyzeUserPerformance analyzes user performance to identify areas for improvement
// Implements Requirement 10.1
func (s *ExerciseService) AnalyzeUserPerformance(userID string) (*UserPerformance, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Get all exercises
	allExercises, err := s.exerciseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Get user's sessions
	sessions, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Get category progress
	categoryProgress, err := s.getCategoryProgress(userID)
	if err != nil {
		categoryProgress = make(map[domain.ExerciseCategory]*CategoryPerformance)
	}

	// Calculate sound performance from session data
	soundPerformance := s.analyzeSoundPerformance(sessions, allExercises)

	// Calculate overall performance
	totalExercises := len(allExercises)
	completedExercises := 0
	for _, e := range allExercises {
		if e.CompletionCount > 0 {
			completedExercises++
		}
	}

	completionRate := 0.0
	if totalExercises > 0 {
		completionRate = float64(completedExercises) / float64(totalExercises) * 100
	}

	// Find weakest category
	var weakestCategory *domain.ExerciseCategory
	var lowestRate float64 = 100.0
	for cat, perf := range categoryProgress {
		if perf.CompletionRate < lowestRate {
			lowestRate = perf.CompletionRate
			weakestCategory = &cat
		}
	}

	// Find weakest sound
	var weakestSound *domain.SoundTarget
	var lowestSoundRate float64 = 100.0
	for sound, perf := range soundPerformance {
		if perf.SuccessRate < lowestSoundRate && perf.TotalAttempts > 0 {
			lowestSoundRate = perf.SuccessRate
			weakestSound = &sound
		}
	}

	// Get recent sessions (last 7 days)
	recentSessions := s.getRecentSessions(sessions, 7)

	return &UserPerformance{
		UserID:            userID,
		TotalExercises:    totalExercises,
		CompletedExercises: completedExercises,
		CompletionRate:    completionRate,
		CategoryProgress:  categoryProgress,
		SoundPerformance:  soundPerformance,
		RecentSessions:    recentSessions,
		WeakestCategory:   weakestCategory,
		WeakestSound:      weakestSound,
	}, nil
}

// getCategoryProgress retrieves category progress for a user
func (s *ExerciseService) getCategoryProgress(userID string) (map[domain.ExerciseCategory]*CategoryPerformance, error) {
	result := make(map[domain.ExerciseCategory]*CategoryPerformance)

	categories := []domain.ExerciseCategory{
		domain.CategoryMouthExercise,
		domain.CategoryTongueTwister,
		domain.CategoryDictionStrategy,
		domain.CategoryPacingStrategy,
	}

	for _, cat := range categories {
		progress, err := s.progressRepo.GetCategoryProgress(userID, cat)
		if err != nil {
			// If no progress found, create empty progress
			result[cat] = &CategoryPerformance{
				Category:       cat,
				TotalExercises: 0,
				CompletedExercises: 0,
				CompletionRate: 0,
				AverageScore:   0,
			}
			continue
		}

		rate := 0.0
		if progress.TotalExercises > 0 {
			rate = float64(progress.CompletedExercises) / float64(progress.TotalExercises) * 100
		}

		result[cat] = &CategoryPerformance{
			Category:         cat,
			TotalExercises:   progress.TotalExercises,
			CompletedExercises: progress.CompletedExercises,
			CompletionRate:   rate,
			AverageScore:     0, // Would need to calculate from session data
		}
	}

	return result, nil
}

// analyzeSoundPerformance analyzes performance for each sound target
func (s *ExerciseService) analyzeSoundPerformance(sessions []domain.PracticeSession, exercises []domain.Exercise) map[domain.SoundTarget]*SoundPerformance {
	result := make(map[domain.SoundTarget]*SoundPerformance)

	// Initialize all sounds
	sounds := []domain.SoundTarget{
		domain.SoundS, domain.SoundZ, domain.SoundR, domain.SoundL,
		domain.SoundTH, domain.SoundSH, domain.SoundCH, domain.SoundJ,
		domain.SoundK, domain.SoundG,
	}
	for _, sound := range sounds {
		result[sound] = &SoundPerformance{
			Sound:            sound,
			TotalAttempts:    0,
			SuccessfulAttempts: 0,
			SuccessRate:      0,
			AverageScore:     0,
		}
	}

	// Build exercise sound map
	exerciseSounds := make(map[string][]domain.SoundTarget)
	for _, e := range exercises {
		exerciseSounds[e.ID] = e.TargetSounds
	}

	// Analyze sessions
	for _, session := range sessions {
		for _, se := range session.Exercises {
			sounds := exerciseSounds[se.ExerciseID]
			for _, sound := range sounds {
				if perf, ok := result[sound]; ok {
					perf.TotalAttempts++
					if se.Score >= 70 { // Consider 70+ as successful
						perf.SuccessfulAttempts++
					}
					perf.AverageScore += float64(se.Score)
				}
			}
		}
	}

	// Calculate success rates
	for _, perf := range result {
		if perf.TotalAttempts > 0 {
			perf.SuccessRate = float64(perf.SuccessfulAttempts) / float64(perf.TotalAttempts) * 100
			perf.AverageScore = perf.AverageScore / float64(perf.TotalAttempts)
		}
	}

	return result
}

// getRecentSessions returns sessions from the last n days
func (s *ExerciseService) getRecentSessions(sessions []domain.PracticeSession, days int) []domain.PracticeSession {
	cutoff := time.Now().AddDate(0, 0, -days)
	var recent []domain.PracticeSession

	for _, session := range sessions {
		if session.StartTime.After(cutoff) {
			recent = append(recent, session)
		}
	}

	return recent
}

// GetRecommendations generates personalized exercise recommendations
// Implements Requirements 10.2, 10.3, 10.4
func (s *ExerciseService) GetRecommendations(userID string, limit int) ([]Recommendation, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if limit <= 0 {
		limit = 5
	}

	// Analyze user performance
	performance, err := s.AnalyzeUserPerformance(userID)
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation

	// Add recommendations for weakest category (Requirement 10.2)
	if performance.WeakestCategory != nil {
		catRecs, err := s.getRecommendationsForCategory(*performance.WeakestCategory, performance.CategoryProgress)
		if err == nil {
			recommendations = append(recommendations, catRecs...)
		}
	}

	// Add recommendations for struggling sounds (Requirement 10.3)
	if performance.WeakestSound != nil {
		soundRecs, err := s.getRecommendationsForSound(*performance.WeakestSound, performance.SoundPerformance)
		if err == nil {
			recommendations = append(recommendations, soundRecs...)
		}
	}

	// Add general practice recommendations if needed
	if len(recommendations) < limit {
		generalRecs, err := s.getGeneralRecommendations(userID, limit-len(recommendations), performance)
		if err == nil {
			recommendations = append(recommendations, generalRecs...)
		}
	}

	// Sort by priority and limit
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// getRecommendationsForCategory returns recommendations for a specific category
func (s *ExerciseService) getRecommendationsForCategory(category domain.ExerciseCategory, categoryProgress map[domain.ExerciseCategory]*CategoryPerformance) ([]Recommendation, error) {
	exercises, err := s.exerciseRepo.GetByCategory(category)
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation
	perf := categoryProgress[category]

	for _, e := range exercises {
		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		reason := "Improve your skills in this category"
		priority := 50

		if perf != nil {
			// Lower completion rate = higher priority
			priority = 100 - int(perf.CompletionRate)
			if priority < 30 {
				priority = 30
			}
		}

		recommendations = append(recommendations, Recommendation{
			Exercise:        &e,
			Reason:          reason,
			Priority:        priority,
			RecommendationType: RecommendationTypeLowCompletion,
		})
	}

	return recommendations, nil
}

// getRecommendationsForSound returns recommendations for a specific sound
func (s *ExerciseService) getRecommendationsForSound(sound domain.SoundTarget, soundPerformance map[domain.SoundTarget]*SoundPerformance) ([]Recommendation, error) {
	exercises, err := s.exerciseRepo.GetByTargetSound(sound)
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation
	perf := soundPerformance[sound]

	for _, e := range exercises {
		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		reason := "Practice this sound to improve your pronunciation"
		priority := 70

		if perf != nil {
			// Lower success rate = higher priority
			priority = 100 - int(perf.SuccessRate)
			if priority < 40 {
				priority = 40
			}
		}

		recommendations = append(recommendations, Recommendation{
			Exercise:        &e,
			Reason:          reason,
			Priority:        priority,
			RecommendationType: RecommendationTypeStrugglingSound,
		})
	}

	return recommendations, nil
}

// getGeneralRecommendations returns general practice recommendations
func (s *ExerciseService) getGeneralRecommendations(userID string, limit int, performance *UserPerformance) ([]Recommendation, error) {
	exercises, err := s.exerciseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation

	for _, e := range exercises {
		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		// Skip if already recommended
		alreadyRecommended := false
		for _, r := range recommendations {
			if r.Exercise.ID == e.ID {
				alreadyRecommended = true
				break
			}
		}
		if alreadyRecommended {
			continue
		}

		reason := "Daily practice recommendation"
		priority := 20

		recommendations = append(recommendations, Recommendation{
			Exercise:        &e,
			Reason:          reason,
			Priority:        priority,
			RecommendationType: RecommendationTypeDailyPractice,
		})

		if len(recommendations) >= limit {
			break
		}
	}

	return recommendations, nil
}

// GetTongueTwistersForSound returns tongue twisters targeting a specific sound
// Implements Requirement 3.8
func (s *ExerciseService) GetTongueTwistersForSound(sound domain.SoundTarget) ([]domain.Exercise, error) {
	return s.exerciseRepo.GetByTargetSound(sound)
}

// GetDailyRecommendationSummary generates a daily summary of recommendations
// Implements Requirement 10.4
func (s *ExerciseService) GetDailyRecommendationSummary(userID string) (*DailyRecommendationSummary, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Get user performance
	performance, err := s.AnalyzeUserPerformance(userID)
	if err != nil {
		return nil, err
	}

	// Get recommendations
	recommendations, err := s.GetRecommendations(userID, 5)
	if err != nil {
		return nil, err
	}

	// Build summary
	summary := &DailyRecommendationSummary{
		Date:            time.Now(),
		TotalExercises:  performance.TotalExercises,
		CompletedExercises: performance.CompletedExercises,
		CompletionRate:  performance.CompletionRate,
		Recommendations: recommendations,
		WeakestArea:     "",
		FocusSound:      "",
	}

	if performance.WeakestCategory != nil {
		summary.WeakestArea = string(*performance.WeakestCategory)
	}

	if performance.WeakestSound != nil {
		summary.FocusSound = string(*performance.WeakestSound)
	}

	return summary, nil
}

// DailyRecommendationSummary represents a daily recommendation summary
type DailyRecommendationSummary struct {
	Date              time.Time
	TotalExercises    int
	CompletedExercises int
	CompletionRate    float64
	Recommendations   []Recommendation
	WeakestArea       string
	FocusSound        string
}

// MarkExerciseCompleted marks an exercise as completed for a user
// Implements Requirement 1.5
func (s *ExerciseService) MarkExerciseCompleted(userID, exerciseID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if exerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

	exercise, err := s.exerciseRepo.GetByID(exerciseID)
	if err != nil {
		return err
	}

	// Increment completion count
	newCount := exercise.CompletionCount + 1
	return s.exerciseRepo.UpdateCompletionCount(exerciseID, newCount)
}

// sortByDifficulty sorts exercises beginner → intermediate → advanced, then by ID for stability.
func sortByDifficulty(exercises []domain.Exercise) {
	order := map[domain.DifficultyLevel]int{
		domain.DifficultyBeginner:     0,
		domain.DifficultyIntermediate: 1,
		domain.DifficultyAdvanced:     2,
	}
	sort.Slice(exercises, func(i, j int) bool {
		oi := order[exercises[i].Difficulty]
		oj := order[exercises[j].Difficulty]
		if oi != oj {
			return oi < oj
		}
		return exercises[i].ID < exercises[j].ID
	})
}

// SortByDifficulty sorts a slice of exercises beginner → intermediate → advanced.
func (s *ExerciseService) SortByDifficulty(exercises []domain.Exercise) {
	sortByDifficulty(exercises)
}

func containsSound(sounds []domain.SoundTarget, target domain.SoundTarget) bool {
	for _, s := range sounds {
		if s == target {
			return true
		}
	}
	return false
}

func containsArticulationPoint(points []domain.ArticulationPoint, target domain.ArticulationPoint) bool {
	for _, p := range points {
		if p == target {
			return true
		}
	}
	return false
}