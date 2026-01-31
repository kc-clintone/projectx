package validate

import "strings"

// IsLikelyAcademicTask applies simple heuristics to determine whether a free-text
// student task looks like an academic problem (used as a fallback when AI is disabled).
func IsLikelyAcademicTask(text string) bool {
	s := strings.ToLower(strings.TrimSpace(text))
	if s == "" {
		return false
	}
	// If it contains numbers or math symbols, consider it academic
	if strings.ContainsAny(s, "0123456789=+-*/^()") || strings.Contains(s, "?") {
		return true
	}
	// very short strings without numbers are unlikely academic
	if len(s) < 5 {
		return false
	}
	// common academic action verbs and phrases and noun forms
	keywords := []string{"solve", "calculate", "prove", "derive", "explain", "define", "compare", "evaluate", "what is", "find", "show that", "compute", "simplify", "integrate", "differentiate", "diagram", "describe", "why", "how", "integral", "integrals", "derivative", "derivatives", "equation", "equations", "factor", "expand", "prove that"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	// longer descriptive prompts are more likely academic
	if len(s) > 60 {
		return true
	}
	return false
}
