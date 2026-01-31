package coach

import "testing"

func TestEstimateTimeForComplexityByGrade(t *testing.T) {
	// complexity 2 for different grades
	if EstimateTimeForComplexity(2, "grade3") < EstimateTimeForComplexity(2, "grade10") {
		t.Fatal("expected grade3 to have larger base time than grade10")
	}
	if EstimateTimeForComplexity(2, "grade11") >= EstimateTimeForComplexity(2, "grade3") {
		t.Fatal("expected grade11 to have smaller base time")
	}
}
