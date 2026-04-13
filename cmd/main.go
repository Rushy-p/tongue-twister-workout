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
	// Initialize repositories
	exerciseRepo := infrastructure.NewInMemoryExerciseRepository()
	sessionRepo := infrastructure.NewInMemorySessionRepository()
	progressRepo := infrastructure.NewInMemoryProgressRepository()
	preferencesRepo := infrastructure.NewInMemoryPreferencesRepository()

	// Initialize services
	exerciseService := service.NewExerciseService(exerciseRepo, progressRepo, sessionRepo)
	sessionService := service.NewSessionService(sessionRepo, progressRepo, exerciseRepo)
	progressService := service.NewProgressService(progressRepo, sessionRepo, exerciseRepo)
	preferencesService := service.NewPreferencesService(preferencesRepo)

	// Create router with all services
	router := handler.NewRouter(
		exerciseService,
		sessionService,
		progressService,
		preferencesService,
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