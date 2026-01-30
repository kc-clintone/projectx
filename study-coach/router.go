package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) }).Methods("GET")
	router.HandleFunc("/api/v1/session", createSessionHandler).Methods("POST")
	router.HandleFunc("/api/v1/submit_results", submitResultsHandler).Methods("POST")
	router.HandleFunc("/api/v1/study_plan/{student_id}", getStudyPlanHandler).Methods("GET")
	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	return router
}
