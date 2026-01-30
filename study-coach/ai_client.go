package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// loadGeminiKey attempts to read GEMINI_API_KEY from environment or from a local .env.local file.
// It returns the raw key (no validation) or an empty string if none is found.
func loadGeminiKey() string {
	if k := os.Getenv("GEMINI_API_KEY"); k != "" {
		return k
	}
	f, err := os.Open(".env.local")
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "GEMINI_API_KEY=") {
			return strings.TrimPrefix(line, "GEMINI_API_KEY=")
		}
		if err == io.EOF {
			break
		}
	}
	return ""
}

// QueryGemini sends a prompt to the configured Gemini-compatible model and returns the
// generated text. For safety during tests and local development, this function will only
// perform real external HTTP calls if the environment variable GEMINI_ENABLED is set to "1".
// The function requires GEMINI_API_KEY to be set (in environment or .env.local).
func QueryGemini(prompt string) (string, error) {
	// gating flag prevents accidental network calls in tests or CI
	if os.Getenv("GEMINI_ENABLED") != "1" {
		return "", fmt.Errorf("gemini disabled: set GEMINI_ENABLED=1 to enable external calls")
	}

	key := loadGeminiKey()
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not configured; set GEMINI_API_KEY in environment or in .env.local")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta2/models/gemini-3-flash-preview:generateText?key=%s", key)

	reqBody := map[string]interface{}{
		"prompt": map[string]interface{}{
			"text": prompt,
		},
		"temperature":     0.2,
		"maxOutputTokens": 300,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 14 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini API error: status=%d body=%s", resp.StatusCode, string(respBytes))
	}

	// Attempt to parse response and extract generated text robustly
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		// if parsing fails, return raw body as fallback
		return string(respBytes), nil
	}

	// Common field: candidates -> [{"output":"..."}] or candidates -> [{"content":[{"text":"..."}]}]
	if v, ok := parsed["candidates"]; ok {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			first := arr[0]
			if m, ok := first.(map[string]interface{}); ok {
				// try common keys
				if out, ok := m["output"].(string); ok && out != "" {
					return out, nil
				}
				if disp, ok := m["display"].(string); ok && disp != "" {
					return disp, nil
				}
				if content, ok := m["content"].([]interface{}); ok && len(content) > 0 {
					if c0, ok := content[0].(map[string]interface{}); ok {
						if txt, ok := c0["text"].(string); ok && txt != "" {
							return txt, nil
						}
					}
				}
			}
		}
	}

	// Some responses use 'output' at top level or 'candidates'[0]['outputText'] etc.
	if out, ok := parsed["output"].(string); ok && out != "" {
		return out, nil
	}
	// Deep search for first string in expected keys
	if s := deepFindString(parsed); s != "" {
		return s, nil
	}

	// fallback: return full JSON as string
	raw, _ := json.Marshal(parsed)
	return string(raw), nil
}

// deepFindString searches nested maps/arrays for the first non-empty string value.
func deepFindString(v interface{}) string {
	switch t := v.(type) {
	case string:
		if t != "" {
			return t
		}
	case map[string]interface{}:
		for _, val := range t {
			if s := deepFindString(val); s != "" {
				return s
			}
		}
	case []interface{}:
		for _, item := range t {
			if s := deepFindString(item); s != "" {
				return s
			}
		}
	}
	return ""
}
