package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

//Request and Response structures for Gemini API

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

//EstimateAssignmentTime takes assignment details and calls the AI

func EstimateAssignmentTime(apiKey string, subject string, grade string, questions string) (string, error) {
	//get the prompt
	systemPrompt := GetTimeEstimationPrompt()

	//combine with the student input
	fullPrompt := fmt.Sprintf("%s\n\nSubject: %s\nGrade Level: %s\nQuestions: %s", systemPrompt, subject, grade, questions)

	//fill out the structs
	requestBody := GeminiRequest{
		Contents: []Content {
			{
				Parts: []Part{
					{Text: fullPrompt},
				},
			},
		},
	}

	//Send it to Google

	apiUrl := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

	//turn Go struct to JSON
	jsonData, _ := json.Marshal(requestBody)

	//Make actual request
	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return "AI response received!", nil
}