package main

import (
	"math"
	"strings"
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
	s = strings.ToLower(strings.TrimSpace(s))
	_, ok := eduSubjects[s]
	return ok
}

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
// next-timers based on actual performance, records topic frequencies, and fills the
// Focus map with short recommendations. If Gemini is enabled (GEMINI_ENABLED=1) the
// function will call QueryGemini to produce a concise, human-friendly recommendation
// per topic; otherwise it uses a simple rule-based recommendation.
func GenerateStudyPlan(s *Session) *StudyPlan {
	plan := &StudyPlan{
		StudentID: s.StudentID,
		Subject:   s.Subject,
		CreatedAt: time.Now(),
		Focus:     map[string]string{},
	}
	var nextTimers []int
	for _, t := range s.Tasks {
		// if incorrect, add focus area and update student topic frequencies
		if t.Correct != nil && !*t.Correct {
			// use topic if available, else prompt snippet
			key := "general"
			if len(t.Topics) > 0 {
				key = t.Topics[0]
			}
			plan.Focus[key] = "review topic, practice similar problems"
			_ = UpdateStudentTopicFrequencies(s.StudentID, t.Topics, true)
		}
		// adjust next timer based on speed
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
	// integrate student-wide topic frequencies into focus (top 3)
	if m, err := LoadStudentTopics(s.StudentID); err == nil {
		type kv struct {
			k string
			v int
		}
		var arr []kv
		for k, v := range m {
			if v > 0 {
				arr = append(arr, kv{k, v})
			}
		}
		for i := 0; i < len(arr); i++ {
			for j := i + 1; j < len(arr); j++ {
				if arr[j].v > arr[i].v {
					arr[i], arr[j] = arr[j], arr[i]
				}
			}
		}
		for i := 0; i < len(arr) && i < 3; i++ {
			if _, ok := plan.Focus[arr[i].k]; !ok {
				plan.Focus[arr[i].k] = "practice more problems in this topic"
			}
		}
	}

	// If Gemini is enabled, attempt to enrich each focus entry with an AI-generated
	// recommendation. This is gated by GEMINI_ENABLED and will behave as a no-op when
	// disabled, keeping the function deterministic for tests.
	for topic := range plan.Focus {
		// build a short prompt for the model
		prompt := "You are an educational assistant. Given a student who struggled or needs practice in the topic '" + topic + "', provide a single concise (one-sentence) actionable recommendation and one short practice suggestion."
		if summary, err := QueryGemini(prompt); err == nil {
			// use the model output as the recommendation (trim to reasonable length)
			if len(summary) > 0 {
				if len(summary) > 300 {
					plan.Focus[topic] = summary[:300]
				} else {
					plan.Focus[topic] = summary
				}
			}
		}
	}

	plan.NextTimers = nextTimers
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
