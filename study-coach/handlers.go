package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// createSessionHandler accepts student info and a list of tasks and returns a session with timers
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// basic validation: only accept educational subjects
	if !IsEducationalSubject(req.Subject) {
		http.Error(w, "subject not educational", http.StatusBadRequest)
		return
	}

	sess := NewSession(req.StudentID, req.StudentLevel, req.Subject, req.Tasks)

	if err := SaveSession(sess); err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// submitResultsHandler accepts completed work and timing info and produces an updated study plan
func submitResultsHandler(w http.ResponseWriter, r *http.Request) {
	var res Submission
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	sess, err := LoadSession(res.SessionID)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	// update session with actual times and correctness
	for i := range res.Results {
		if i < len(sess.Tasks) {
			sess.Tasks[i].ActualSeconds = res.Results[i].ActualSeconds
			// copy correct boolean into pointer
			c := res.Results[i].Correct
			sess.Tasks[i].Correct = &c
			sess.Tasks[i].SubmittedAt = time.Now()
		}
	}

	plan := GenerateStudyPlan(sess)

	if err := SaveStudyPlan(sess.StudentID, plan); err != nil {
		http.Error(w, "failed to save plan", http.StatusInternalServerError)
		return
	}

	// --- gamification: update profile, streaks, achievements, snapshots ---
	now := time.Now()
	var profile *ProfileResponse
	p, err := LoadProfile(sess.StudentID)
	if err != nil {
		// create default profile
		profile = &ProfileResponse{StudentID: sess.StudentID, Topics: map[string]int{}, Achievements: []string{}, Progress: map[string]int{}, WeeklyRecap: []TopicImprovement{}, Streak: 0}
	} else {
		profile = p
	}

	// compute streak
	if profile.LastActive.IsZero() {
		// first activity
		profile.Streak = 1
		AddAchievement(profile, "First Steps")
	} else {
		// compare date differences
		days := int(now.Sub(profile.LastActive).Hours() / 24)
		if days == 0 {
			// same day, do not change streak
		} else if days == 1 {
			profile.Streak = profile.Streak + 1
		} else {
			profile.Streak = 1
		}
	}
	profile.LastActive = now

	// session-based achievements
	correctCount := 0
	fastCount := 0
	total := len(sess.Tasks)
	for _, t := range sess.Tasks {
		if t.Correct != nil && *t.Correct {
			correctCount++
		}
		if t.ActualSeconds > 0 && t.ActualSeconds < t.EstimatedSecs {
			fastCount++
		}
	}
	if total > 0 {
		if correctCount*100/total >= 80 {
			AddAchievement(profile, "Accuracy Ace")
		}
		if fastCount*2 > total { // more than half
			AddAchievement(profile, "Quick Solver")
		}
	}

	// streak achievements
	if profile.Streak >= 3 {
		AddAchievement(profile, "3-Day Streak")
	}
	if profile.Streak >= 7 {
		AddAchievement(profile, "7-Day Streak")
	}

	// persist profile and snapshot of topics
	if err := SaveProfile(profile.StudentID, profile); err == nil {
		// build snapshot from current topics
		if topics, err2 := LoadStudentTopics(profile.StudentID); err2 == nil {
			snap := &TopicSnapshot{Timestamp: now, Topics: topics}
			_ = SaveSnapshot(profile.StudentID, snap)
		}
	}

	// --- end gamification update ---

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func getStudyPlanHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	plan, err := LoadStudyPlan(id)
	if err != nil {
		http.Error(w, "plan not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func getStudentTopicsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	m, err := LoadStudentTopics(id)
	if err != nil {
		http.Error(w, "failed to load topics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	// load topics and recent snapshot
	topics, _ := LoadStudentTopics(id)
	snapshot, _ := LoadSnapshot(id)
	// build a simple profile response
	profile := ProfileResponse{
		StudentID:    id,
		Topics:       topics,
		Achievements: []string{},
		Progress:     map[string]int{},
		WeeklyRecap:  []TopicImprovement{},
	}
	if snapshot != nil {
		// compute weekly recap naive: compare snapshot topics to current (if any)
		for k, v := range topics {
			prev := 0
			if snapshot.Topics != nil {
				prev = snapshot.Topics[k]
			}
			if v > prev {
				profile.WeeklyRecap = append(profile.WeeklyRecap, TopicImprovement{Topic: k, ImprovedBy: v - prev})
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// helper to append unique achievement and award badge/points
func AddAchievement(p *ProfileResponse, name string) {
	for _, a := range p.Achievements {
		if a == name {
			return
		}
	}
	p.Achievements = append(p.Achievements, name)
	if b, ok := GetBadgeByName(name); ok {
		// add to earned badges and points
		p.EarnedBadges = append(p.EarnedBadges, b)
		p.Points += b.Points
	}
}
