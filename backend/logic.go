package main

import (
	"math"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kc-clintone/study-coach/model"
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

// IsEducationalSubject checks if a subject is in the whitelist.
func IsEducationalSubject(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	_, ok := eduSubjects[s]
	return ok
}

// NewSession creates a new study session, initializing task metadata.
func NewSession(studentID, studentLevel, subject string, tasks []Task) *Session {
	for i := range tasks {
		if tasks[i].Complexity == 0 {
			tasks[i].Complexity = InferComplexity(tasks[i].Prompt, studentLevel)
		}
		tasks[i].EstimatedSecs = EstimateTimeForComplexity(tasks[i].Complexity)
		// extract topics
		tasks[i].Topics = ExtractTopics(tasks[i].Prompt)
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

// EstimateTimeForComplexity returns a per-task estimated seconds based on complexity and student level.
func EstimateTimeForComplexity(c int) int {
	base := 60 // 1 minute per complexity unit
	return int(math.Max(30, float64(base*c)))
}

// InferComplexity inspects a task prompt and returns a complexity score 1..5.
// This uses simple heuristics and is intentionally lightweight for the hackathon.
func InferComplexity(prompt, level string) int {
	p := strings.ToLower(prompt)
	if strings.Contains(p, "integral") || strings.Contains(p, "derivative") || strings.Contains(p, "prove") {
		return 5
	}
	if strings.Contains(p, "simplify") || strings.Contains(p, "factor") {
		return 3
	}
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

// ExtractTopics returns a small list of topic keywords using naive heuristics
func ExtractTopics(text string) []string {
	text = strings.ToLower(text)
	// split by non alpha
	fields := strings.FieldsFunc(text, func(r rune) bool { return !(r >= 'a' && r <= 'z') })
	// stopwords (simple English list)
	stop := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "that": true, "this": true,
		"from": true, "what": true, "when": true, "where": true, "how": true, "why": true,
		"are": true, "is": true, "a": true, "an": true, "of": true, "to": true, "in": true,
		"on": true, "by": true, "be": true, "as": true, "or": true, "it": true,
	}

	freq := map[string]int{}
	for _, f := range fields {
		if len(f) < 3 { // skip short words
			continue
		}
		if stop[f] {
			continue
		}
		s := simpleStem(f)
		freq[s]++
	}
	// select top 3
	type kv struct {
		k string
		v int
	}
	var arr []kv
	for k, v := range freq {
		arr = append(arr, kv{k, v})
	}
	// simple sort
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[j].v > arr[i].v {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
	out := []string{}
	for i := 0; i < len(arr) && i < 3; i++ {
		out = append(out, arr[i].k)
	}
	return out
}

// simpleStem performs a very small set of stemming rules (not a full porter stemmer)
func simpleStem(s string) string {
	if strings.HasSuffix(s, "ies") && len(s) > 4 {
		return s[:len(s)-3] + "y"
	}
	for _, suf := range []string{"ing", "ed", "es", "s"} {
		if strings.HasSuffix(s, suf) && len(s) > len(suf)+2 {
			return s[:len(s)-len(suf)]
		}
	}
	return s
}

// UpdateStudentTopicFrequencies increments incorrect-topic counts for a student
func UpdateStudentTopicFrequencies(studentID string, topics []string, incorrect bool) error {
	if !incorrect || len(topics) == 0 {
		return nil
	}
	m, err := LoadStudentTopics(studentID)
	if err != nil {
		return err
	}
	for _, t := range topics {
		m[t] = m[t] + 1
	}
	return SaveStudentTopics(studentID, m)
}

// GenerateStudyPlan analyzes a completed session and returns a StudyPlan. It adjusts
// next-timers based on actual performance: faster completions reduce suggested time,
// slower completions increase it. It also records topic frequencies and fills the
// Focus map with short recommendations. If Gemini is enabled (GEMINI_ENABLED=1) the
// function will call QueryGemini to produce a concise, human-friendly recommendation
// per topic; otherwise it uses a simple rule-based recommendation.
func GenerateStudyPlan(sess *model.Session) *model.StudyPlan {
	plan := &model.StudyPlan{StudentID: sess.StudentID, Subject: sess.Subject, CreatedAt: time.Now(), Focus: map[string]string{}, NextTimers: []int{}}
	// naive topic aggregation
	topics := map[string]int{}
	for _, t := range sess.Tasks {
		for _, tp := range t.Topics {
			topics[tp] = topics[tp] + 1
		}
		// adjust next timer based on actual performance
		newTimer := t.EstimatedSecs
		if t.ActualSeconds > 0 {
			if t.ActualSeconds < t.EstimatedSecs {
				// faster than estimate -> reduce next timer by 25%
				newTimer = int(float64(t.EstimatedSecs) * 0.75)
			} else if t.ActualSeconds > t.EstimatedSecs {
				// slower than estimate -> increase next timer by 25%
				newTimer = int(float64(t.EstimatedSecs) * 1.25)
			}
		}
		if newTimer < 15 {
			newTimer = 15
		}
		plan.NextTimers = append(plan.NextTimers, newTimer)
		// update topic frequencies if incorrect
		if t.Correct != nil && !*t.Correct {
			_ = UpdateStudentTopicFrequencies(sess.StudentID, t.Topics, true)
		}
	}
	// pick topics with most counts
	for k := range topics {
		plan.Focus[k] = "review topic, practice similar problems"
	}
	// attempt to enrich with AI if enabled
	if os.Getenv("GEMINI_ENABLED") == "1" {
		// Build prompt and call QueryGemini
		prompt := "Suggest one short recommendation per topic"
		if summary, err := QueryGemini(prompt); err == nil {
			for k := range plan.Focus {
				plan.Focus[k] = summary
			}
		}
	}
	return plan
}

// badgeCatalog defines metadata for available badges
var badgeCatalog = map[string]Badge{
	"First Steps":      {ID: "first_steps", Name: "First Steps", Icon: "🏁", Color: "#6ad", Description: "Completed your first session", Points: 10},
	"Accuracy Ace":     {ID: "accuracy_ace", Name: "Accuracy Ace", Icon: "🎯", Color: "#3a8", Description: "80%+ accuracy in a session", Points: 20},
	"Quick Solver":     {ID: "quick_solver", Name: "Quick Solver", Icon: "⚡", Color: "#f6a", Description: "Solved most tasks faster than estimate", Points: 15},
	"3-Day Streak":     {ID: "streak_3", Name: "3-Day Streak", Icon: "🔥", Color: "#f90", Description: "Active 3 days in a row", Points: 10},
	"7-Day Streak":     {ID: "streak_7", Name: "7-Day Streak", Icon: "🏆", Color: "#fc0", Description: "Active 7 days in a row", Points: 50},
	"Speed Demon":      {ID: "speed_demon", Name: "Speed Demon", Icon: "🚀", Color: "#a6f", Description: "All tasks completed faster than estimate", Points: 25},
	"Topic Apprentice": {ID: "topic_apprentice", Name: "Topic Apprentice", Icon: "📘", Color: "#48a", Description: "Practice a topic multiple times", Points: 15},
	"Topic Master":     {ID: "topic_master", Name: "Topic Master", Icon: "📜", Color: "#6c3", Description: "Mastered a topic (many improvements)", Points: 60},
}

// GetBadgeByName returns a copy of a Badge metadata by achievement name
func GetBadgeByName(name string) (Badge, bool) {
	b, ok := badgeCatalog[name]
	return b, ok
}
