package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kc-clintone/study-coach/model"
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

// EnsureDir is an exported wrapper that guarantees the storage directory exists.
func EnsureDir() error { return ensureDir() }

// GetStorageDir returns the configured storage directory path.
func GetStorageDir() string { return storageDir }

// SaveSession persists a session to the file system
func SaveSession(s *model.Session) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(storageDir, s.ID+".session.json")
	return os.WriteFile(path, b, 0o644)
}

func LoadSession(id string) (*model.Session, error) {
	path := filepath.Join(storageDir, id+".session.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s model.Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveStudyPlan(studentID string, p *model.StudyPlan) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	path := filepath.Join(storageDir, studentID+".plan.json")
	return os.WriteFile(path, b, 0o644)
}

func LoadStudyPlan(studentID string) (*model.StudyPlan, error) {
	path := filepath.Join(storageDir, studentID+".plan.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p model.StudyPlan
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

// SaveStudentTopics persists per-student topic frequency map
func SaveStudentTopics(studentID string, topics map[string]int) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(topics, "", "  ")
	path := filepath.Join(storageDir, studentID+".topics.json")
	return os.WriteFile(path, b, 0o644)
}

// LoadStudentTopics loads per-student topic frequencies; returns empty map if not found
func LoadStudentTopics(studentID string) (map[string]int, error) {
	path := filepath.Join(storageDir, studentID+".topics.json")
	b, err := os.ReadFile(path)
	if err != nil {
		// if file missing, return empty map
		return map[string]int{}, nil
	}
	var m map[string]int
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ErrNotFound() error { return errors.New("not found") }

// SaveProfile persists a student's profile information
func SaveProfile(studentID string, profile *model.ProfileResponse) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(profile, "", "  ")
	path := filepath.Join(storageDir, studentID+".profile.json")
	return os.WriteFile(path, b, 0o644)
}

// LoadProfile loads a student's profile information; returns error if not found
func LoadProfile(studentID string) (*model.ProfileResponse, error) {
	path := filepath.Join(storageDir, studentID+".profile.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var profile model.ProfileResponse
	if err := json.Unmarshal(b, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// SaveSnapshot persists a snapshot of the student's progress
func SaveSnapshot(studentID string, snapshot *model.TopicSnapshot) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(snapshot, "", "  ")
	path := filepath.Join(storageDir, studentID+".snapshot.json")
	return os.WriteFile(path, b, 0o644)
}

// LoadSnapshot loads a student's progress snapshot; returns error if not found
func LoadSnapshot(studentID string) (*model.TopicSnapshot, error) {
	path := filepath.Join(storageDir, studentID+".snapshot.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snapshot model.TopicSnapshot
	if err := json.Unmarshal(b, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// SaveUser persists a User account as JSON under storage dir
func SaveUser(u *model.User) error {
	if err := ensureDir(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(u, "", "  ")
	path := filepath.Join(storageDir, "user_"+u.Username+".json")
	return os.WriteFile(path, b, 0o600)
}

// LoadUser loads a User account by username
func LoadUser(username string) (*model.User, error) {
	path := filepath.Join(storageDir, "user_"+username+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var u model.User
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// DeleteUser removes a user file
func DeleteUser(username string) error {
	path := filepath.Join(storageDir, "user_"+username+".json")
	return os.Remove(path)
}

// ListUsers returns all usernames
func ListUsers() ([]string, error) {
	if err := ensureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "user_") && strings.HasSuffix(name, ".json") {
			out = append(out, strings.TrimSuffix(strings.TrimPrefix(name, "user_"), ".json"))
		}
	}
	return out, nil
}
