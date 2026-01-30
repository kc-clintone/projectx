package main

import (
	"os"
	"reflect"
	"testing"
)

func TestExtractTopics_StopwordsAndStemming(t *testing.T) {
	text := "Solve the integrals of x^2 and x^3; integrating practice problems"
	out := ExtractTopics(text)
	if len(out) == 0 {
		t.Fatal("expected topics")
	}
	// check that stopwords like "the" and "of" are not present
	for _, w := range out {
		if w == "the" || w == "of" {
			t.Fatalf("stopword leaked: %s", w)
		}
	}
}

func TestStudentTopicPersistence(t *testing.T) {
	_ = os.RemoveAll(storageDir)
	_ = ensureDir()
	m := map[string]int{"integral": 2, "practice": 1}
	if err := SaveStudentTopics("s123", m); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadStudentTopics("s123")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, loaded) {
		t.Fatalf("mismatch: %v vs %v", m, loaded)
	}
}
