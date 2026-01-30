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
