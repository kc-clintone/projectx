package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) }).Methods("GET")
	// auth
	router.HandleFunc("/api/v1/register", registerHandler).Methods("POST")
	router.HandleFunc("/api/v1/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/v1/logout", logoutHandler).Methods("POST")
	router.HandleFunc("/api/v1/me", getMeHandler).Methods("GET")

	router.HandleFunc("/api/v1/session", createSessionHandler).Methods("POST")
	router.HandleFunc("/api/v1/submit_results", submitResultsHandler).Methods("POST")
	router.HandleFunc("/api/v1/study_plan/{student_id}", getStudyPlanHandler).Methods("GET")
	router.HandleFunc("/api/v1/student/{student_id}/profile", getProfileHandler).Methods("GET")
	router.HandleFunc("/api/v1/student/{student_id}/topics", getStudentTopicsHandler).Methods("GET")
	// serve static frontend files from ../frontend (project root has frontend/)
	fs := http.FileServer(http.Dir("../frontend"))
	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", fs))
	// convenient root redirect to UI
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/index.html", http.StatusFound)
	})
	return router
}
