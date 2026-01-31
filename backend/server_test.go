package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kc-clintone/study-coach/storage"
)

func TestCreateSessionAndSubmitFlow(t *testing.T) {
	// cleanup data dir
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	router := setupRouter()

	reqBody := CreateSessionRequest{
		StudentID:    "stu1",
		StudentLevel: "grade10",
		Subject:      "math",
		Tasks:        []Task{{ID: "t1", Prompt: "1+1"}, {ID: "t2", Prompt: "integral of x"}},
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create session failed: %d %s", rec.Code, rec.Body.String())
	}
	var sess Session
	json.Unmarshal(rec.Body.Bytes(), &sess)

	// submit results
	sub := Submission{SessionID: sess.ID, Results: []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	}{
		{ActualSeconds: 30, Correct: true},
		{ActualSeconds: 400, Correct: false},
	}}
	sb, _ := json.Marshal(sub)
	req2 := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(sb))
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("submit failed: %d %s", rec2.Code, rec2.Body.String())
	}
}
