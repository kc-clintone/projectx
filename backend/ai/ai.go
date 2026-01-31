package ai

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
// generated text. It will only perform real external HTTP calls if GEMINI_ENABLED=1
// and GEMINI_API_KEY is configured. The request URL can be overridden in tests and
// alternate deployments by setting GEMINI_API_URL environment variable; the value
// will be used as the base URL and the API key will be appended as a query parameter.
func QueryGemini(prompt string) (string, error) {
	if os.Getenv("GEMINI_ENABLED") != "1" {
		return "", fmt.Errorf("gemini disabled: set GEMINI_ENABLED=1 to enable external calls")
	}

	key := loadGeminiKey()
	if key == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not configured; set GEMINI_API_KEY in environment or in .env.local")
	}

	// Allow overriding the base URL for testing or alternative endpoints
	baseURL := os.Getenv("GEMINI_API_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta2/models/gemini-3-flash-preview:generateText"
	}
	// append key as query parameter (handle existing query string)
	url := baseURL
	if strings.Contains(url, "?") {
		url = fmt.Sprintf("%s&key=%s", url, key)
	} else {
		url = fmt.Sprintf("%s?key=%s", url, key)
	}

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

	var parsed map[string]interface{}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return string(respBytes), nil
	}

	if v, ok := parsed["candidates"]; ok {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			first := arr[0]
			if m, ok := first.(map[string]interface{}); ok {
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

	if out, ok := parsed["output"].(string); ok && out != "" {
		return out, nil
	}
	if s := deepFindString(parsed); s != "" {
		return s, nil
	}

	raw, _ := json.Marshal(parsed)
	return string(raw), nil
}

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
