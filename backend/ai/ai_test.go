package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestQueryGeminiMock starts a small HTTP server that mimics a Gemini API response
// and verifies that QueryGemini correctly fetches the text from the mocked endpoint.
func TestQueryGeminiMock(t *testing.T) {
	// start mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// return a simple JSON with a top-level output
		json.NewEncoder(w).Encode(map[string]interface{}{"output": "mocked response"})
	}))
	defer srv.Close()

	// ensure env variables
	envKey := os.Getenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_API_KEY", "testkey")
	os.Setenv("GEMINI_ENABLED", "1")
	os.Setenv("GEMINI_API_URL", srv.URL)
	defer func() {
		os.Setenv("GEMINI_API_KEY", envKey)
		os.Unsetenv("GEMINI_ENABLED")
		os.Unsetenv("GEMINI_API_URL")
	}()

	out, err := QueryGemini("hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "mocked response" {
		t.Fatalf("unexpected response: %s", out)
	}
}
