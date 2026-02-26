package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
)

// --- Core Data Models ---

type UserProfile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Grade string `json:"grade"`
}

type QuizQuestion struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"` // "MCQ" or "OPEN"
	Text             string   `json:"text"`
	Options          []string `json:"options,omitempty"`
	CorrectIndex     int      `json:"correctIndex,omitempty"`
	Explanation      string   `json:"explanation"`
	TimeLimitSeconds int      `json:"timeLimitSeconds"`
}

type QuestionPerformance struct {
	QuestionID       string  `json:"questionId"`
	TimeSpentSeconds float64 `json:"timeSpentSeconds"`
	Attempts         int     `json:"attempts"`
	IsCorrect        bool    `json:"isCorrect"`
	UserResponse     string  `json:"userResponse"`
	Revisits         int     `json:"revisits"`
	AIFeedback       string  `json:"aiFeedback,omitempty"`
}

type QuizResult struct {
	TotalQuestions int                   `json:"totalQuestions"`
	CorrectAnswers int                   `json:"correctAnswers"`
	TotalTimeSpent float64               `json:"totalTimeSpent"`
	Performances   []QuestionPerformance `json:"performances"`
	Mode           string                `json:"mode"`
}

type AIAnalysis struct {
	Summary            string   `json:"summary"`
	Strengths          []string `json:"strengths"`
	Weaknesses         []string `json:"weaknesses"`
	Recommendations    []string `json:"recommendations"`
	GradedPerformances []struct {
		QuestionID string `json:"questionId"`
		IsCorrect  bool   `json:"isCorrect"`
		AIFeedback string `json:"aiFeedback"`
	} `json:"gradedPerformances"`
}

type QuizHistoryItem struct {
	Timestamp time.Time  `json:"timestamp"`
	Result    QuizResult `json:"result"`
	Analysis  AIAnalysis `json:"analysis"`
}

type DashboardData struct {
	Profile         UserProfile       `json:"profile"`
	TotalQuizzes    int               `json:"totalQuizzes"`
	AverageScore    float64           `json:"averageScore"`
	RecentActivity  []QuizHistoryItem `json:"recentActivity"`
	Recommendations []string          `json:"recommendations"`
	Weaknesses      []string          `json:"weaknesses"`
}

// --- Request/Response Structures ---

type GenerateRequest struct {
	Topic string `json:"topic"`
	Mode  string `json:"mode"`
	Type  string `json:"type"`
	Image string `json:"image,omitempty"` // Base64
	Goal  string `json:"goal"`            // "QUIZ" or "PLAN"
}

type StudyPlanModule struct {
	Topic        string `json:"topic"`
	Description  string `json:"description"`
	Activity     string `json:"activity"`
	TimeEstimate string `json:"timeEstimate"`
}

type StudyPlanContent struct {
	Title        string            `json:"title"`
	Introduction string            `json:"introduction"`
	Modules      []StudyPlanModule `json:"modules"`
}

type GenerateResponse struct {
	Questions []QuizQuestion    `json:"questions,omitempty"`
	StudyPlan *StudyPlanContent `json:"studyPlan,omitempty"`
	IsValid   bool              `json:"isValid"`
	Reason    string            `json:"reason,omitempty"`
}

// --- Database & State ---

var (
	db           *sql.DB
	geminiClient *genai.Client
)

const SystemGuardrail = `You are a strict Educational Guardian for EduPulse. 
Your SOLE purpose is to facilitate academic and professional learning. 
1. Only generate content related to school subjects, professional skills, or constructive hobbies.
2. REJECT any topics that are: violent, sexually explicit, involving illegal activities, purely celebrity gossip, hateful, or otherwise inappropriate for a classroom.
3. If a topic is inappropriate or non-educational, set "isValidTopic" to false and provide a polite reason.`

// --- Database Initialization ---

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./edupulse.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create Tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			grade TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS quiz_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			score_percent REAL,
			full_json TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Fatalf("Error creating table: %v\nQuery: %s", err, q)
		}
	}
	log.Println("Database initialized successfully.")

	// Migrations for new columns (Ignore errors if they already exist)
	migrationQueries := []string{
		`ALTER TABLE users ADD COLUMN recommendations TEXT;`,
		`ALTER TABLE users ADD COLUMN last_quiz_count INTEGER DEFAULT 0;`,
	}
	for _, q := range migrationQueries {
		db.Exec(q)
	}
}

// --- Gemini Service ---

func initGemini() error {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	fmt.Println("key", apiKey)
	var err error
	geminiClient, err = genai.NewClient(ctx, option.WithAPIKey(apiKey))
	return err
}

func generateContentWithRetry(ctx context.Context, model *genai.GenerativeModel, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	var err error
	delay := 4 * time.Second
	maxRetries := 3

	for i := 0; i <= maxRetries; i++ {
		resp, err := model.GenerateContent(ctx, parts...)
		if err == nil {
			return resp, nil
		}
		fmt.Printf("Generation error: %v\n", err)
		fmt.Println("Response", resp)

		if i < maxRetries && (strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "503")) {
			log.Printf("Gemini API rate limit/busy (attempt %d/%d). Retrying in %v...", i+1, maxRetries, delay)
			time.Sleep(delay)
			delay *= 2
			continue
		}
		return nil, err
	}
	return nil, err
}

func generateQuizAI(ctx context.Context, profile UserProfile, req GenerateRequest) (*GenerateResponse, error) {
	model := geminiClient.GenerativeModel("gemini-3-flash-preview")
	model.ResponseMIMEType = "application/json"
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(SystemGuardrail)},
	}

	var prompt string

	if req.Goal == "PLAN" {
		model.ResponseSchema = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"isValidTopic":    {Type: genai.TypeBoolean},
				"rejectionReason": {Type: genai.TypeString},
				"studyPlan": {
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"title":        {Type: genai.TypeString},
						"introduction": {Type: genai.TypeString},
						"modules": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"topic":        {Type: genai.TypeString},
									"description":  {Type: genai.TypeString},
									"activity":     {Type: genai.TypeString},
									"timeEstimate": {Type: genai.TypeString},
								},
								Required: []string{"topic", "description", "activity", "timeEstimate"},
							},
						},
					},
					Required: []string{"title", "introduction", "modules"},
				},
			},
			Required: []string{"isValidTopic"},
		}
		prompt = fmt.Sprintf(`Generate a structured study plan for:
		Topic: %s
		Target Audience: %s
		
		Follow System Instructions. If valid, provide a step-by-step study plan.`, req.Topic, profile.Grade)
	} else {
		// Default to Quiz
		model.ResponseSchema = &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"isValidTopic":    {Type: genai.TypeBoolean},
				"rejectionReason": {Type: genai.TypeString},
				"questions": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"id":               {Type: genai.TypeString},
							"type":             {Type: genai.TypeString, Enum: []string{"MCQ", "OPEN"}},
							"text":             {Type: genai.TypeString},
							"options":          {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
							"correctIndex":     {Type: genai.TypeInteger},
							"explanation":      {Type: genai.TypeString},
							"timeLimitSeconds": {Type: genai.TypeInteger},
						},
						Required: []string{"id", "type", "text", "explanation", "timeLimitSeconds"},
					},
				},
			},
			Required: []string{"isValidTopic", "questions"},
		}

		prompt = fmt.Sprintf(`Evaluate and generate a 5-question quiz for:
		Prompt/Topic: %s
		Difficulty: %s
		Question Format Preference: %s
		
		Follow the System Instructions for content safety.
		If valid, generate questions based on the format preference.`+
			`\n\nStrict Requirements:
		- If Format is MCQ: All questions must be MCQ with 4 options.
		- If Format is OPEN: All questions must be open-ended requiring a written response.
		- If Format is MIXED: Provide a blend of MCQ and open-ended.
		- Ensure all questions are academically rigorous.
		- Provide a clear 'explanation' for the ideal answer for every question.`,
			req.Topic, profile.Grade, req.Type)
	}

	parts := []genai.Part{genai.Text(prompt)}

	if req.Image != "" {
		// Expecting "data:image/png;base64,..."
		if idx := strings.Index(req.Image, ";base64,"); idx > 0 {
			mimeType := req.Image[5:idx] // e.g. "image/png"
			subtype := "jpeg"
			if split := strings.Split(mimeType, "/"); len(split) == 2 {
				subtype = split[1]
			}

			if decoded, err := base64.StdEncoding.DecodeString(req.Image[idx+8:]); err == nil {
				parts = append(parts, genai.ImageData(subtype, decoded))
			}
		}
	}

	resp, err := generateContentWithRetry(ctx, model, parts...)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("model not found (check region/access): %w", err)
		}
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no content generated")
	}

	var rawResp struct {
		IsValidTopic    bool              `json:"isValidTopic"`
		RejectionReason string            `json:"rejectionReason"`
		Questions       []QuizQuestion    `json:"questions"`
		StudyPlan       *StudyPlanContent `json:"studyPlan"`
	}

	// Extract JSON from the first part
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			if err := json.Unmarshal([]byte(txt), &rawResp); err != nil {
				return nil, err
			}
			break
		}
	}

	return &GenerateResponse{
		Questions: rawResp.Questions,
		StudyPlan: rawResp.StudyPlan,
		IsValid:   rawResp.IsValidTopic,
		Reason:    rawResp.RejectionReason,
	}, nil
}

func analyzePerformanceAI(ctx context.Context, profile UserProfile, result QuizResult, questions []QuizQuestion) (*AIAnalysis, error) {
	model := geminiClient.GenerativeModel("gemini-3-flash-preview")
	model.ResponseMIMEType = "application/json"
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(SystemGuardrail)},
	}
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary":         {Type: genai.TypeString},
			"strengths":       {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
			"weaknesses":      {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
			"recommendations": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
			"gradedPerformances": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"questionId": {Type: genai.TypeString},
						"isCorrect":  {Type: genai.TypeBoolean},
						"aiFeedback": {Type: genai.TypeString},
					},
					Required: []string{"questionId", "isCorrect", "aiFeedback"},
				},
			},
		},
		Required: []string{"summary", "strengths", "weaknesses", "recommendations", "gradedPerformances"},
	}

	// Prepare data for AI
	type AnalysisInput struct {
		QuestionID    string `json:"questionId"`
		QuestionText  string `json:"questionText"`
		UserResponse  string `json:"userResponse"`
		CorrectAnswer string `json:"correctAnswer"`
	}

	var inputs []AnalysisInput
	for i, q := range questions {
		perf := result.Performances[i]
		correct := q.Explanation
		if q.Type == "MCQ" && len(q.Options) > q.CorrectIndex {
			correct = q.Options[q.CorrectIndex]
		}
		inputs = append(inputs, AnalysisInput{
			QuestionID:    q.ID,
			QuestionText:  q.Text,
			UserResponse:  perf.UserResponse,
			CorrectAnswer: correct,
		})
	}

	inputJSON, _ := json.Marshal(inputs)
	prompt := fmt.Sprintf(`Evaluate these quiz results for a %s student.
    DATA: %s
    
    Analyze user responses against correct answers.
    For every OPEN type question, analyze the 'userResponse' against the 'questionText' and 'correctAnswer'.
    Decide if it is 'isCorrect' (Boolean) and provide 'aiFeedback'.
    Even for MCQ, provide brief feedback if they were wrong.`, profile.Grade, string(inputJSON))

	resp, err := generateContentWithRetry(ctx, model, genai.Text(prompt))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("model not found (check region/access): %w", err)
		}
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	var analysis AIAnalysis
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			if err := json.Unmarshal([]byte(txt), &analysis); err != nil {
				return nil, err
			}
			break
		}
	}
	return &analysis, nil
}

// --- Handlers ---

func getUserFromSession(r *http.Request) (*UserProfile, error) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		return nil, fmt.Errorf("no session ID")
	}

	var user UserProfile
	err := db.QueryRow(`
		SELECT u.id, u.name, u.email, u.grade 
		FROM users u 
		JOIN sessions s ON u.id = s.user_id 
		WHERE s.session_id = ?`, sessionID).Scan(&user.ID, &user.Name, &user.Email, &user.Grade)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Grade    string `json:"grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	res, err := db.Exec("INSERT INTO users (name, email, password_hash, grade) VALUES (?, ?, ?, ?)",
		req.Name, req.Email, string(hashedPassword), req.Grade)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := res.LastInsertId()
	sessionID := fmt.Sprintf("sess_%d_%d", userID, time.Now().UnixNano())
	db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, userID)

	json.NewEncoder(w).Encode(map[string]string{"sessionId": sessionID})
}

func handleGuest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Grade string `json:"grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a temporary user with a unique identifier
	tempID := time.Now().UnixNano()
	email := fmt.Sprintf("guest_%d@edupulse.local", tempID)
	password := fmt.Sprintf("guest_pass_%d", tempID)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	res, err := db.Exec("INSERT INTO users (name, email, password_hash, grade) VALUES (?, ?, ?, ?)", req.Name, email, string(hashedPassword), req.Grade)
	if err != nil {
		http.Error(w, "Guest creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, _ := res.LastInsertId()
	sessionID := fmt.Sprintf("sess_%d_%d", userID, tempID)
	db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, userID)

	json.NewEncoder(w).Encode(map[string]string{"sessionId": sessionID})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id int
	var hash string
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", req.Email).Scan(&id, &hash)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionID := fmt.Sprintf("sess_%d_%d", id, time.Now().UnixNano())
	db.Exec("INSERT INTO sessions (session_id, user_id) VALUES (?, ?)", sessionID, id)

	json.NewEncoder(w).Encode(map[string]string{"sessionId": sessionID})
}

func handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name            string `json:"name"`
		Grade           string `json:"grade"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update Basic Info
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Grade != "" {
		user.Grade = req.Grade
	}

	_, err = db.Exec("UPDATE users SET name = ?, grade = ? WHERE id = ?", user.Name, user.Grade, user.ID)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Handle Password Change
	if req.NewPassword != "" {
		if strings.HasPrefix(user.Email, "guest_") {
			http.Error(w, "Guest accounts cannot change passwords.", http.StatusForbidden)
			return
		}

		var hash string
		if err := db.QueryRow("SELECT password_hash FROM users WHERE id = ?", user.ID).Scan(&hash); err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
			http.Error(w, "Invalid current password", http.StatusUnauthorized)
			return
		}

		newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(newHash), user.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := generateQuizAI(r.Context(), *user, req)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			http.Error(w, "AI Rate Limit Exceeded. Please wait a moment.", http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var result QuizResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Reconstruct questions context from result for analysis
	var questions []QuizQuestion
	for _, p := range result.Performances {
		questions = append(questions, QuizQuestion{
			ID:   p.QuestionID,
			Type: "UNKNOWN",
			Text: "Referenced in performance",
		})
	}

	analysis, err := analyzePerformanceAI(r.Context(), *user, result, questions)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			http.Error(w, "AI Rate Limit Exceeded. Please wait a moment.", http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to DB
	fullJson, _ := json.Marshal(QuizHistoryItem{
		Timestamp: time.Now(),
		Result:    result,
		Analysis:  *analysis,
	})

	score := 0.0
	if result.TotalQuestions > 0 {
		correct := 0
		for _, gp := range analysis.GradedPerformances {
			if gp.IsCorrect {
				correct++
			}
		}
		score = (float64(correct) / float64(result.TotalQuestions)) * 100
	}

	_, dbErr := db.Exec("INSERT INTO quiz_history (user_id, score_percent, full_json) VALUES (?, ?, ?)",
		user.ID, score, string(fullJson))
	if dbErr != nil {
		log.Printf("Error saving history: %v", dbErr)
	}

	json.NewEncoder(w).Encode(analysis)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch Stats
	var totalQuizzes int
	var avgScore sql.NullFloat64
	db.QueryRow("SELECT COUNT(*), AVG(score_percent) FROM quiz_history WHERE user_id = ?", user.ID).Scan(&totalQuizzes, &avgScore)

	// Fetch Recent History
	rows, err := db.Query("SELECT full_json FROM quiz_history WHERE user_id = ? ORDER BY created_at DESC LIMIT 5", user.ID)
	history := make([]QuizHistoryItem, 0)
	allWeaknesses := make([]string, 0)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var jsonStr string
			rows.Scan(&jsonStr)
			var item QuizHistoryItem
			if json.Unmarshal([]byte(jsonStr), &item) == nil {
				history = append(history, item)
				allWeaknesses = append(allWeaknesses, item.Analysis.Weaknesses...)
			}
		}
	}

	// Check cache for recommendations
	var storedRecs sql.NullString
	var lastCount int
	db.QueryRow("SELECT recommendations, last_quiz_count FROM users WHERE id = ?", user.ID).Scan(&storedRecs, &lastCount)

	var recommendations []string

	// Only generate if new quizzes taken OR no recommendations exist
	if totalQuizzes > lastCount || !storedRecs.Valid || storedRecs.String == "" {
		if len(allWeaknesses) > 0 {
			// Simple deduplication
			weaknessMap := make(map[string]bool)
			var uniqueWeaknesses []string
			for _, w := range allWeaknesses {
				if !weaknessMap[w] {
					weaknessMap[w] = true
					uniqueWeaknesses = append(uniqueWeaknesses, w)
				}
			}
			if len(uniqueWeaknesses) > 5 {
				uniqueWeaknesses = uniqueWeaknesses[:5]
			}

			model := geminiClient.GenerativeModel("gemini-3-flash-preview")
			prompt := fmt.Sprintf("Based on these academic weaknesses for a %s student: %v. Suggest 3 specific, short study topics or actions.", user.Grade, uniqueWeaknesses)
			resp, err := model.GenerateContent(r.Context(), genai.Text(prompt))
			if err == nil && len(resp.Candidates) > 0 {
				if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					lines := strings.Split(string(txt), "\n")
					for _, l := range lines {
						clean := strings.TrimSpace(strings.TrimLeft(l, "*-•1234567890. "))
						if clean != "" {
							recommendations = append(recommendations, clean)
						}
					}
				}
			}
			// Save to DB
			if len(recommendations) > 0 {
				recJSON, _ := json.Marshal(recommendations)
				db.Exec("UPDATE users SET recommendations = ?, last_quiz_count = ? WHERE id = ?", string(recJSON), totalQuizzes, user.ID)
			}
		}
	} else {
		json.Unmarshal([]byte(storedRecs.String), &recommendations)
	}

	if len(recommendations) == 0 {
		recommendations = []string{"Complete more quizzes to get personalized insights.", "Review your recent quiz feedback."}
	}

	data := DashboardData{
		Profile:         *user,
		TotalQuizzes:    totalQuizzes,
		AverageScore:    avgScore.Float64,
		RecentActivity:  history,
		Recommendations: recommendations,
		Weaknesses:      allWeaknesses,
	}

	json.NewEncoder(w).Encode(data)
}

func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
}

func main() {
	loadEnv()
	initDB()
	if err := initGemini(); err != nil {
		log.Printf("Warning: Gemini client failed to init (check API_KEY): %v", err)
	}

	// Serve Index
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Serve Static Assets
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API Routes
	http.HandleFunc("/api/signup", handleSignup)
	http.HandleFunc("/api/guest", handleGuest)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/profile", handleUpdateProfile)
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/analyze", handleAnalyze)
	http.HandleFunc("/api/dashboard", handleDashboard)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("EduPulse (Go Version) running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
