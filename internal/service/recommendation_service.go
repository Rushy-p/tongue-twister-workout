package service

import (
	"errors"
	"sort"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
	"speech-practice-app/internal/infrastructure"
)

// RecommendationService handles recommendation-related business logic
type RecommendationService struct {
	exerciseRepo  infrastructure.ExerciseRepository
	progressRepo  infrastructure.ProgressRepository
	sessionRepo   infrastructure.SessionRepository
	recommendationRepo infrastructure.RecommendationRepository
}

// RecommendationRepository defines the interface for recommendation data access
type RecommendationRepository interface {
	SaveRecommendation(rec *domain.RecommendationRecord) error
	GetRecommendation(id string) (*domain.RecommendationRecord, error)
	GetUserRecommendations(userID string) ([]domain.RecommendationRecord, error)
	GetRejectedRecommendations(userID string) ([]domain.RejectedRecommendation, error)
	SaveRejectedRecommendation(rec *domain.RejectedRecommendation) error
	GetAcceptedRecommendations(userID string) ([]domain.AcceptedRecommendation, error)
	SaveAcceptedRecommendation(rec *domain.AcceptedRecommendation) error
}

// InMemoryRecommendationRepository provides in-memory storage for recommendations
type InMemoryRecommendationRepository struct {
	recommendations     map[string]domain.RecommendationRecord
	rejectedRecommendations map[string]map[string]domain.RejectedRecommendation
	acceptedRecommendations map[string]map[string]domain.AcceptedRecommendation
	mu                  sync.RWMutex
}

// NewRecommendationService creates a new RecommendationService
func NewRecommendationService(
	exerciseRepo infrastructure.ExerciseRepository,
	progressRepo infrastructure.ProgressRepository,
	sessionRepo infrastructure.SessionRepository,
	recommendationRepo infrastructure.RecommendationRepository,
) *RecommendationService {
	return &RecommendationService{
		exerciseRepo:  exerciseRepo,
		progressRepo:  progressRepo,
		sessionRepo:   sessionRepo,
		recommendationRepo: recommendationRepo,
	}
}

// Recommendation types
const (
	RecommendationTypeLowCompletion RecommendationType = "low_completion"
	RecommendationTypeStrugglingSound RecommendationType = "struggling_sound"
	RecommendationTypeStreakMilestone RecommendationType = "streak_milestone"
	RecommendationTypeDailyPractice   RecommendationType = "daily_practice"
)

// Recommendation represents a personalized exercise recommendation
type Recommendation struct {
	Exercise          *domain.Exercise
	Reason            string
	Priority          int
	RecommendationType RecommendationType
}

// RecommendationSummary represents a daily recommendation summary
type RecommendationSummary struct {
	Date                time.Time
	TotalExercises      int
	CompletedExercises  int
	CompletionRate      float64
	Recommendations     []Recommendation
	WeakestArea         string
	FocusSound          string
	CurrentStreak       int
	ShouldAdvanceLevel  bool
}

// AnalyzeUserPerformance analyzes user performance to identify areas for improvement
// Implements Requirement 10.1
func (s *RecommendationService) AnalyzeUserPerformance(userID string) (*UserPerformance, error) {
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
func (s *RecommendationService) getCategoryProgress(userID string) (map[domain.ExerciseCategory]*CategoryPerformance, error) {
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
				Category:         cat,
				TotalExercises:   0,
				CompletedExercises: 0,
				CompletionRate:   0,
				AverageScore:     0,
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
			AverageScore:     0,
		}
	}

	return result, nil
}

// analyzeSoundPerformance analyzes performance for each sound target
func (s *RecommendationService) analyzeSoundPerformance(sessions []domain.PracticeSession, exercises []domain.Exercise) map[domain.SoundTarget]*SoundPerformance {
	result := make(map[domain.SoundTarget]*SoundPerformance)

	// Initialize all sounds
	sounds := []domain.SoundTarget{
		domain.SoundS, domain.SoundZ, domain.SoundR, domain.SoundL,
		domain.SoundTH, domain.SoundSH, domain.SoundCH, domain.SoundJ,
		domain.SoundK, domain.SoundG,
	}
	for _, sound := range sounds {
		result[sound] = &SoundPerformance{
			Sound:              sound,
			TotalAttempts:      0,
			SuccessfulAttempts: 0,
			SuccessRate:        0,
			AverageScore:       0,
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
					if se.Score >= 70 {
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
func (s *RecommendationService) getRecentSessions(sessions []domain.PracticeSession, days int) []domain.PracticeSession {
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
// Implements Requirements 10.2, 10.3
func (s *RecommendationService) GetRecommendations(userID string, limit int) ([]Recommendation, error) {
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

	// Get rejected recommendations to filter out
	rejectedRecs, err := s.recommendationRepo.GetRejectedRecommendations(userID)
	if err != nil {
		rejectedRecs = []domain.RejectedRecommendation{}
	}

	// Create a set of rejected exercise IDs that are still within 7 days
	rejectedExerciseIDs := make(map[string]bool)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	for _, rec := range rejectedRecs {
		if rec.RejectedAt.After(sevenDaysAgo) {
			rejectedExerciseIDs[rec.ExerciseID] = true
		}
	}

	var recommendations []Recommendation

	// Add recommendations for weakest category (Requirement 10.2)
	if performance.WeakestCategory != nil {
		catRecs, err := s.getRecommendationsForCategory(*performance.WeakestCategory, performance.CategoryProgress, rejectedExerciseIDs)
		if err == nil {
			recommendations = append(recommendations, catRecs...)
		}
	}

	// Add recommendations for struggling sounds (Requirement 10.3)
	if performance.WeakestSound != nil {
		soundRecs, err := s.getRecommendationsForSound(*performance.WeakestSound, performance.SoundPerformance, rejectedExerciseIDs)
		if err == nil {
			recommendations = append(recommendations, soundRecs...)
		}
	}

	// Add streak-based advanced recommendations (Requirement 10.5)
	streak, err := s.progressRepo.GetStreak(userID)
	if err == nil && streak.CurrentStreak >= 7 {
		advancedRecs, err := s.getAdvancedRecommendations(userID, rejectedExerciseIDs)
		if err == nil {
			recommendations = append(recommendations, advancedRecs...)
		}
	}

	// Add general practice recommendations if needed
	if len(recommendations) < limit {
		generalRecs, err := s.getGeneralRecommendations(userID, limit-len(recommendations), performance, rejectedExerciseIDs)
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
func (s *RecommendationService) getRecommendationsForCategory(category domain.ExerciseCategory, categoryProgress map[domain.ExerciseCategory]*CategoryPerformance, rejectedExerciseIDs map[string]bool) ([]Recommendation, error) {
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

		// Skip rejected exercises within 7 days
		if rejectedExerciseIDs[e.ID] {
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
			Exercise:          &e,
			Reason:            reason,
			Priority:          priority,
			RecommendationType: RecommendationTypeLowCompletion,
		})
	}

	return recommendations, nil
}

// getRecommendationsForSound returns recommendations for a specific sound
// Implements Requirement 10.3 - prioritize tongue twisters for struggling sounds
func (s *RecommendationService) getRecommendationsForSound(sound domain.SoundTarget, soundPerformance map[domain.SoundTarget]*SoundPerformance, rejectedExerciseIDs map[string]bool) ([]Recommendation, error) {
	exercises, err := s.exerciseRepo.GetByTargetSound(sound)
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation
	perf := soundPerformance[sound]

	// First, prioritize tongue twisters for struggling sounds (Requirement 10.3)
	for _, e := range exercises {
		if e.Category != domain.CategoryTongueTwister {
			continue
		}

		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		// Skip rejected exercises within 7 days
		if rejectedExerciseIDs[e.ID] {
			continue
		}

		reason := "Practice this tongue twister to improve your " + string(sound) + " sound"
		priority := 80 // Higher priority for tongue twisters targeting struggling sounds

		if perf != nil {
			// Lower success rate = higher priority
			priority = 100 - int(perf.SuccessRate)
			if priority < 50 {
				priority = 50
			}
		}

		recommendations = append(recommendations, Recommendation{
			Exercise:          &e,
			Reason:            reason,
			Priority:          priority,
			RecommendationType: RecommendationTypeStrugglingSound,
		})
	}

	// Then add other exercises targeting this sound
	for _, e := range exercises {
		if e.Category == domain.CategoryTongueTwister {
			continue // Already added
		}

		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		// Skip rejected exercises within 7 days
		if rejectedExerciseIDs[e.ID] {
			continue
		}

		reason := "Practice this exercise to improve your " + string(sound) + " sound"
		priority := 60

		if perf != nil {
			priority = 100 - int(perf.SuccessRate)
			if priority < 40 {
				priority = 40
			}
		}

		recommendations = append(recommendations, Recommendation{
			Exercise:          &e,
			Reason:            reason,
			Priority:          priority,
			RecommendationType: RecommendationTypeStrugglingSound,
		})
	}

	return recommendations, nil
}

// getAdvancedRecommendations returns advanced exercises for users with 7+ day streaks
// Implements Requirement 10.5
func (s *RecommendationService) getAdvancedRecommendations(userID string, rejectedExerciseIDs map[string]bool) ([]Recommendation, error) {
	exercises, err := s.exerciseRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var recommendations []Recommendation

	for _, e := range exercises {
		// Only recommend advanced exercises
		if e.Difficulty != domain.DifficultyAdvanced {
			continue
		}

		// Skip completed exercises
		if e.CompletionCount > 0 {
			continue
		}

		// Skip rejected exercises within 7 days
		if rejectedExerciseIDs[e.ID] {
			continue
		}

		reason := "Congratulations on your 7-day streak! Try this advanced exercise."
		priority := 90 // High priority for streak milestones

		recommendations = append(recommendations, Recommendation{
			Exercise:          &e,
			Reason:            reason,
			Priority:          priority,
			RecommendationType: RecommendationTypeStreakMilestone,
		})
	}

	return recommendations, nil
}

// getGeneralRecommendations returns general practice recommendations
func (s *RecommendationService) getGeneralRecommendations(userID string, limit int, performance *UserPerformance, rejectedExerciseIDs map[string]bool) ([]Recommendation, error) {
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

		// Skip rejected exercises within 7 days
		if rejectedExerciseIDs[e.ID] {
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
			Exercise:          &e,
			Reason:            reason,
			Priority:          priority,
			RecommendationType: RecommendationTypeDailyPractice,
		})

		if len(recommendations) >= limit {
			break
		}
	}

	return recommendations, nil
}

// GetDailyRecommendationSummary generates a daily summary of recommendations
// Implements Requirement 10.4
func (s *RecommendationService) GetDailyRecommendationSummary(userID string) (*RecommendationSummary, error) {
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

	// Get current streak
	streak, err := s.progressRepo.GetStreak(userID)
	currentStreak := 0
	if err == nil {
		currentStreak = streak.CurrentStreak
	}

	// Determine if user should advance to more challenging exercises
	shouldAdvance := currentStreak >= 7

	// Build summary
	summary := &RecommendationSummary{
		Date:               time.Now(),
		TotalExercises:     performance.TotalExercises,
		CompletedExercises: performance.CompletedExercises,
		CompletionRate:     performance.CompletionRate,
		Recommendations:    recommendations,
		WeakestArea:        "",
		FocusSound:         "",
		CurrentStreak:      currentStreak,
		ShouldAdvanceLevel: shouldAdvance,
	}

	if performance.WeakestCategory != nil {
		summary.WeakestArea = string(*performance.WeakestCategory)
	}

	if performance.WeakestSound != nil {
		summary.FocusSound = string(*performance.WeakestSound)
	}

	return summary, nil
}

// AcceptRecommendation marks a recommendation as accepted
// Implements Requirement 10.6
func (s *RecommendationService) AcceptRecommendation(userID, exerciseID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if exerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

	acceptedRec := domain.AcceptedRecommendation{
		ID:          generateRecommendationID(),
		UserID:      userID,
		ExerciseID:  exerciseID,
		AcceptedAt:  time.Now(),
	}

	return s.recommendationRepo.SaveAcceptedRecommendation(&acceptedRec)
}

// RejectRecommendation marks a recommendation as rejected
// Implements Requirements 10.6, 10.7
func (s *RecommendationService) RejectRecommendation(userID, exerciseID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if exerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

	rejectedRec := domain.RejectedRecommendation{
		ID:         generateRecommendationID(),
		UserID:     userID,
		ExerciseID: exerciseID,
		RejectedAt: time.Now(),
	}

	return s.recommendationRepo.SaveRejectedRecommendation(&rejectedRec)
}

// GetAcceptedRecommendations returns all accepted recommendations for a user
func (s *RecommendationService) GetAcceptedRecommendations(userID string) ([]domain.AcceptedRecommendation, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.recommendationRepo.GetAcceptedRecommendations(userID)
}

// GetRejectedRecommendations returns all rejected recommendations for a user
func (s *RecommendationService) GetRejectedRecommendations(userID string) ([]domain.RejectedRecommendation, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.recommendationRepo.GetRejectedRecommendations(userID)
}

// IsRecommendationRejected checks if a recommendation was recently rejected (within 7 days)
func (s *RecommendationService) IsRecommendationRejected(userID, exerciseID string) (bool, error) {
	if userID == "" {
		return false, errors.New("user ID cannot be empty")
	}
	if exerciseID == "" {
		return false, errors.New("exercise ID cannot be empty")
	}

	rejectedRecs, err := s.recommendationRepo.GetRejectedRecommendations(userID)
	if err != nil {
		return false, nil
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	for _, rec := range rejectedRecs {
		if rec.ExerciseID == exerciseID && rec.RejectedAt.After(sevenDaysAgo) {
			return true, nil
		}
	}

	return false, nil
}

// generateRecommendationID generates a unique ID for recommendations
func generateRecommendationID() string {
	return "rec-" + time.Now().Format("20060102150405")
}

// NewInMemoryRecommendationRepository creates a new in-memory recommendation repository
func NewInMemoryRecommendationRepository() *InMemoryRecommendationRepository {
	return &InMemoryRecommendationRepository{
		recommendations:       make(map[string]domain.RecommendationRecord),
		rejectedRecommendations: make(map[string]map[string]domain.RejectedRecommendation),
		acceptedRecommendations: make(map[string]map[string]domain.AcceptedRecommendation),
	}
}

// SaveRecommendation saves a recommendation record
func (r *InMemoryRecommendationRepository) SaveRecommendation(rec *domain.RecommendationRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recommendations[rec.ID] = *rec
	return nil
}

// GetRecommendation returns a recommendation by ID
func (r *InMemoryRecommendationRepository) GetRecommendation(id string) (*domain.RecommendationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if rec, exists := r.recommendations[id]; exists {
		return &rec, nil
	}
	return nil, errors.New("recommendation not found")
}

// GetUserRecommendations returns all recommendations for a user
func (r *InMemoryRecommendationRepository) GetUserRecommendations(userID string) ([]domain.RecommendationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.RecommendationRecord
	for _, rec := range r.recommendations {
		if rec.UserID == userID {
			result = append(result, rec)
		}
	}
	return result, nil
}

// GetRejectedRecommendations returns all rejected recommendations for a user
func (r *InMemoryRecommendationRepository) GetRejectedRecommendations(userID string) ([]domain.RejectedRecommendation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if userRecs, exists := r.rejectedRecommendations[userID]; exists {
		result := make([]domain.RejectedRecommendation, 0, len(userRecs))
		for _, rec := range userRecs {
			result = append(result, rec)
		}
		return result, nil
	}
	return []domain.RejectedRecommendation{}, nil
}

// SaveRejectedRecommendation saves a rejected recommendation
func (r *InMemoryRecommendationRepository) SaveRejectedRecommendation(rec *domain.RejectedRecommendation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.rejectedRecommendations[rec.UserID] == nil {
		r.rejectedRecommendations[rec.UserID] = make(map[string]domain.RejectedRecommendation)
	}
	r.rejectedRecommendations[rec.UserID][rec.ExerciseID] = *rec
	return nil
}

// GetAcceptedRecommendations returns all accepted recommendations for a user
func (r *InMemoryRecommendationRepository) GetAcceptedRecommendations(userID string) ([]domain.AcceptedRecommendation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if userRecs, exists := r.acceptedRecommendations[userID]; exists {
		result := make([]domain.AcceptedRecommendation, 0, len(userRecs))
		for _, rec := range userRecs {
			result = append(result, rec)
		}
		return result, nil
	}
	return []domain.AcceptedRecommendation{}, nil
}

// SaveAcceptedRecommendation saves an accepted recommendation
func (r *InMemoryRecommendationRepository) SaveAcceptedRecommendation(rec *domain.AcceptedRecommendation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.acceptedRecommendations[rec.UserID] == nil {
		r.acceptedRecommendations[rec.UserID] = make(map[string]domain.AcceptedRecommendation)
	}
	r.acceptedRecommendations[rec.UserID][rec.ExerciseID] = *rec
	return nil
}