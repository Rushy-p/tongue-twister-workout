package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"speech-practice-app/internal/handler"
	"speech-practice-app/internal/infrastructure"
	"speech-practice-app/internal/service"
)

func main() {
	// Initialize file storage (Req 14.1 — store all data locally)
	storage, err := infrastructure.NewFileStorage("./data")
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	// Initialize repositories — file-backed for persistence, in-memory for exercises
	exerciseRepo := infrastructure.NewInMemoryExerciseRepository()

	sessionRepo, err := infrastructure.NewFileSessionRepository(storage)
	if err != nil {
		log.Fatalf("Failed to init session repository: %v", err)
	}

	progressRepo, err := infrastructure.NewFileProgressRepository(storage)
	if err != nil {
		log.Fatalf("Failed to init progress repository: %v", err)
	}

	preferencesRepo, err := infrastructure.NewFilePreferencesRepository(storage)
	if err != nil {
		log.Fatalf("Failed to init preferences repository: %v", err)
	}

	recommendationRepo := infrastructure.NewInMemoryRecommendationRepository()

	// Initialize services
	exerciseService := service.NewExerciseService(exerciseRepo, progressRepo, sessionRepo)
	sessionService := service.NewSessionService(sessionRepo, progressRepo, exerciseRepo)
	progressService := service.NewProgressService(progressRepo, sessionRepo, exerciseRepo)
	preferencesService := service.NewPreferencesService(preferencesRepo)
	recommendationService := service.NewRecommendationService(exerciseRepo, progressRepo, sessionRepo, recommendationRepo)

	// Create router with all services
	router := handler.NewRouter(
		exerciseService,
		sessionService,
		progressService,
		preferencesService,
		recommendationService,
	)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Visit http://localhost:%s", port)

	// Start server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
