package handler

import (
	"context"
	"log"
	"net/http"
	"time"
)

// ContextKey type for context values
type ContextKey string

const startTimeKey ContextKey = "startTime"

// LoggingMiddleware logs HTTP requests with timing information
// Validates: Requirements 11.1, 11.2 (request timing tracking)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Add start time to context for handlers to access
		ctx := context.WithValue(r.Context(), startTimeKey, start)
		
		// Log request
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// Call next handler
		next.ServeHTTP(w, r.WithContext(ctx))
		
		// Log completion with timing
		duration := time.Since(start)
		log.Printf("Request completed in %v", duration)
		
		// Log slow requests (potential performance issues)
		if duration > 1*time.Second {
			log.Printf("WARNING: Slow request detected: %s took %v", r.URL.Path, duration)
		}
	})
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(next http.Handler) http.Handler {
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

// MethodMiddleware ensures the request uses the expected method
func MethodMiddleware(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expected {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}

// Chain applies multiple middleware to a handler
func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}