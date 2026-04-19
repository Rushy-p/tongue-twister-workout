package domain

import (
	"time"
)

// ExerciseCategory represents the type of exercise
type ExerciseCategory string

const (
	CategoryMouthExercise ExerciseCategory = "mouth_exercise"
	CategoryTongueTwister ExerciseCategory = "tongue_twister"
	CategoryDictionStrategy ExerciseCategory = "diction_strategy"
	CategoryPacingStrategy  ExerciseCategory = "pacing_strategy"
)

// DifficultyLevel represents the difficulty of an exercise
type DifficultyLevel string

const (
	DifficultyBeginner     DifficultyLevel = "beginner"
	DifficultyIntermediate DifficultyLevel = "intermediate"
	DifficultyAdvanced     DifficultyLevel = "advanced"
)

// ArticulationPoint represents the target area for mouth exercises
type ArticulationPoint string

const (
	ArticulationLips       ArticulationPoint = "lips"
	ArticulationTeeth      ArticulationPoint = "teeth"
	ArticulationTongueTip  ArticulationPoint = "tongue_tip"
	ArticulationTongueBlade ArticulationPoint = "tongue_blade"
	ArticulationTongueBack ArticulationPoint = "tongue_back"
	ArticulationSoftPalate ArticulationPoint = "soft_palate"
	ArticulationJaw        ArticulationPoint = "jaw"
)

// SoundTarget represents the target sound for tongue twisters
type SoundTarget string

const (
	SoundS  SoundTarget = "s"
	SoundZ  SoundTarget = "z"
	SoundR  SoundTarget = "r"
	SoundL  SoundTarget = "l"
	SoundTH SoundTarget = "th"
	SoundSH SoundTarget = "sh"
	SoundCH SoundTarget = "ch"
	SoundJ  SoundTarget = "j"
	SoundK  SoundTarget = "k"
	SoundG  SoundTarget = "g"
)

// FocusArea represents the focus area for strategies
type FocusArea string

const (
	// For diction strategies
	FocusClarity    FocusArea = "clarity"
	FocusPrecision  FocusArea = "precision"
	FocusVolume     FocusArea = "volume"
	FocusResonance  FocusArea = "resonance"
	// For pacing strategies
	FocusPausePlacement   FocusArea = "pause_placement"
	FocusSyllableTiming   FocusArea = "syllable_timing"
	FocusPhraseGrouping   FocusArea = "phrase_grouping"
	FocusEmphasisPatterns FocusArea = "emphasis_patterns"
	// Additional focus areas
	FocusBreathControl FocusArea = "breath_control"
	FocusMouthOpening  FocusArea = "mouth_opening"
	FocusTonguePlacement FocusArea = "tongue_placement"
	FocusLipRounding   FocusArea = "lip_rounding"
)

// Exercise represents a single practice exercise
type Exercise struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Description         string              `json:"description"`
	Category            ExerciseCategory    `json:"category"`
	Difficulty          DifficultyLevel     `json:"difficulty"`
	TargetSounds        []SoundTarget       `json:"target_sounds,omitempty"`
	ArticulationPoints  []ArticulationPoint `json:"articulation_points,omitempty"`
	Duration            time.Duration       `json:"duration"`
	Repetitions         int                 `json:"repetitions"`
	Instructions        string              `json:"instructions"`
	PhoneticGuidance    string              `json:"phonetic_guidance,omitempty"`
	AudioURL            string              `json:"audio_url,omitempty"`
	CompletionCount     int                 `json:"completion_count"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

// TongueTwister extends Exercise with tongue twister specific fields
type TongueTwister struct {
	Exercise
	TargetSound         SoundTarget `json:"target_sound"`
	PhoneticTranscription string    `json:"phonetic_transcription"`
	HighlightedSounds   []string    `json:"highlighted_sounds"`
	DifficultyScore     int         `json:"difficulty_score"`
	PopularityRating    int         `json:"popularity_rating"`
}

// Strategy represents a diction or pacing strategy
type Strategy struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Category        ExerciseCategory `json:"category"`
	FocusArea       FocusArea   `json:"focus_area"`
	ExamplePhrases  []string    `json:"example_phrases"`
	PracticePassages []string   `json:"practice_passages"`
	AudioExamples   []string    `json:"audio_examples"`
	MasteryLevel    int         `json:"mastery_level"`
	Instructions    string      `json:"instructions"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// IsDictionStrategy returns true if the strategy is a diction strategy
func (s *Strategy) IsDictionStrategy() bool {
	return s.Category == CategoryDictionStrategy
}

// IsPacingStrategy returns true if the strategy is a pacing strategy
func (s *Strategy) IsPacingStrategy() bool {
	return s.Category == CategoryPacingStrategy
}

// Validate checks if the exercise has valid data
func (e *Exercise) Validate() error {
	if e.ID == "" {
		return ErrInvalidExerciseID
	}
	if e.Name == "" {
		return ErrInvalidExerciseName
	}
	if e.Category == "" {
		return ErrInvalidCategory
	}
	if e.Difficulty == "" {
		return ErrInvalidDifficulty
	}
	return nil
}

// GetDurationSeconds returns the duration in seconds
func (e Exercise) GetDurationSeconds() int {
	return int(e.Duration.Seconds())
}

// IsCompleted returns true if the exercise has been completed at least once
func (e Exercise) IsCompleted() bool {
	return e.CompletionCount > 0
}

// Domain errors for exercise validation
var (
	ErrInvalidExerciseID    = &ValidationError{"exercise ID cannot be empty"}
	ErrInvalidExerciseName  = &ValidationError{"exercise name cannot be empty"}
	ErrInvalidCategory      = &ValidationError{"exercise category cannot be empty"}
	ErrInvalidDifficulty    = &ValidationError{"difficulty level cannot be empty"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}