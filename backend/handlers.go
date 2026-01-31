package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"

	"github.com/kc-clintone/study-coach/ai"
	"github.com/kc-clintone/study-coach/coach"
	"github.com/kc-clintone/study-coach/model"
	"github.com/kc-clintone/study-coach/storage"
)

// registerHandler creates a new local user with bcrypt hashed password
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if _, err := storage.LoadUser(req.Username); err == nil {
		http.Error(w, "user exists", http.StatusConflict)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &model.User{Username: req.Username, PasswordHash: string(hash), Email: req.Email, CreatedAt: time.Now()}
	if err := storage.SaveUser(u); err != nil {
		http.Error(w, "failed to save user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("created"))
}

// loginHandler authenticates a user and sets a session cookie
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	u, err := storage.LoadUser(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	// set a very simple cookie-based session (no signing) for hackathon; path set for future extension
	cookie := &http.Cookie{Name: "sc_session", Value: req.Username, Path: "/", HttpOnly: true, Expires: time.Now().Add(7 * 24 * time.Hour)}
	http.SetCookie(w, cookie)
	w.Write([]byte("ok"))
}

// logoutHandler clears the session cookie
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{Name: "sc_session", Value: "", Path: "/", HttpOnly: true, Expires: time.Unix(0, 0)}
	http.SetCookie(w, cookie)
	w.Write([]byte("ok"))
}

// getMeHandler returns basic info about current user from cookie
func getMeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sc_session")
	if err != nil || c.Value == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	u, err := storage.LoadUser(c.Value)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	// build a simple response with recent plan & achievements
	profile, _ := storage.LoadProfile(u.Username)
	resp := map[string]interface{}{
		"username":      u.Username,
		"email":         u.Email,
		"student_level": u.StudentLevel,
		"profile":       profile,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// createSessionHandler accepts student info and a list of tasks and returns a session with timers
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// basic validation: only accept educational subjects
	if !coach.IsEducationalSubject(req.Subject) {
		http.Error(w, "subject not educational", http.StatusBadRequest)
		return
	}

	// Validate each task to ensure it's academically aligned using AI when available,
	// with a heuristic fallback. If any task is non-academic, reject the request.
	for _, t := range req.Tasks {
		prompt := fmt.Sprintf("Classify the following student task as 'academic' or 'non-academic'. Reply with a single word.\n\nTask: %s", t.Prompt)
		isAcademic := false
		if resp, err := ai.QueryGemini(prompt); err == nil {
			if strings.Contains(strings.ToLower(resp), "academic") {
				isAcademic = true
			}
		}
		// fallback to heuristic
		if !isAcademic {
			if coach.IsLikelyAcademicTask(t.Prompt) {
				isAcademic = true
			}
		}
		if !isAcademic {
			http.Error(w, "task not academic: "+t.Prompt, http.StatusBadRequest)
			return
		}
	}

	sess := coach.NewSession(req.StudentID, req.StudentLevel, req.Subject, req.Tasks)

	if err := storage.SaveSession(sess); err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// submitResultsHandler accepts completed work and timing info and produces an updated study plan
func submitResultsHandler(w http.ResponseWriter, r *http.Request) {
	var res model.Submission
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	sess, err := storage.LoadSession(res.SessionID)
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

	plan := coach.GenerateStudyPlan(sess)

	if err := storage.SaveStudyPlan(sess.StudentID, plan); err != nil {
		http.Error(w, "failed to save plan", http.StatusInternalServerError)
		return
	}

	// --- gamification: update profile, streaks, achievements, snapshots ---
	now := time.Now()
	var profile *model.ProfileResponse
	p, err := storage.LoadProfile(sess.StudentID)
	if err != nil {
		// create default profile
		profile = &model.ProfileResponse{StudentID: sess.StudentID, Topics: map[string]int{}, Achievements: []string{}, Progress: map[string]int{}, WeeklyRecap: []model.TopicImprovement{}, Streak: 0}
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

	// Persist profile and snapshot of topics
	if err := storage.SaveProfile(profile.StudentID, profile); err == nil {
		// build snapshot from current topics
		if topics, err2 := storage.LoadStudentTopics(profile.StudentID); err2 == nil {
			snap := &model.TopicSnapshot{Timestamp: now, Topics: topics}
			_ = storage.SaveSnapshot(profile.StudentID, snap)
		}
	}

	// attach last plan to profile for quick dashboard access
	profile.LastPlan = plan
	_ = storage.SaveProfile(profile.StudentID, profile)

	// --- attach AI-generated summary to profile (if enabled) ---
	// Build a concise prompt describing recent session and plan to generate a short summary
	prompt := buildSummaryPrompt(sess, plan)
	if summary, err := ai.QueryGemini(prompt); err == nil {
		profile.Summary = summary
		// persist updated profile with summary
		_ = storage.SaveProfile(profile.StudentID, profile)
	}

	// --- end gamification update ---

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// getStudyPlanHandler retrieves the study plan for a student
func getStudyPlanHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	plan, err := storage.LoadStudyPlan(id)
	if err != nil {
		http.Error(w, "plan not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// getStudentTopicsHandler retrieves the topics for a student
func getStudentTopicsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	m, err := storage.LoadStudentTopics(id)
	if err != nil {
		http.Error(w, "failed to load topics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// getProfileHandler retrieves the profile for a student
func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["student_id"]
	// load topics and recent snapshot
	topics, _ := storage.LoadStudentTopics(id)
	snapshot, _ := storage.LoadSnapshot(id)
	// build a simple profile response
	profile := model.ProfileResponse{
		StudentID:    id,
		Topics:       topics,
		Achievements: []string{},
		Progress:     map[string]int{},
		WeeklyRecap:  []model.TopicImprovement{},
	}
	if snapshot != nil {
		// compute weekly recap naive: compare snapshot topics to current (if any)
		for k, v := range topics {
			prev := 0
			if snapshot.Topics != nil {
				prev = snapshot.Topics[k]
			}
			if v > prev {
				profile.WeeklyRecap = append(profile.WeeklyRecap, model.TopicImprovement{Topic: k, ImprovedBy: v - prev})
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// helper to append unique achievement and award badge/points
func AddAchievement(p *model.ProfileResponse, name string) {
	for _, a := range p.Achievements {
		if a == name {
			return
		}
	}
	p.Achievements = append(p.Achievements, name)
	if b, ok := coach.GetBadgeByName(name); ok {
		// add to earned badges and points
		p.EarnedBadges = append(p.EarnedBadges, b)
		p.Points += b.Points
	}
}

// buildSummaryPrompt composes a short prompt describing the session and study plan for the LLM
func buildSummaryPrompt(sess *model.Session, plan *model.StudyPlan) string {
	var b strings.Builder
	b.WriteString("You are an educational assistant. Provide a concise summary (2-3 sentences) of the student's recent session, strengths, weaknesses, and 2 quick recommendations.\n\n")
	b.WriteString("Session tasks:\n")
	for i, t := range sess.Tasks {
		b.WriteString(fmt.Sprintf("%d. %s (est %ds, actual %ds, correct: %v)\n", i+1, t.Prompt, t.EstimatedSecs, t.ActualSeconds, t.Correct != nil && *t.Correct))
	}
	b.WriteString("\nStudy plan focus areas:\n")
	for k, v := range plan.Focus {
		b.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	return b.String()
}
