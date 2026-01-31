package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kc-clintone/study-coach/storage"
)

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func TestCreateSessionTaskValidation(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	// task that is clearly non-academic
	reqBody := CreateSessionRequest{StudentID: "s1", StudentLevel: "grade9", Subject: "math", Tasks: []Task{{ID: "t1", Prompt: "buy milk and eggs"}}}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400 for non-academic task, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitResultsAwardsAchievements(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	// create a session with two simple academic tasks
	reqBody := CreateSessionRequest{StudentID: "stu_ach", StudentLevel: "grade10", Subject: "math", Tasks: []Task{{ID: "t1", Prompt: "1+1"}, {ID: "t2", Prompt: "integrate x"}}}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("create session failed: %d %s", rec.Code, rec.Body.String())
	}
	var sess Session
	json.Unmarshal(rec.Body.Bytes(), &sess)

	// submit results: both correct to trigger Accuracy Ace and First Steps
	sub := Submission{SessionID: sess.ID, Results: []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	}{
		{ActualSeconds: 30, Correct: true},
		{ActualSeconds: 90, Correct: true},
	}}
	sb, _ := json.Marshal(sub)
	req2 := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(sb))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("submit failed: %d %s", rec2.Code, rec2.Body.String())
	}

	// load persisted profile and check achievements
	p, err := storage.LoadProfile("stu_ach")
	if err != nil {
		t.Fatalf("failed to load profile: %v", err)
	}
	if !contains(p.Achievements, "First Steps") {
		t.Fatalf("expected First Steps achievement, got %v", p.Achievements)
	}
	if !contains(p.Achievements, "Accuracy Ace") {
		t.Fatalf("expected Accuracy Ace achievement, got %v", p.Achievements)
	}
}
