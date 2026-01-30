package main

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// simplistic subject whitelist
var eduSubjects = map[string]bool{
	"math":      true,
	"physics":   true,
	"chemistry": true,
	"biology":   true,
	"history":   true,
	"english":   true,
}

func IsEducationalSubject(s string) bool {
	_, ok := eduSubjects[s]
	return ok
}

func NewSession(studentID, studentLevel, subject string, tasks []Task) *Session {
	for i := range tasks {
		if tasks[i].Complexity == 0 {
			tasks[i].Complexity = InferComplexity(tasks[i].Prompt, studentLevel)
		}
		tasks[i].EstimatedSecs = EstimateTimeForComplexity(tasks[i].Complexity)
	}
	return &Session{
		ID:           uuid.NewString(),
		StudentID:    studentID,
		StudentLevel: studentLevel,
		Subject:      subject,
		Tasks:        tasks,
		CreatedAt:    time.Now(),
	}
}

func EstimateTimeForComplexity(c int) int {
	base := 60 // 1 minute per complexity unit
	return int(math.Max(30, float64(base*c)))
}

func InferComplexity(prompt, level string) int {
	// naive heuristics for hackathon: length-based
	l := len(prompt)
	switch {
	case l < 40:
		return 1
	case l < 120:
		return 2
	case l < 300:
		return 3
	case l < 600:
		return 4
	default:
		return 5
	}
}

func GenerateStudyPlan(s *Session) *StudyPlan {
	plan := &StudyPlan{
		StudentID: s.StudentID,
		Subject:   s.Subject,
		CreatedAt: time.Now(),
		Focus:     map[string]string{},
	}
	var nextTimers []int
	for _, t := range s.Tasks {
		// if incorrect, add focus area
		if t.Correct != nil && !*t.Correct {
			plan.Focus[t.Prompt] = "review topic, practice similar problems"
		}
		// adjust next timer based on speed: if student was faster than estimate, reduce time, else increase
		est := t.EstimatedSecs
		actual := t.ActualSeconds
		if actual == 0 {
			actual = est
		}
		ratio := float64(actual) / float64(est)
		var adj float64 = 1.0
		if ratio < 0.8 {
			adj = 0.9
		} else if ratio > 1.2 {
			adj = 1.15
		}
		next := int(math.Round(float64(est) * adj))
		nextTimers = append(nextTimers, next)
	}
	plan.NextTimers = nextTimers
	return plan
}
