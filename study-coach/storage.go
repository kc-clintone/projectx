package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	storageDir = "./data"
	mu         sync.Mutex
)

func ensureDir() error {
	mu.Lock()
	defer mu.Unlock()
	return os.MkdirAll(storageDir, 0o755)
}

func SaveSession(s *Session) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(storageDir, s.ID+".session.json")
	return os.WriteFile(path, b, 0o644)
}

func LoadSession(id string) (*Session, error) {
	path := filepath.Join(storageDir, id+".session.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveStudyPlan(studentID string, p *StudyPlan) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	path := filepath.Join(storageDir, studentID+".plan.json")
	return os.WriteFile(path, b, 0o644)
}

func LoadStudyPlan(studentID string) (*StudyPlan, error) {
	path := filepath.Join(storageDir, studentID+".plan.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p StudyPlan
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func ListSessions() ([]string, error) {
	if err := ensureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, e.Name())
	}
	return out, nil
}

func ErrNotFound() error { return errors.New("not found") }
