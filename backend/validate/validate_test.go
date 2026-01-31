package validate

import "testing"

func TestIsLikelyAcademicTask_Positive(t *testing.T) {
	cases := []string{
		"1+1",
		"integral of x",
		"Solve for x in the equation x^2 - 4 = 0",
		"Explain photosynthesis",
		"What is the derivative of x^2?",
	}
	for _, c := range cases {
		if !IsLikelyAcademicTask(c) {
			t.Fatalf("expected academic for %q", c)
		}
	}
}

func TestIsLikelyAcademicTask_Negative(t *testing.T) {
	cases := []string{
		"buy milk and eggs",
		"call mom",
		"play soccer",
		"hello",
	}
	for _, c := range cases {
		if IsLikelyAcademicTask(c) {
			t.Fatalf("expected non-academic for %q", c)
		}
	}
}
