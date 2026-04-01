package repository

import (
	"errors"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

// ExerciseRepository defines the interface for exercise data access
type ExerciseRepository interface {
	GetAll() ([]domain.Exercise, error)
	GetByID(id string) (*domain.Exercise, error)
	GetByCategory(category domain.ExerciseCategory) ([]domain.Exercise, error)
	GetByDifficulty(difficulty domain.DifficultyLevel) ([]domain.Exercise, error)
	GetByTargetSound(sound domain.SoundTarget) ([]domain.Exercise, error)
	GetByArticulationPoint(point domain.ArticulationPoint) ([]domain.Exercise, error)
	Save(exercise *domain.Exercise) error
	UpdateCompletionCount(id string, count int) error
}

// InMemoryExerciseRepository provides in-memory storage for exercises
type InMemoryExerciseRepository struct {
	exercises map[string]domain.Exercise
	mu        sync.RWMutex
}

// NewInMemoryExerciseRepository creates a new in-memory exercise repository
func NewInMemoryExerciseRepository() *InMemoryExerciseRepository {
	repo := &InMemoryExerciseRepository{
		exercises: make(map[string]domain.Exercise),
	}
	repo.seedExercises()
	return repo
}

// GetAll returns all exercises
func (r *InMemoryExerciseRepository) GetAll() ([]domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exercises := make([]domain.Exercise, 0, len(r.exercises))
	for _, e := range r.exercises {
		exercises = append(exercises, e)
	}
	return exercises, nil
}

// GetByID returns an exercise by its ID
func (r *InMemoryExerciseRepository) GetByID(id string) (*domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exercise, exists := r.exercises[id]
	if !exists {
		return nil, errors.New("exercise not found")
	}
	return &exercise, nil
}

// GetByCategory returns exercises filtered by category
func (r *InMemoryExerciseRepository) GetByCategory(category domain.ExerciseCategory) ([]domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Exercise
	for _, e := range r.exercises {
		if e.Category == category {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetByDifficulty returns exercises filtered by difficulty
func (r *InMemoryExerciseRepository) GetByDifficulty(difficulty domain.DifficultyLevel) ([]domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Exercise
	for _, e := range r.exercises {
		if e.Difficulty == difficulty {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetByTargetSound returns exercises filtered by target sound
func (r *InMemoryExerciseRepository) GetByTargetSound(sound domain.SoundTarget) ([]domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Exercise
	for _, e := range r.exercises {
		for _, ts := range e.TargetSounds {
			if ts == sound {
				result = append(result, e)
				break
			}
		}
	}
	return result, nil
}

// GetByArticulationPoint returns exercises filtered by articulation point
func (r *InMemoryExerciseRepository) GetByArticulationPoint(point domain.ArticulationPoint) ([]domain.Exercise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Exercise
	for _, e := range r.exercises {
		for _, ap := range e.ArticulationPoints {
			if ap == point {
				result = append(result, e)
				break
			}
		}
	}
	return result, nil
}

// Save adds or updates an exercise
func (r *InMemoryExerciseRepository) Save(exercise *domain.Exercise) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if exercise.ID == "" {
		return errors.New("exercise ID cannot be empty")
	}
	exercise.UpdatedAt = time.Now()
	if exercise.CreatedAt.IsZero() {
		exercise.CreatedAt = time.Now()
	}
	r.exercises[exercise.ID] = *exercise
	return nil
}

// UpdateCompletionCount updates the completion count for an exercise
func (r *InMemoryExerciseRepository) UpdateCompletionCount(id string, count int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	exercise, exists := r.exercises[id]
	if !exists {
		return errors.New("exercise not found")
	}
	exercise.CompletionCount = count
	exercise.UpdatedAt = time.Now()
	r.exercises[id] = exercise
	return nil
}

// seedExercises adds sample exercise data for testing
func (r *InMemoryExerciseRepository) seedExercises() {
	now := time.Now()

	exercises := []domain.Exercise{
		{
			ID:                "ex-001",
			Name:              "Lip Rounding Practice",
			Description:       "Practice rounding your lips to improve sound articulation",
			Category:          domain.CategoryMouthExercise,
			Difficulty:        domain.DifficultyBeginner,
			ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationLips, domain.ArticulationJaw},
			Duration:          60 * time.Second,
			Repetitions:       10,
			Instructions:      "Round your lips as if saying 'ooh', hold for 3 seconds, then release. Repeat.",
			PhoneticGuidance:  "Start with a relaxed mouth, then pucker lips forward",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-002",
			Name:              "Tongue Tip Exercises",
			Description:       "Strengthen tongue tip for better articulation",
			Category:          domain.CategoryMouthExercise,
			Difficulty:        domain.DifficultyBeginner,
			ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationTongueTip, domain.ArticulationTeeth},
			Duration:          90 * time.Second,
			Repetitions:       15,
			Instructions:      "Touch tongue to the roof of your mouth behind your teeth, then to the back of your teeth. Repeat.",
			PhoneticGuidance:  "Focus on the tip of your tongue making contact",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-003",
			Name:              "She Sells Seashells",
			Description:       "Classic tongue twister for S sound practice",
			Category:          domain.CategoryTongueTwister,
			Difficulty:        domain.DifficultyIntermediate,
			TargetSounds:      []domain.SoundTarget{domain.SoundS},
			Duration:          30 * time.Second,
			Repetitions:       5,
			Instructions:      "Say 'She sells seashells by the seashore' slowly at first, then gradually increase speed.",
			PhoneticGuidance:  "Place tongue behind teeth, blow air between tongue and teeth",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-004",
			Name:              "Red Lorry Yellow Lorry",
			Description:       "Tongue twister for R and L sounds",
			Category:          domain.CategoryTongueTwister,
			Difficulty:        domain.DifficultyAdvanced,
			TargetSounds:      []domain.SoundTarget{domain.SoundR, domain.SoundL},
			Duration:          45 * time.Second,
			Repetitions:       5,
			Instructions:      "Say 'Red lorry yellow lorry' clearly and quickly.",
			PhoneticGuidance:  "R: curl tongue back, L: touch tongue to roof of mouth",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-005",
			Name:              "Clear Speech Strategy",
			Description:       "Focus on precise articulation for clearer speech",
			Category:          domain.CategoryDictionStrategy,
			Difficulty:        domain.DifficultyIntermediate,
			Duration:          120 * time.Second,
			Repetitions:       3,
			Instructions:      "Choose a paragraph and read it aloud, emphasizing each word clearly. Focus on enunciating each sound.",
			PhoneticGuidance:  "Take time with each syllable, ensure complete mouth formation",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-006",
			Name:              "Pause and Pacing",
			Description:       "Learn to control your pacing for better communication",
			Category:          domain.CategoryPacingStrategy,
			Difficulty:        domain.DifficultyBeginner,
			Duration:          90 * time.Second,
			Repetitions:       5,
			Instructions:      "Read a sentence and deliberately pause at each comma and period. Count to 2 at each pause.",
			PhoneticGuidance:  "Natural speech has pauses - use them to emphasize key points",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-007",
			Name:              "The Red Dragon",
			Description:       "Tongue twister for R sound mastery",
			Category:          domain.CategoryTongueTwister,
			Difficulty:        domain.DifficultyIntermediate,
			TargetSounds:      []domain.SoundTarget{domain.SoundR},
			Duration:          30 * time.Second,
			Repetitions:       5,
			Instructions:      "Say 'The red dragon is rather ragged' three times quickly.",
			PhoneticGuidance:  "Curl the tongue back and vibrate for the R sound",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-008",
			Name:              "Jaw Exercise",
			Description:       "Loosen jaw for better articulation",
			Category:          domain.CategoryMouthExercise,
			Difficulty:        domain.DifficultyBeginner,
			ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationJaw},
			Duration:          60 * time.Second,
			Repetitions:       10,
			Instructions:      "Open your mouth wide, hold for 2 seconds, then close. Then move jaw side to side.",
			PhoneticGuidance:  "Keep jaw relaxed, don't strain",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-009",
			Name:              "Sixth Sick Sheik",
			Description:       "Advanced S sound tongue twister",
			Category:          domain.CategoryTongueTwister,
			Difficulty:        domain.DifficultyAdvanced,
			TargetSounds:      []domain.SoundTarget{domain.SoundS, domain.SoundTH},
			Duration:          45 * time.Second,
			Repetitions:       3,
			Instructions:      "Say 'The sixth sick sheikh's sixth sheep's sick' clearly.",
			PhoneticGuidance:  "S: air between tongue and teeth, TH: tongue between teeth",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "ex-010",
			Name:              "Volume Control",
			Description:       "Practice varying your volume for emphasis",
			Category:          domain.CategoryDictionStrategy,
			Difficulty:        domain.DifficultyBeginner,
			Duration:          90 * time.Second,
			Repetitions:       5,
			Instructions:      "Read a sentence at whisper, then normal volume, then loud. Notice how your mouth changes.",
			PhoneticGuidance:  "Project from your diaphragm, not your throat",
			CompletionCount:   0,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	for _, e := range exercises {
		r.exercises[e.ID] = e
	}
}