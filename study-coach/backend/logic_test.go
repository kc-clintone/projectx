package main

import (
	"testing"
)

func TestEstimateTimeForComplexity(t *testing.T) {
	if EstimateTimeForComplexity(1) < 30 {
		t.Fatal("estimated time too small")
	}
	if EstimateTimeForComplexity(3) < 150 {
		t.Fatal("unexpected estimate")
	}
}

func TestInferComplexity(t *testing.T) {
	if InferComplexity("short question", "grade1") != 1 {
		t.Fatal("short should be complexity 1")
	}
	if InferComplexity("this is a medium length prompt that has more than forty characters", "grade1") < 2 {
		t.Fatal("expected >=2")
	}
}

func TestGenerateStudyPlanAdjustsTimers(t *testing.T) {
	tasks := []Task{
		{ID: "1", Prompt: "q1", EstimatedSecs: 120, ActualSeconds: 60, Complexity: 2},
		{ID: "2", Prompt: "q2", EstimatedSecs: 120, ActualSeconds: 180, Complexity: 3},
	}
	s := &Session{ID: "s1", StudentID: "st1", Subject: "math", Tasks: tasks}
	p := GenerateStudyPlan(s)
	if len(p.NextTimers) != 2 {
		t.Fatalf("expected 2 timers, got %d", len(p.NextTimers))
	}
	if p.NextTimers[0] >= tasks[0].EstimatedSecs {
		t.Fatal("fast student should get reduced time")
	}
	if p.NextTimers[1] <= tasks[1].EstimatedSecs {
		t.Fatal("slow student should get increased time")
	}
}
