package coach

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kc-clintone/study-coach/ai"
	"github.com/kc-clintone/study-coach/model"
	"github.com/kc-clintone/study-coach/storage"
)

// IsEducationalSubject returns true for supported subject keywords.
func IsEducationalSubject(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	supported := map[string]bool{"math": true, "physics": true, "chemistry": true, "biology": true, "history": true, "english": true}
	return supported[s]
}

// NewSession prepares a Session with inferred complexity, estimated times and topics.
func NewSession(studentID, studentLevel, subject string, tasks []model.Task) *model.Session {
	for i := range tasks {
		if tasks[i].Complexity == 0 {
			tasks[i].Complexity = InferComplexity(tasks[i].Prompt, studentLevel)
		}
		tasks[i].EstimatedSecs = EstimateTimeForComplexity(tasks[i].Complexity)
		// extract topics
		tasks[i].Topics = ExtractTopics(tasks[i].Prompt)
	}
	return &model.Session{
		ID:           uuid.NewString(),
		StudentID:    studentID,
		StudentLevel: studentLevel,
		Subject:      subject,
		Tasks:        tasks,
		CreatedAt:    time.Now(),
	}
}

// EstimateTimeForComplexity converts a complexity score to seconds.
func EstimateTimeForComplexity(c int) int {
	base := 60
	return int(math.Max(30, float64(base*c)))
}

// InferComplexity uses simple heuristics (prompt length) to infer complexity.
func InferComplexity(prompt, level string) int {
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

// ExtractTopics applies tokenization, stopword removal and a small stemmer to produce keywords.
func ExtractTopics(text string) []string {
	text = strings.ToLower(text)
	fields := strings.FieldsFunc(text, func(r rune) bool { return !(r >= 'a' && r <= 'z') })
	stop := map[string]bool{"the": true, "and": true, "for": true, "with": true, "that": true, "this": true, "from": true, "what": true, "when": true, "where": true, "how": true, "why": true, "are": true, "is": true, "a": true, "an": true, "of": true, "to": true, "in": true, "on": true, "by": true, "be": true, "as": true, "or": true, "it": true}
	freq := map[string]int{}
	for _, f := range fields {
		if len(f) < 3 {
			continue
		}
		if stop[f] {
			continue
		}
		s := simpleStem(f)
		freq[s]++
	}
	type kv struct {
		k string
		v int
	}
	var arr []kv
	for k, v := range freq {
		arr = append(arr, kv{k, v})
	}
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
	m, err := storage.LoadStudentTopics(studentID)
	if err != nil {
		return err
	}
	for _, t := range topics {
		m[t] = m[t] + 1
	}
	return storage.SaveStudentTopics(studentID, m)
}

// GenerateStudyPlan produces a StudyPlan from a completed session. It uses storage and
// the ai package to enrich focus recommendations when enabled.
func GenerateStudyPlan(s *model.Session) *model.StudyPlan {
	plan := &model.StudyPlan{StudentID: s.StudentID, Subject: s.Subject, CreatedAt: time.Now(), Focus: map[string]string{}}
	var nextTimers []int
	for _, t := range s.Tasks {
		if t.Correct != nil && !*t.Correct {
			key := "general"
			if len(t.Topics) > 0 {
				key = t.Topics[0]
			}
			plan.Focus[key] = "review topic, practice similar problems"
			_ = UpdateStudentTopicFrequencies(s.StudentID, t.Topics, true)
		}
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
		next := int(mathRound(float64(est) * adj))
		nextTimers = append(nextTimers, next)
	}
	if m, err := storage.LoadStudentTopics(s.StudentID); err == nil {
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

	// enrich with AI if available
	for topic := range plan.Focus {
		prompt := fmt.Sprintf("You are an educational assistant. Given a student who struggled or needs practice in the topic '%s', provide a single concise (one-sentence) actionable recommendation and one short practice suggestion.", topic)
		if summary, err := ai.QueryGemini(prompt); err == nil {
			if len(summary) > 300 {
				plan.Focus[topic] = summary[:300]
			} else {
				plan.Focus[topic] = summary
			}
		}
	}

	plan.NextTimers = nextTimers
	return plan
}

func mathRound(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

// badge catalog and helper
var BadgeCatalog = map[string]model.Badge{
	"First Steps":      {ID: "first_steps", Name: "First Steps", Icon: "🏁", Color: "#6ad", Description: "Completed your first session", Points: 10},
	"Accuracy Ace":     {ID: "accuracy_ace", Name: "Accuracy Ace", Icon: "🎯", Color: "#3a8", Description: "80%+ accuracy in a session", Points: 20},
	"Quick Solver":     {ID: "quick_solver", Name: "Quick Solver", Icon: "⚡", Color: "#f6a", Description: "Solved most tasks faster than estimate", Points: 15},
	"3-Day Streak":     {ID: "streak_3", Name: "3-Day Streak", Icon: "🔥", Color: "#f90", Description: "Active 3 days in a row", Points: 10},
	"7-Day Streak":     {ID: "streak_7", Name: "7-Day Streak", Icon: "🏆", Color: "#fc0", Description: "Active 7 days in a row", Points: 50},
	"Speed Demon":      {ID: "speed_demon", Name: "Speed Demon", Icon: "🚀", Color: "#a6f", Description: "All tasks completed faster than estimate", Points: 25},
	"Topic Apprentice": {ID: "topic_apprentice", Name: "Topic Apprentice", Icon: "📘", Color: "#48a", Description: "Practice a topic multiple times", Points: 15},
	"Topic Master":     {ID: "topic_master", Name: "Topic Master", Icon: "📜", Color: "#6c3", Description: "Mastered a topic (many improvements)", Points: 60},
}

// GetBadgeByName returns a badge by its human name.
func GetBadgeByName(name string) (model.Badge, bool) {
	b, ok := BadgeCatalog[name]
	return b, ok
}

// IsLikelyAcademicTask applies simple heuristics to determine whether a free-text
// student task looks like an academic problem (used as a fallback when AI is disabled).
func IsLikelyAcademicTask(text string) bool {
	s := strings.ToLower(strings.TrimSpace(text))
	if len(s) < 5 {
		return false
	}
	// common academic action verbs and phrases
	keywords := []string{"solve", "calculate", "prove", "derive", "explain", "define", "compare", "evaluate", "what is", "find", "show that", "compute", "simplify", "integrate", "differentiate", "diagram", "describe", "why", "how"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	// presence of numbers, math symbols or question mark often indicates an academic prompt
	if strings.ContainsAny(s, "0123456789=+-*/^()") || strings.Contains(s, "?") {
		return true
	}
	// longer descriptive prompts are more likely academic
	if len(s) > 60 {
		return true
	}
	return false
}
