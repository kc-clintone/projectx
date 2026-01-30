package models

type AssignmentRequest struct {
	Subject     string `json:"subject"`
	Grade       string `json:"grade"`
	Content     string `json:"content"`         // OCR text or user input
	ImageBase64 string `json:"image,omitempty"` // For multimodal analysis
}
