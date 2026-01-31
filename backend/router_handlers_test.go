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

func TestCreateSessionInvalidSubject(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	reqBody := CreateSessionRequest{StudentID: "s1", StudentLevel: "grade9", Subject: "cooking", Tasks: []Task{{ID: "t1", Prompt: "chop onions"}}}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-educational subject, got %d", rec.Code)
	}
}

func TestGetStudyPlanNotFound(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	req := httptest.NewRequest("GET", "/api/v1/study_plan/doesnotexist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing plan, got %d", rec.Code)
	}
}

func TestSubmitResultsSessionNotFound(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	sub := Submission{SessionID: "nope", Results: []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	}{{ActualSeconds: 10, Correct: true}}}
	b, _ := json.Marshal(sub)
	req := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing session on submit, got %d", rec.Code)
	}
}

func TestIsEducationalSubject(t *testing.T) {
	if !IsEducationalSubject("math") {
		t.Fatal("math should be educational")
	}
	if IsEducationalSubject("cooking") {
		t.Fatal("cooking should not be educational")
	}
}
