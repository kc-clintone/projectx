package model

import "time"

// Task represents a single question or assignment
type Task struct {
	ID            string    `json:"id"`
	Prompt        string    `json:"prompt"`
	EstimatedSecs int       `json:"estimated_secs"`
	ActualSeconds int       `json:"actual_seconds,omitempty"`
	Complexity    int       `json:"complexity"` // 1..5
	Correct       *bool     `json:"correct,omitempty"`
	SubmittedAt   time.Time `json:"submitted_at,omitempty"`
	Topics        []string  `json:"topics,omitempty"`
}

type Session struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	StudentLevel string    `json:"student_level"`
	Subject      string    `json:"subject"`
	Tasks        []Task    `json:"tasks"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateSessionRequest incoming payload
type CreateSessionRequest struct {
	StudentID    string `json:"student_id"`
	StudentLevel string `json:"student_level"`
	Subject      string `json:"subject"`
	Tasks        []Task `json:"tasks"`
}

// Submission contains user's results
type Submission struct {
	SessionID string `json:"session_id"`
	Results   []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	} `json:"results"`
}

// StudyPlan is a simple plan returned after analysis
type StudyPlan struct {
	StudentID  string            `json:"student_id"`
	Subject    string            `json:"subject"`
	CreatedAt  time.Time         `json:"created_at"`
	Focus      map[string]string `json:"focus"`       // topic -> recommendation
	NextTimers []int             `json:"next_timers"` // per-task suggested seconds
}

// TopicSnapshot is a timestamped snapshot of a student's topic frequencies
type TopicSnapshot struct {
	Timestamp time.Time      `json:"timestamp"`
	Topics    map[string]int `json:"topics"`
}

// Badge contains metadata for a badge/achievement
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
	Points      int    `json:"points"`
}

// User represents an authenticated user account stored locally.
// PasswordHash should contain a bcrypt hash of the user's password.
type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email,omitempty"`
	StudentLevel string    `json:"student_level,omitempty"` // e.g., grade10
	CreatedAt    time.Time `json:"created_at"`
}

// ProfileResponse returned by the student profile endpoint
type ProfileResponse struct {
	StudentID    string             `json:"student_id"`
	Topics       map[string]int     `json:"topics"`
	Achievements []string           `json:"achievements"`
	EarnedBadges []Badge            `json:"earned_badges"`
	Points       int                `json:"points"`
	Progress     map[string]int     `json:"progress_percent"`
	WeeklyRecap  []TopicImprovement `json:"weekly_recap"`
	// new fields for gamification
	Streak     int       `json:"streak"`
	LastActive time.Time `json:"last_active,omitempty"`
	// Summary is an optional concise AI-generated summary of strengths/weaknesses
	Summary string `json:"summary,omitempty"`
}

type TopicImprovement struct {
	Topic      string `json:"topic"`
	ImprovedBy int    `json:"improved_by"`
}
