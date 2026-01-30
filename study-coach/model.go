package main

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
