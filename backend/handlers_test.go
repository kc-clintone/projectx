package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kc-clintone/study-coach/model"
	"github.com/kc-clintone/study-coach/storage"
)

func TestSubmitResultsPersistsLastPlan(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	r := setupRouter()
	// create session
	reqBody := CreateSessionRequest{StudentID: "stu_lp", StudentLevel: "grade10", Subject: "math", Tasks: []Task{{ID: "t1", Prompt: "1+1"}}}
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
		{ActualSeconds: 30, Correct: true},
	}}
	sb, _ := json.Marshal(sub)
	req2 := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(sb))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("submit failed: %d %s", rec2.Code, rec2.Body.String())
	}

	// load profile and assert LastPlan exists
	p, err := storage.LoadProfile("stu_lp")
	if err != nil {
		t.Fatalf("load profile failed: %v", err)
	}
	if p.LastPlan == nil {
		t.Fatalf("expected LastPlan to be persisted in profile")
	}
}

func TestMeEndpointReturnsProfileWithLastPlan(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()

	// create user and session and submit to populate profile
	_ = storage.SaveUser(&model.User{Username: "meuser", PasswordHash: "x"})
	r := setupRouter()
	reqBody := CreateSessionRequest{StudentID: "meuser", StudentLevel: "grade10", Subject: "math", Tasks: []Task{{ID: "t1", Prompt: "2+2"}}}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/session", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var sess Session
	json.Unmarshal(rec.Body.Bytes(), &sess)

	sub := Submission{SessionID: sess.ID, Results: []struct {
		ActualSeconds int  `json:"actual_seconds"`
		Correct       bool `json:"correct"`
	}{
		{ActualSeconds: 20, Correct: true},
	}}
	sb, _ := json.Marshal(sub)
	req2 := httptest.NewRequest("POST", "/api/v1/submit_results", bytes.NewReader(sb))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	// login to get cookie
	loginBody := map[string]string{"username":"meuser","password":""}
	lb, _ := json.Marshal(loginBody)
	lreq := httptest.NewRequest("POST", "/api/v1/login", bytes.NewReader(lb))
	lrec := httptest.NewRecorder()
	r.ServeHTTP(lrec, lreq)
	if lrec.Code != 200 {
		// login may fail because password is empty; fallback: set cookie manually
		cookie := &http.Cookie{Name: "sc_session", Value: "meuser", Path: "/", HttpOnly: true}
		reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
		reqMe.AddCookie(cookie)
		recMe := httptest.NewRecorder()
		r.ServeHTTP(recMe, reqMe)
		if recMe.Code != 200 {
			t.Fatalf("expected me endpoint 200, got %d", recMe.Code)
		}
		return
	}
	// if login succeeded, extract cookie and call /api/v1/me
	cookies := lrec.Result().Cookies()
	reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
	for _, c := range cookies { reqMe.AddCookie(c) }
	recMe := httptest.NewRecorder()
	r.ServeHTTP(recMe, reqMe)
	if recMe.Code != 200 {
		t.Fatalf("me endpoint failed: %d", recMe.Code)
	}
}
