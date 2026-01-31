package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// --- Core Data Models ---

type UserProfile struct {
	Name  string `json:"name"`
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

// --- State Management (Simulated DB) ---

type Session struct {
	ID        string
	Profile   UserProfile
	Questions []QuizQuestion
	History   []QuizHistoryItem
	CreatedAt time.Time
}

var (
	sessionStore = make(map[string]*Session)
	storeMutex   sync.RWMutex
	geminiClient *genai.Client
)

const SystemGuardrail = `You are a strict Educational Guardian for EduPulse AI. 
Your SOLE purpose is to facilitate academic and professional learning. 
1. Only generate content related to school subjects, professional skills, or constructive hobbies.
2. REJECT any topics that are: violent, sexually explicit, involving illegal activities, purely celebrity gossip, hateful, or otherwise inappropriate for a classroom.
3. If a topic is inappropriate or non-educational, set "isValidTopic" to false and provide a polite reason.`

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

	fmt.Println("key", apiKey);
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
		fmt.Printf("Generation error: %v\n", err);
		fmt.Println("Response", resp);

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

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var profile UserProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionID := fmt.Sprintf("sess_%d", time.Now().UnixNano())
	storeMutex.Lock()
	sessionStore[sessionID] = &Session{
		ID:        sessionID,
		Profile:   profile,
		CreatedAt: time.Now(),
	}
	storeMutex.Unlock()
	saveSessions()

	json.NewEncoder(w).Encode(map[string]string{"sessionId": sessionID})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	storeMutex.RLock()
	sess, exists := sessionStore[sessionID]
	storeMutex.RUnlock()

	if !exists {
		log.Printf("Auth failed: Session %s not found (server likely restarted)", sessionID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := generateQuizAI(r.Context(), sess.Profile, req)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			http.Error(w, "AI Rate Limit Exceeded. Please wait a moment.", http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.IsValid {
		storeMutex.Lock()
		sessionStore[sessionID].Questions = resp.Questions
		storeMutex.Unlock()
		saveSessions()
	}

	json.NewEncoder(w).Encode(resp)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	storeMutex.RLock()
	sess, exists := sessionStore[sessionID]
	storeMutex.RUnlock()

	if !exists {
		log.Printf("Auth failed: Session %s not found", sessionID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var result QuizResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	analysis, err := analyzePerformanceAI(r.Context(), sess.Profile, result, sess.Questions)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			http.Error(w, "AI Rate Limit Exceeded. Please wait a moment.", http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to History (Equivalent to db.saveQuizResult)
	storeMutex.Lock()
	if s, ok := sessionStore[sessionID]; ok {
		s.History = append(s.History, QuizHistoryItem{
			Timestamp: time.Now(),
			Result:    result,
			Analysis:  *analysis,
		})
	}
	storeMutex.Unlock()
	saveSessions()

	json.NewEncoder(w).Encode(analysis)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	storeMutex.RLock()
	sess, exists := sessionStore[sessionID]
	storeMutex.RUnlock()

	if !exists {
		log.Printf("Auth failed: Session %s not found", sessionID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(sess.History)
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

func saveSessions() {
	storeMutex.RLock()
	defer storeMutex.RUnlock()

	data, err := json.MarshalIndent(sessionStore, "", "  ")
	if err != nil {
		log.Printf("Error marshalling sessions: %v", err)
		return
	}

	if err := os.WriteFile("sessions.json", data, 0644); err != nil {
		log.Printf("Error saving sessions: %v", err)
	}
}

func loadSessions() {
	data, err := os.ReadFile("sessions.json")
	if err != nil {
		return // File might not exist yet (first run)
	}

	storeMutex.Lock()
	defer storeMutex.Unlock()
	json.Unmarshal(data, &sessionStore)
	log.Printf("Loaded %d sessions from disk", len(sessionStore))
}

func main() {
	loadEnv()
	loadSessions()
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
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/analyze", handleAnalyze)
	http.HandleFunc("/api/history", handleHistory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("EduPulse AI (Go Version) running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
