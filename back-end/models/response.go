package models

type StudyPlan struct {
	TotalDuration int `json:"total_duration"` // minutes
	Tasks         []struct {
		QuestionID    string `json:"question_id"`
		EstimatedTime int    `json:"estimated_time"`
		Tip           string `json:"tip"` // AI advice based on weakness
	} `json:"tasks"`
}
