package main

import (
	"os"
	"testing"

	"github.com/kc-clintone/study-coach/storage"
)

func TestAddAchievementAwardsBadgeAndPoints(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()
	p := &ProfileResponse{StudentID: "u1", Achievements: []string{}, EarnedBadges: []Badge{}, Points: 0}
	AddAchievement(p, "First Steps")
	if len(p.Achievements) != 1 {
		t.Fatalf("expected achievement, got %v", p.Achievements)
	}
	if p.Points <= 0 {
		t.Fatalf("expected points awarded, got %d", p.Points)
	}
	AddAchievement(p, "First Steps") // duplicate shouldn't add
	if len(p.Achievements) != 1 {
		t.Fatalf("duplicate achievement added")
	}
}

func TestPointsPersistedWithProfile(t *testing.T) {
	_ = os.RemoveAll(storage.GetStorageDir())
	_ = storage.EnsureDir()
	p := &ProfileResponse{StudentID: "u2", Achievements: []string{}, EarnedBadges: []Badge{}, Points: 0}
	AddAchievement(p, "Accuracy Ace")
	if err := storage.SaveProfile(p.StudentID, p); err != nil {
		t.Fatal(err)
	}
	loaded, err := storage.LoadProfile(p.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Points != p.Points {
		t.Fatalf("points mismatch %d vs %d", loaded.Points, p.Points)
	}
}
