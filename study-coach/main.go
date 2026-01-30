package main

import (
	"log"
	"net/http"
)

func main() {
	router := setupRouter()

	log.Println("study-coach server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
