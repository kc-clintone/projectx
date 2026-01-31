package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"

	"github.com/kc-clintone/study-coach/ai"
	"github.com/kc-clintone/study-coach/coach"
	"github.com/kc-clintone/study-coach/model"
	"github.com/kc-clintone/study-coach/storage"
	"github.com/kc-clintone/study-coach/validate"
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

	// require student level (grade) to personalize timers
	if strings.TrimSpace(req.StudentLevel) == "" {
		http.Error(w, "student_level required", http.StatusBadRequest)
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
			if validate.IsLikelyAcademicTask(t.Prompt) {
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

	// Try to generate a study plan via AI; fall back to heuristic generator
	var plan *model.StudyPlan
	aiPrompt := buildPlanPrompt(sess)
	aiResp, _ := ai.QueryGemini(aiPrompt)
	if strings.TrimSpace(aiResp) != "" {
		// try to extract a JSON object from the LLM output (supports ```json blocks or plain JSON)
		if jsonStr, ok := extractJSONFromString(aiResp); ok {
			var aiPlan model.StudyPlan
			if json.Unmarshal([]byte(jsonStr), &aiPlan) == nil {
				if aiPlan.StudentID == "" {
					aiPlan.StudentID = sess.StudentID
				}
				if aiPlan.Subject == "" {
					aiPlan.Subject = sess.Subject
				}
				if aiPlan.CreatedAt.IsZero() {
					aiPlan.CreatedAt = time.Now()
				}
				// ensure focus and timers populated
				if len(aiPlan.Focus) == 0 {
					aiPlan.Focus = parseFocusFromText(aiResp)
				}
				if len(aiPlan.NextTimers) == 0 {
					for _, t := range sess.Tasks {
						aiPlan.NextTimers = append(aiPlan.NextTimers, t.EstimatedSecs)
					}
				}
				plan = &aiPlan
			}
		}
	}

	// If AI didn't produce a structured plan, try to salvage simple focus lines from AI text
	if plan == nil && strings.TrimSpace(aiResp) != "" {
		if f := parseFocusFromText(aiResp); len(f) > 0 {
			p := &model.StudyPlan{StudentID: sess.StudentID, Subject: sess.Subject, CreatedAt: time.Now(), Focus: f}
			for _, t := range sess.Tasks {
				p.NextTimers = append(p.NextTimers, t.EstimatedSecs)
			}
			plan = p
		}
	}

	if plan == nil {
		plan = coach.GenerateStudyPlan(sess)
	}

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

	// capture existing achievements so we can return newly-earned ones
	existing := make(map[string]bool)
	for _, a := range profile.Achievements {
		existing[a] = true
	}
	var newAchievements []model.AchievementRecord

	// compute streak
	if profile.LastActive.IsZero() {
		// first activity
		if rec := AddAchievement(profile, "First Steps"); rec != nil {
			newAchievements = append(newAchievements, *rec)
		}
		profile.Streak = 1
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
		if t.ActualSeconds > 0 && t.EstimatedSecs > 0 && t.ActualSeconds < t.EstimatedSecs {
			fastCount++
		}
	}
	if total > 0 {
		if correctCount*100/total >= 80 {
			if rec := AddAchievement(profile, "Accuracy Ace"); rec != nil {
				newAchievements = append(newAchievements, *rec)
			}
		}
		if fastCount*2 > total { // more than half
			if rec := AddAchievement(profile, "Quick Solver"); rec != nil {
				newAchievements = append(newAchievements, *rec)
			}
		}
	}

	// streak achievements
	if profile.Streak >= 3 {
		if rec := AddAchievement(profile, "3-Day Streak"); rec != nil {
			newAchievements = append(newAchievements, *rec)
		}
	}
	if profile.Streak >= 7 {
		if rec := AddAchievement(profile, "7-Day Streak"); rec != nil {
			newAchievements = append(newAchievements, *rec)
		}
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
	// return plan plus newly-earned achievements so UI can display them immediately
	resp := map[string]interface{}{"plan": plan, "new_achievements": newAchievements}
	json.NewEncoder(w).Encode(resp)
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
func AddAchievement(p *model.ProfileResponse, name string) *model.AchievementRecord {
	for _, a := range p.Achievements {
		if a == name {
			return nil
		}
	}
	p.Achievements = append(p.Achievements, name)
	if b, ok := coach.GetBadgeByName(name); ok {
		// add to earned badges and points
		p.EarnedBadges = append(p.EarnedBadges, b)
		p.Points += b.Points
	}
	rec := model.AchievementRecord{Name: name, EarnedAt: time.Now()}
	p.RecentAchievements = append(p.RecentAchievements, rec)
	return &rec
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

// new helper to ask AI to produce a JSON study plan
func buildPlanPrompt(sess *model.Session) string {
	var b strings.Builder
	b.WriteString("You are an educational assistant. Based on the student's recent session data, produce a JSON object matching the StudyPlan schema exactly. Respond ONLY with the JSON.\n\n")
	b.WriteString("Schema example:\n")
	b.WriteString(`{"student_id":"<id>","subject":"<subject>","created_at":"2024-01-02T15:04:05Z","focus":{"algebra":"short recommendation"},"next_timers":[60,120]}` + "\n\n")
	b.WriteString("If you cannot produce well-formed JSON, wrap the JSON in triple backticks with optional language hint (```json ... ```). Use topic keys as short strings and recommendations as concise single sentences.\n\n")
	b.WriteString("Session data:\n")
	for i, t := range sess.Tasks {
		b.WriteString(fmt.Sprintf("%d. prompt: %s | est: %d | actual: %d | correct: %v\n", i+1, t.Prompt, t.EstimatedSecs, t.ActualSeconds, t.Correct != nil && *t.Correct))
	}
	b.WriteString(fmt.Sprintf("\nSubject: %s\nStudentID: %s\n", sess.Subject, sess.StudentID))
	b.WriteString("Return only the JSON object. Do not include any explanatory text.\n")
	return b.String()
}

// extractJSONFromString attempts to locate a JSON object inside a free-form string.
func extractJSONFromString(s string) (string, bool) {
	// first look for ```json ... ``` blocks
	re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
	if m := re.FindStringSubmatch(s); len(m) == 2 {
		return m[1], true
	}
	// then look for ``` ... ``` blocks
	re2 := regexp.MustCompile("(?s)```\\s*(\\{.*?\\})\\s*```")
	if m := re2.FindStringSubmatch(s); len(m) == 2 {
		return m[1], true
	}
	// otherwise find the first balanced JSON object by scanning braces (ignoring quoted braces)
	start := strings.Index(s, "{")
	if start == -1 {
		return "", false
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' && !escape {
			escape = true
			continue
		}
		if ch == '"' && !escape {
			inString = !inString
		}
		escape = false
		if inString {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

// parseFocusFromText looks for simple lines like "topic: recommendation" and returns a map.
func parseFocusFromText(s string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(s, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		// remove leading list markers
		if strings.HasPrefix(ln, "- ") || strings.HasPrefix(ln, "* ") {
			ln = strings.TrimSpace(ln[2:])
		}
		// split at first ':' or ' - '
		if idx := strings.Index(ln, ":"); idx != -1 {
			key := strings.TrimSpace(ln[:idx])
			val := strings.TrimSpace(ln[idx+1:])
			if key != "" && val != "" {
				out[key] = val
				continue
			}
		}
		if idx := strings.Index(ln, " - "); idx != -1 {
			key := strings.TrimSpace(ln[:idx])
			val := strings.TrimSpace(ln[idx+3:])
			if key != "" && val != "" {
				out[key] = val
			}
		}
	}
	return out
}
