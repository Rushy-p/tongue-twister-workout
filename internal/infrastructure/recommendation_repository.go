package infrastructure

import (
	"errors"
	"sync"
	"time"

	"speech-practice-app/internal/domain"
)

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
	recommendations         map[string]domain.RecommendationRecord
	rejectedRecommendations map[string]map[string]domain.RejectedRecommendation
	acceptedRecommendations map[string]map[string]domain.AcceptedRecommendation
	mu                      sync.RWMutex
}

// NewInMemoryRecommendationRepository creates a new in-memory recommendation repository
func NewInMemoryRecommendationRepository() *InMemoryRecommendationRepository {
	return &InMemoryRecommendationRepository{
		recommendations:         make(map[string]domain.RecommendationRecord),
		rejectedRecommendations: make(map[string]map[string]domain.RejectedRecommendation),
		acceptedRecommendations: make(map[string]map[string]domain.AcceptedRecommendation),
	}
}

// SaveRecommendation saves a recommendation record
func (r *InMemoryRecommendationRepository) SaveRecommendation(rec *domain.RecommendationRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rec.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if rec.ExerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

	rec.CreatedAt = time.Now()
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

	if rec.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if rec.ExerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

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

	if rec.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if rec.ExerciseID == "" {
		return errors.New("exercise ID cannot be empty")
	}

	if r.acceptedRecommendations[rec.UserID] == nil {
		r.acceptedRecommendations[rec.UserID] = make(map[string]domain.AcceptedRecommendation)
	}
	r.acceptedRecommendations[rec.UserID][rec.ExerciseID] = *rec
	return nil
}