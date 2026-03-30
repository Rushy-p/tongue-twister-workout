package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// logger is a middleware that logs HTTP requests
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Request completed in %v", time.Since(start))
	})
}

// recovery is a middleware that recovers from panics
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// cors adds CORS headers to responses
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create router
	mux := http.NewServeMux()
	
	// Register routes
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/exercises", handleExercises)
	mux.HandleFunc("/exercises/", handleExerciseDetail)
	mux.HandleFunc("/session", handleSession)
	mux.HandleFunc("/session/start", handleSessionStart)
	mux.HandleFunc("/session/complete", handleSessionComplete)
	mux.HandleFunc("/progress", handleProgress)
	mux.HandleFunc("/preferences", handlePreferences)
	mux.HandleFunc("/preferences/update", handlePreferencesUpdate)
	mux.HandleFunc("/recommendations", handleRecommendations)
	
	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	
	// Templates
	mux.HandleFunc("/template/", handleTemplate)
	
	// Apply middleware
	handler := cors(logger(recovery(mux)))
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	log.Printf("Visit http://localhost:%s", port)
	
	// Start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

// handleIndex handles the home page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Speech Practice App</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Welcome to Speech Practice</h1>
		<p>Improve your speech clarity and fluency through daily practice.</p>
		<div class="actions">
			<a href="/exercises" class="btn btn-primary">Browse Exercises</a>
			<a href="/session" class="btn btn-secondary">Start Practice</a>
		</div>
	</main>
</body>
</html>`))
}

// handleExercises handles the exercise library page
func handleExercises(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Exercise Library - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Exercise Library</h1>
		<p>Choose a category to begin practicing.</p>
		<div class="categories">
			<a href="/exercises/mouth" class="card">
				<h2>Mouth Exercises</h2>
				<p>Articulation and mouth positioning exercises</p>
			</a>
			<a href="/exercises/twisters" class="card">
				<h2>Tongue Twisters</h2>
				<p>Practice specific sounds with fun phrases</p>
			</a>
			<a href="/exercises/diction" class="card">
				<h2>Diction Strategies</h2>
				<p>Improve clarity and precision in speech</p>
			</a>
			<a href="/exercises/pacing" class="card">
				<h2>Pacing Strategies</h2>
				<p>Control speed and rhythm of your speech</p>
			</a>
		</div>
	</main>
</body>
</html>`))
}

// handleExerciseDetail handles individual exercise pages
func handleExerciseDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Exercise Detail - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Exercise Detail</h1>
		<p>Exercise details will be displayed here.</p>
		<a href="/exercises" class="btn btn-secondary">Back to Library</a>
	</main>
</body>
</html>`))
}

// handleSession handles the practice session page
func handleSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Practice Session - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Practice Session</h1>
		<p>Start a new practice session or resume a saved one.</p>
		<div class="actions">
			<a href="/session/start" class="btn btn-primary">Start New Session</a>
		</div>
	</main>
</body>
</html>`))
}

// handleSessionStart handles starting a new session
func handleSessionStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Session Started - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Session in Progress</h1>
		<p>Practice your exercises. Timer will be displayed here.</p>
		<a href="/session/complete" class="btn btn-primary">Complete Session</a>
	</main>
</body>
</html>`))
}

// handleSessionComplete handles completing a session
func handleSessionComplete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Session Complete - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Session Complete!</h1>
		<p>Great job! Your session summary will be displayed here.</p>
		<a href="/" class="btn btn-primary">Return Home</a>
	</main>
</body>
</html>`))
}

// handleProgress handles the progress page
func handleProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Progress - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Your Progress</h1>
		<p>Track your practice history and achievements.</p>
		<div class="stats">
			<div class="stat-card">
				<h3>Current Streak</h3>
				<p class="stat-value">0 days</p>
			</div>
			<div class="stat-card">
				<h3>Total Exercises</h3>
				<p class="stat-value">0</p>
			</div>
			<div class="stat-card">
				<h3>Total Practice Time</h3>
				<p class="stat-value">0 minutes</p>
			</div>
		</div>
	</main>
</body>
</html>`))
}

// handlePreferences handles the preferences page
func handlePreferences(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Settings - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Settings</h1>
		<p>Customize your practice experience.</p>
		<form action="/preferences/update" method="POST">
			<div class="form-group">
				<label for="difficulty">Difficulty Level</label>
				<select id="difficulty" name="difficulty">
					<option value="beginner">Beginner</option>
					<option value="intermediate">Intermediate</option>
					<option value="advanced">Advanced</option>
				</select>
			</div>
			<div class="form-group">
				<label for="duration">Default Duration</label>
				<select id="duration" name="duration">
					<option value="30">30 seconds</option>
					<option value="60" selected>60 seconds</option>
					<option value="90">90 seconds</option>
					<option value="120">120 seconds</option>
				</select>
			</div>
			<div class="form-group">
				<label>
					<input type="checkbox" name="audio" checked> Enable Audio Feedback
				</label>
			</div>
			<button type="submit" class="btn btn-primary">Save Settings</button>
		</form>
	</main>
</body>
</html>`))
}

// handlePreferencesUpdate handles updating preferences
func handlePreferencesUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Preferences would be saved here
		log.Printf("Preferences updated")
	}
	http.Redirect(w, r, "/preferences", http.StatusSeeOther)
}

// handleRecommendations handles the recommendations page
func handleRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Recommendations - Speech Practice</title>
	<link rel="stylesheet" href="/static/style.css">
</head>
<body>
	<header>
		<nav>
			<a href="/">Home</a>
			<a href="/exercises">Exercises</a>
			<a href="/session">Practice</a>
			<a href="/progress">Progress</a>
			<a href="/preferences">Settings</a>
		</nav>
	</header>
	<main>
		<h1>Daily Recommendations</h1>
		<p>Personalized exercises based on your practice history.</p>
		<p>Complete more exercises to receive recommendations.</p>
	</main>
</body>
</html>`))
}

// handleTemplate handles template requests
func handleTemplate(w http.ResponseWriter, r *http.Request) {
	// Template serving would be implemented here
	http.NotFound(w, r)
}