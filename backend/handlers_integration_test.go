package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kc-clintone/study-coach/storage"
)

func TestFullCreateSubmitAndPlanFlow(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	// create session with grade provided
	reqBody := CreateSessionRequest{StudentID: "fullflow", StudentLevel: "grade10", Subject: "math", Tasks: []Task{{ID: "t1", Prompt: "1+1"}, {ID: "t2", Prompt: "integral of x"}}}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("create session failed: %d %s", rec.Code, rec.Body.String())
	}
	var sess Session
	json.Unmarshal(rec.Body.Bytes(), &sess)

	// submit results
	sub := Submission{SessionID: sess.ID, Results: []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	}{
		{ActualSeconds: 20, Correct: true},
		{ActualSeconds: 120, Correct: false},
	}}
	sb, _ := json.Marshal(sub)
	req2 := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(sb))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("submit failed: %d %s", rec2.Code, rec2.Body.String())
	}

	// fetch study plan
	req3 := httptest.NewRequest("GET", "/api/v1/study_plan/fullflow", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != 200 {
		t.Fatalf("get plan failed: %d %s", rec3.Code, rec3.Body.String())
	}
}
