package main

import (
	"github.com/kc-clintone/study-coach/ai"
)

// QueryGemini delegates to pkg/ai.QueryGemini.
func QueryGemini(prompt string) (string, error) {
	return ai.QueryGemini(prompt)
}
