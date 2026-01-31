package main

import "strings"

func normalizeSubject(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
