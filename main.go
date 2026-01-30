
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// User represents the account structure
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Name     string `json:"name"`
}

// AuthResponse for login/register
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func main() {
	// Middleware for CORS
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Routes
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/quiz/save", handleSaveQuiz)

	fmt.Println("EduPulse Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux)))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Mock logic
	resp := AuthResponse{
		Token: "mock-jwt-token",
		User: User{ID: "1", Email: "student@example.com", Name: "Demo Student"},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	// Registration logic...
	w.WriteHeader(http.StatusCreated)
}

func handleSaveQuiz(w http.ResponseWriter, r *http.Request) {
	// Persist quiz results to database...
	w.WriteHeader(http.StatusOK)
}
