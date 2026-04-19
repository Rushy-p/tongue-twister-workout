package infrastructure

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

// seedExercises populates the repository with all exercise content.
func (r *InMemoryExerciseRepository) seedExercises() {
	r.seedMouthExercises()
	r.seedTongueTwisters()
	r.seedDictionStrategies()
	r.seedPacingStrategies()
}

// seedMouthExercises loads all mouth and tongue exercises (me-001 through me-055).
func (r *InMemoryExerciseRepository) seedMouthExercises() {
	now := time.Now()
	exercises := []domain.Exercise{
		{ID: "me-001", Name: "Lip Rounding Practice", Description: "Practice rounding your lips to improve sound articulation", Category: domain.CategoryMouthExercise, Difficulty: domain.DifficultyBeginner, ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationLips, domain.ArticulationJaw}, Duration: 60 * time.Second, Repetitions: 10, Instructions: "Round your lips as if saying 'ooh', hold for 3 seconds, then release. Repeat.", PhoneticGuidance: "Start with a relaxed mouth, then pucker lips forward", CreatedAt: now, UpdatedAt: now},
		{ID: "me-002", Name: "Tongue Tip Exercises", Description: "Strengthen tongue tip for better articulation", Category: domain.CategoryMouthExercise, Difficulty: domain.DifficultyBeginner, ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationTongueTip, domain.ArticulationTeeth}, Duration: 90 * time.Second, Repetitions: 15, Instructions: "Touch tongue to the roof of your mouth behind your teeth, then to the back of your teeth. Repeat.", PhoneticGuidance: "Focus on the tip of your tongue making contact", CreatedAt: now, UpdatedAt: now},
		{ID: "me-003", Name: "Jaw Exercise", Description: "Loosen jaw for better articulation", Category: domain.CategoryMouthExercise, Difficulty: domain.DifficultyBeginner, ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationJaw}, Duration: 60 * time.Second, Repetitions: 10, Instructions: "Open your mouth wide, hold for 2 seconds, then close. Then move jaw side to side.", PhoneticGuidance: "Keep jaw relaxed, don't strain", CreatedAt: now, UpdatedAt: now},
		{ID: "me-004", Name: "Tongue Blade Workout", Description: "Strengthen tongue blade for advanced sounds", Category: domain.CategoryMouthExercise, Difficulty: domain.DifficultyIntermediate, ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationTongueBlade, domain.ArticulationTongueBack}, Duration: 90 * time.Second, Repetitions: 12, Instructions: "Press tongue blade against roof of mouth, hold, then release. Vary pressure.", PhoneticGuidance: "Focus on the middle of your tongue", CreatedAt: now, UpdatedAt: now},
		{ID: "me-005", Name: "Soft Palate Lift", Description: "Exercise the soft palate for better resonance", Category: domain.CategoryMouthExercise, Difficulty: domain.DifficultyAdvanced, ArticulationPoints: []domain.ArticulationPoint{domain.ArticulationSoftPalate}, Duration: 120 * time.Second, Repetitions: 8, Instructions: "Yawn to feel the soft palate lift. Say 'ah' with raised soft palate. Hold and release.", PhoneticGuidance: "Imagine the sound coming from the back of your throat", CreatedAt: now, UpdatedAt: now},
	}
	for _, e := range exercises {
		r.exercises[e.ID] = e
	}
	for _, e := range seedMouthExercisesExtended(now) {
		r.exercises[e.ID] = e
	}
}

// seedTongueTwisters loads all tongue twister exercises (tt-001 through tt-100+).
func (r *InMemoryExerciseRepository) seedTongueTwisters() {
	now := time.Now()
	exercises := []domain.Exercise{
		{ID: "tt-001", Name: "She Sells Seashells", Description: "Classic tongue twister for S sound practice", Category: domain.CategoryTongueTwister, Difficulty: domain.DifficultyIntermediate, TargetSounds: []domain.SoundTarget{domain.SoundS}, Duration: 30 * time.Second, Repetitions: 5, Instructions: "Say 'She sells seashells by the seashore' slowly at first, then gradually increase speed.", PhoneticGuidance: "Place tongue behind teeth, blow air between tongue and teeth", CreatedAt: now, UpdatedAt: now},
		{ID: "tt-002", Name: "Red Lorry Yellow Lorry", Description: "Tongue twister for R and L sounds", Category: domain.CategoryTongueTwister, Difficulty: domain.DifficultyAdvanced, TargetSounds: []domain.SoundTarget{domain.SoundR, domain.SoundL}, Duration: 45 * time.Second, Repetitions: 5, Instructions: "Say 'Red lorry yellow lorry' clearly and quickly.", PhoneticGuidance: "R: curl tongue back, L: touch tongue to roof of mouth", CreatedAt: now, UpdatedAt: now},
		{ID: "tt-003", Name: "The Red Dragon", Description: "Tongue twister for R sound mastery", Category: domain.CategoryTongueTwister, Difficulty: domain.DifficultyIntermediate, TargetSounds: []domain.SoundTarget{domain.SoundR}, Duration: 30 * time.Second, Repetitions: 5, Instructions: "Say 'The red dragon is rather ragged' three times quickly.", PhoneticGuidance: "Curl the tongue back and vibrate for the R sound", CreatedAt: now, UpdatedAt: now},
		{ID: "tt-004", Name: "Sixth Sick Sheik", Description: "Advanced S sound tongue twister", Category: domain.CategoryTongueTwister, Difficulty: domain.DifficultyAdvanced, TargetSounds: []domain.SoundTarget{domain.SoundS, domain.SoundTH}, Duration: 45 * time.Second, Repetitions: 3, Instructions: "Say 'The sixth sick sheikh's sixth sheep's sick' clearly.", PhoneticGuidance: "S: air between tongue and teeth, TH: tongue between teeth", CreatedAt: now, UpdatedAt: now},
		{ID: "tt-005", Name: "Unique New York", Description: "Tongue twister for J/Y sounds", Category: domain.CategoryTongueTwister, Difficulty: domain.DifficultyBeginner, TargetSounds: []domain.SoundTarget{domain.SoundJ}, Duration: 30 * time.Second, Repetitions: 5, Instructions: "Say 'Unique New York, unique New York, you know you need unique New York' clearly.", PhoneticGuidance: "Focus on the Y sound at the start of unique", CreatedAt: now, UpdatedAt: now},
	}
	for _, e := range exercises {
		r.exercises[e.ID] = e
	}
	for _, e := range seedTongueTwistersExtended(now) {
		r.exercises[e.ID] = e
	}
}

// seedDictionStrategies loads all diction strategy exercises (ds-001 through ds-012+).
func (r *InMemoryExerciseRepository) seedDictionStrategies() {
	now := time.Now()
	exercises := []domain.Exercise{
		{ID: "ds-001", Name: "Clear Speech Strategy", Description: "Focus on precise articulation for clearer speech", Category: domain.CategoryDictionStrategy, Difficulty: domain.DifficultyIntermediate, Duration: 120 * time.Second, Repetitions: 3, Instructions: "Choose a paragraph and read it aloud, emphasizing each word clearly. Focus on enunciating each sound.", PhoneticGuidance: "Take time with each syllable, ensure complete mouth formation", CreatedAt: now, UpdatedAt: now},
		{ID: "ds-002", Name: "Volume Control", Description: "Practice varying your volume for emphasis", Category: domain.CategoryDictionStrategy, Difficulty: domain.DifficultyBeginner, Duration: 90 * time.Second, Repetitions: 5, Instructions: "Read a sentence at whisper, then normal volume, then loud. Notice how your mouth changes.", PhoneticGuidance: "Project from your diaphragm, not your throat", CreatedAt: now, UpdatedAt: now},
		{ID: "ds-003", Name: "Precision Enunciation", Description: "Practice precise mouth positions for each sound", Category: domain.CategoryDictionStrategy, Difficulty: domain.DifficultyAdvanced, Duration: 150 * time.Second, Repetitions: 3, Instructions: "Select 10 words and say each one slowly, exaggerating each sound. Record and playback.", PhoneticGuidance: "Each sound has a specific mouth position - find it", CreatedAt: now, UpdatedAt: now},
		{ID: "ds-004", Name: "Resonance Building", Description: "Develop proper resonance for richer voice", Category: domain.CategoryDictionStrategy, Difficulty: domain.DifficultyIntermediate, Duration: 100 * time.Second, Repetitions: 4, Instructions: "Hum at different pitches, feeling vibration in your face. Then apply to speech.", PhoneticGuidance: "Feel the vibration in your nasal cavity and face", CreatedAt: now, UpdatedAt: now},
		{ID: "ds-005", Name: "Breath Control Basics", Description: "Learn to control breath for sustained speech", Category: domain.CategoryDictionStrategy, Difficulty: domain.DifficultyBeginner, Duration: 80 * time.Second, Repetitions: 5, Instructions: "Take a deep breath and try to speak on one breath for as long as possible. Count how many words you can say.", PhoneticGuidance: "Breathe from diaphragm, not chest", CreatedAt: now, UpdatedAt: now},
	}
	for _, e := range exercises {
		r.exercises[e.ID] = e
	}
	for _, e := range seedDictionStrategiesExtended(now) {
		r.exercises[e.ID] = e
	}
}

// seedPacingStrategies loads all pacing strategy exercises (ps-001 through ps-010+).
func (r *InMemoryExerciseRepository) seedPacingStrategies() {
	now := time.Now()
	exercises := []domain.Exercise{
		{ID: "ps-001", Name: "Pause and Pacing", Description: "Learn to control your pacing for better communication", Category: domain.CategoryPacingStrategy, Difficulty: domain.DifficultyBeginner, Duration: 90 * time.Second, Repetitions: 5, Instructions: "Read a sentence and deliberately pause at each comma and period. Count to 2 at each pause.", PhoneticGuidance: "Natural speech has pauses - use them to emphasize key points", CreatedAt: now, UpdatedAt: now},
		{ID: "ps-002", Name: "Syllable Timing", Description: "Practice even syllable timing for clearer speech", Category: domain.CategoryPacingStrategy, Difficulty: domain.DifficultyIntermediate, Duration: 100 * time.Second, Repetitions: 4, Instructions: "Say a sentence with equal emphasis on each syllable. Then compare to natural speech.", PhoneticGuidance: "Metronome can help - set to 60 BPM and speak one syllable per beat", CreatedAt: now, UpdatedAt: now},
		{ID: "ps-003", Name: "Phrase Grouping", Description: "Learn to group words into meaningful phrases", Category: domain.CategoryPacingStrategy, Difficulty: domain.DifficultyIntermediate, Duration: 90 * time.Second, Repetitions: 5, Instructions: "Mark phrases in a text with slashes. Read each phrase as a single unit.", PhoneticGuidance: "Phrases are thought units - complete one before moving on", CreatedAt: now, UpdatedAt: now},
		{ID: "ps-004", Name: "Emphasis Patterns", Description: "Practice placing emphasis on key words", Category: domain.CategoryPacingStrategy, Difficulty: domain.DifficultyAdvanced, Duration: 120 * time.Second, Repetitions: 4, Instructions: "Say the same sentence with emphasis on different words. Notice how meaning changes.", PhoneticGuidance: "Emphasized words are usually content words (nouns, verbs, adjectives)", CreatedAt: now, UpdatedAt: now},
		{ID: "ps-005", Name: "Slow Motion Practice", Description: "Practice at slow speed to build muscle memory", Category: domain.CategoryPacingStrategy, Difficulty: domain.DifficultyBeginner, Duration: 120 * time.Second, Repetitions: 3, Instructions: "Choose a paragraph and read it at half your normal speed. Focus on each sound.", PhoneticGuidance: "Slow down until you can say every sound clearly", CreatedAt: now, UpdatedAt: now},
	}
	for _, e := range exercises {
		r.exercises[e.ID] = e
	}
	for _, e := range seedPacingStrategiesExtended(now) {
		r.exercises[e.ID] = e
	}
}
